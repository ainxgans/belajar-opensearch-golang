package search

type Params struct {
	Q         string
	Brand     string
	Category  string
	Tags      []string
	PriceMin  *float64
	PriceMax  *float64
	RatingMin *float64
	Attrs     map[string]string // name -> value
	Sort      string            // relevance|price_asc|price_desc|rating|newest
	Page      int
	Size      int
}

func (p Params) from() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.size()
}

func (p Params) size() int {
	if p.Size <= 0 || p.Size > 100 {
		return 20
	}
	return p.Size
}

// filters (don't affect score) used both for post_filter and facets.
func (p Params) filters() []map[string]any {
	var f []map[string]any
	if p.Brand != "" {
		f = append(f, map[string]any{"term": map[string]any{"brand": p.Brand}})
	}
	if p.Category != "" {
		// prefix for hierarchy: "Electronics" matches "Electronics/Phones"
		f = append(f, map[string]any{"prefix": map[string]any{"category": p.Category}})
	}
	if len(p.Tags) > 0 {
		f = append(f, map[string]any{"terms": map[string]any{"tags": p.Tags}})
	}
	if p.PriceMin != nil || p.PriceMax != nil {
		rng := map[string]any{}
		if p.PriceMin != nil {
			rng["gte"] = *p.PriceMin
		}
		if p.PriceMax != nil {
			rng["lte"] = *p.PriceMax
		}
		f = append(f, map[string]any{"range": map[string]any{"price": rng}})
	}
	if p.RatingMin != nil {
		f = append(f, map[string]any{"range": map[string]any{"rating": map[string]any{"gte": *p.RatingMin}}})
	}
	for name, val := range p.Attrs {
		f = append(f, map[string]any{
			"nested": map[string]any{
				"path": "attributes",
				"query": map[string]any{
					"bool": map[string]any{"filter": []map[string]any{
						{"term": map[string]any{"attributes.name": name}},
						{"term": map[string]any{"attributes.value": val}},
					}},
				},
			},
		})
	}
	return f
}

func (p Params) queryClause() map[string]any {
	if p.Q == "" {
		return map[string]any{"match_all": map[string]any{}}
	}
	inner := map[string]any{
		"bool": map[string]any{
			"should": []map[string]any{
				{"multi_match": map[string]any{
					"query":     p.Q,
					"type":      "best_fields",
					"fields":    []string{"name^3", "description", "brand.text", "category.text"},
					"fuzziness": "AUTO",
				}},
				{"match_phrase_prefix": map[string]any{"name": map[string]any{"query": p.Q, "boost": 2}}},
			},
			"minimum_should_match": 1,
		},
	}
	// function_score: boost high rating + in-stock products
	return map[string]any{
		"function_score": map[string]any{
			"query": inner,
			"functions": []map[string]any{
				{"field_value_factor": map[string]any{"field": "rating", "modifier": "log1p", "factor": 1.2, "missing": 0}},
				{"filter": map[string]any{"range": map[string]any{"stock": map[string]any{"gt": 0}}}, "weight": 1.5},
			},
			"boost_mode": "multiply",
			"score_mode": "sum",
		},
	}
}

func sortClause(sort string) []any {
	switch sort {
	case "price_asc":
		return []any{map[string]any{"price": "asc"}}
	case "price_desc":
		return []any{map[string]any{"price": "desc"}}
	case "rating":
		return []any{map[string]any{"rating": "desc"}}
	case "newest":
		return []any{map[string]any{"created_at": "desc"}}
	default:
		return []any{"_score"}
	}
}

func aggs() map[string]any {
	return map[string]any{
		"brands":     map[string]any{"terms": map[string]any{"field": "brand", "size": 20}},
		"categories": map[string]any{"terms": map[string]any{"field": "category", "size": 20}},
		"tags":       map[string]any{"terms": map[string]any{"field": "tags", "size": 20}},
		"price_ranges": map[string]any{"range": map[string]any{"field": "price", "ranges": []map[string]any{
			{"to": 100000}, {"from": 100000, "to": 1000000}, {"from": 1000000, "to": 5000000}, {"from": 5000000},
		}}},
		"rating_stats": map[string]any{"stats": map[string]any{"field": "rating"}},
		"attributes": map[string]any{
			"nested": map[string]any{"path": "attributes"},
			"aggs": map[string]any{
				"names": map[string]any{
					"terms": map[string]any{"field": "attributes.name", "size": 20},
					"aggs":  map[string]any{"values": map[string]any{"terms": map[string]any{"field": "attributes.value", "size": 20}}},
				},
			},
		},
	}
}

func BuildSearch(p Params) map[string]any {
	body := map[string]any{
		"from":  p.from(),
		"size":  p.size(),
		"query": p.queryClause(),
		"sort":  sortClause(p.Sort),
		"aggs":  aggs(),
		"highlight": map[string]any{
			"fields": map[string]any{"name": map[string]any{}, "description": map[string]any{}},
		},
	}
	// post_filter: facet counts still reflect the whole query (correct e-commerce behavior)
	if f := p.filters(); len(f) > 0 {
		body["post_filter"] = map[string]any{"bool": map[string]any{"filter": f}}
	}
	// did-you-mean suggestion
	if p.Q != "" {
		body["suggest"] = map[string]any{
			"dym": map[string]any{"text": p.Q, "term": map[string]any{"field": "name"}},
		}
	}
	return body
}

func BuildAutocomplete(q string) map[string]any {
	return map[string]any{
		"size": 0,
		"suggest": map[string]any{
			"complete": map[string]any{
				"prefix":     q,
				"completion": map[string]any{"field": "name.suggest", "size": 10, "skip_duplicates": true, "fuzzy": map[string]any{"fuzziness": 1}},
			},
		},
	}
}
