package product

import (
	"fmt"
	"math/rand"
)

var (
	adjectives = []string{"Pro", "Max", "Lite", "Plus", "Mini", "Ultra", "Air", "Prime"}
	nouns      = []string{"Phone", "Laptop", "Watch", "Speaker", "Camera", "Tablet", "Headset", "Monitor"}

	defaultBrands     = []string{"Acme", "Globex", "Initech", "Umbrella", "Hooli"}
	defaultCategories = []string{"Electronics", "Home", "Sports", "Toys", "Office"}
	defaultTags       = []string{"new", "sale", "featured", "limited", "eco", "bestseller"}

	attrPool = []Attribute{
		{Name: "color", Value: "black"},
		{Name: "color", Value: "white"},
		{Name: "color", Value: "red"},
		{Name: "size", Value: "S"},
		{Name: "size", Value: "M"},
		{Name: "size", Value: "L"},
		{Name: "material", Value: "plastic"},
		{Name: "material", Value: "metal"},
		{Name: "material", Value: "fabric"},
	}
)

// defaultsFor fills unset fields of cfg with sane defaults.
func defaultsFor(cfg GenerateConfig) GenerateConfig {
	if cfg.Count <= 0 {
		cfg.Count = 100
	}
	if cfg.Prefix == "" {
		cfg.Prefix = "GEN"
	}
	if cfg.Seed == 0 {
		cfg.Seed = 1
	}
	if cfg.PriceMin <= 0 {
		cfg.PriceMin = 500
	}
	if cfg.PriceMax <= 0 || cfg.PriceMax <= cfg.PriceMin {
		cfg.PriceMax = 500000
	}
	if len(cfg.Brands) == 0 {
		cfg.Brands = defaultBrands
	}
	if len(cfg.Categories) == 0 {
		cfg.Categories = defaultCategories
	}
	if len(cfg.Tags) == 0 {
		cfg.Tags = defaultTags
	}
	if cfg.AttrMax <= 0 {
		cfg.AttrMax = 3
	}
	return cfg
}

// pickSkewed returns an index in [0,n) biased toward smaller values,
// so early elements of a pool show up more often than late ones.
func pickSkewed(rng *rand.Rand, n int) int {
	if n <= 0 {
		return 0
	}
	// ponytail: square a uniform float to skew toward 0 instead of a real
	// weighted-distribution table; good enough for fake data generation.
	f := rng.Float64() * rng.Float64()
	i := int(f * float64(n))
	if i >= n {
		i = n - 1
	}
	return i
}

// Generate builds a deterministic slice of fake products from cfg.
// The same cfg.Seed (with the rest of cfg held constant) always produces
// the same output.
func Generate(cfg GenerateConfig) []Product {
	cfg = defaultsFor(cfg)
	rng := rand.New(rand.NewSource(cfg.Seed))

	products := make([]Product, 0, cfg.Count)
	for i := 0; i < cfg.Count; i++ {
		adj := adjectives[pickSkewed(rng, len(adjectives))]
		noun := nouns[pickSkewed(rng, len(nouns))]

		p := Product{
			SKU:         fmt.Sprintf("%s-%d-%d", cfg.Prefix, cfg.Seed, i),
			Name:        fmt.Sprintf("%s %s %s %d", cfg.Prefix, adj, noun, i),
			Description: fmt.Sprintf("%s %s, generated for testing.", adj, noun),
			Brand:       cfg.Brands[pickSkewed(rng, len(cfg.Brands))],
			Category:    cfg.Categories[pickSkewed(rng, len(cfg.Categories))],
			Price:       roundToHundred(cfg.PriceMin + rng.Float64()*(cfg.PriceMax-cfg.PriceMin)),
			Stock:       rng.Intn(501),
			Rating:      roundTo1Decimal(rng.Float64() * 5.0),
			Tags:        pickTags(rng, cfg.Tags),
			Attributes:  pickAttributes(rng, cfg.AttrMax),
		}
		products = append(products, p)
	}
	return products
}

// pickTags returns a random subset (0..3) of pool, no duplicates.
func pickTags(rng *rand.Rand, pool []string) []string {
	n := rng.Intn(4) // 0..3
	if n > len(pool) {
		n = len(pool)
	}
	if n == 0 {
		return nil
	}
	idx := rng.Perm(len(pool))[:n]
	tags := make([]string, n)
	for i, j := range idx {
		tags[i] = pool[j]
	}
	return tags
}

// pickAttributes returns 0..attrMax random attributes from attrPool.
func pickAttributes(rng *rand.Rand, attrMax int) []Attribute {
	n := rng.Intn(attrMax + 1)
	if n > len(attrPool) {
		n = len(attrPool)
	}
	if n == 0 {
		return nil
	}
	idx := rng.Perm(len(attrPool))[:n]
	attrs := make([]Attribute, n)
	for i, j := range idx {
		attrs[i] = attrPool[j]
	}
	return attrs
}

func roundToHundred(v float64) float64 {
	return float64(int(v/100+0.5)) * 100
}

func roundTo1Decimal(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
