package search

import (
	"encoding/json"

	"belajar-opensearch-golang/internal/product"
)

type SearchResult struct {
	Items       []product.Product `json:"items"`
	Total       int64             `json:"total"`
	Facets      map[string]any    `json:"facets"`
	Suggestions []string          `json:"suggestions"`
	TookMs      int               `json:"took_ms"`
}

type osResponse struct {
	Took int `json:"took"`
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source    product.Product     `json:"_source"`
			Highlight map[string][]string `json:"highlight"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]json.RawMessage `json:"aggregations"`
	Suggest      struct {
		Dym []struct {
			Options []struct {
				Text string `json:"text"`
			} `json:"options"`
		} `json:"dym"`
	} `json:"suggest"`
}

func ParseSearch(raw []byte) (SearchResult, error) {
	var r osResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return SearchResult{}, err
	}
	res := SearchResult{
		Items:       make([]product.Product, 0, len(r.Hits.Hits)),
		Total:       r.Hits.Total.Value,
		Facets:      map[string]any{},
		Suggestions: []string{},
		TookMs:      r.Took,
	}
	for _, h := range r.Hits.Hits {
		res.Items = append(res.Items, h.Source)
	}
	for k, v := range r.Aggregations {
		var m map[string]any
		_ = json.Unmarshal(v, &m)
		res.Facets[k] = m
	}
	for _, s := range r.Suggest.Dym {
		for _, o := range s.Options {
			res.Suggestions = append(res.Suggestions, o.Text)
		}
	}
	return res, nil
}

type acResponse struct {
	Suggest struct {
		Complete []struct {
			Options []struct {
				Text string `json:"text"`
			} `json:"options"`
		} `json:"complete"`
	} `json:"suggest"`
}

func ParseSuggestions(raw []byte) ([]string, error) {
	var r acResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	out := []string{}
	for _, c := range r.Suggest.Complete {
		for _, o := range c.Options {
			out = append(out, o.Text)
		}
	}
	return out, nil
}
