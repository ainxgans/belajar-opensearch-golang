package product

import (
	"fmt"
	"math/rand"
	"time"
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
	now := time.Now()

	products := make([]Product, 0, cfg.Count)
	for i := 0; i < cfg.Count; i++ {
		products = append(products, generateOne(cfg, rng, now, i))
	}
	return products
}

// GenerateStream produces cfg.Count products in slices of at most chunk,
// invoking fn for each slice. It never holds more than chunk products in
// memory, so it scales to large counts (e.g. 500k) where Generate's full
// slice would be wasteful. The slice passed to fn is reused after fn
// returns, so fn must not retain it.
func GenerateStream(cfg GenerateConfig, chunk int, fn func([]Product) error) error {
	cfg = defaultsFor(cfg)
	if chunk <= 0 {
		chunk = 1000
	}
	rng := rand.New(rand.NewSource(cfg.Seed))
	now := time.Now()

	buf := make([]Product, 0, chunk)
	for i := 0; i < cfg.Count; i++ {
		buf = append(buf, generateOne(cfg, rng, now, i))
		if len(buf) == chunk {
			if err := fn(buf); err != nil {
				return err
			}
			buf = buf[:0]
		}
	}
	if len(buf) > 0 {
		return fn(buf)
	}
	return nil
}

func generateOne(cfg GenerateConfig, rng *rand.Rand, now time.Time, i int) Product {
	adj := adjectives[pickSkewed(rng, len(adjectives))]
	noun := nouns[pickSkewed(rng, len(nouns))]
	return Product{
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
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// BulkGenerateCount is the fixed size of a zero-input bulk seed.
const BulkGenerateCount = 500_000

// bulk* are larger pools so a 500k seed has realistic facet spread instead
// of only a handful of distinct brands/categories.
var (
	bulkBrands = []string{
		"Acme", "Globex", "Initech", "Umbrella", "Hooli", "Stark", "Wayne",
		"Wonka", "Cyberdyne", "Soylent", "Tyrell", "Aperture", "BlackMesa",
		"Massive", "Vandelay", "Gekko", "Prestige", "Oscorp", "Nakatomi", "Pied",
	}
	bulkCategories = []string{
		"Electronics", "Home", "Sports", "Toys", "Office", "Garden", "Beauty",
		"Automotive", "Grocery", "Books", "Pet", "Health",
	}
)

// BulkPreset returns a zero-input config for seeding BulkGenerateCount
// products. seed varies the data per run (dynamic), so no manual input is
// needed; SKUs embed the seed so runs don't collide.
func BulkPreset(seed int64) GenerateConfig {
	return GenerateConfig{
		Count:      BulkGenerateCount,
		Prefix:     "BULK",
		Seed:       seed,
		PriceMin:   500,
		PriceMax:   5_000_000,
		Brands:     bulkBrands,
		Categories: bulkCategories,
		Tags:       defaultTags,
		AttrMax:    4,
	}
}

// pickTags returns a random subset (0..3) of pool, no duplicates.
func pickTags(rng *rand.Rand, pool []string) []string {
	n := min(rng.Intn(4), len(pool)) // 0..3
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
	n := min(rng.Intn(attrMax+1), len(attrPool))
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
