package product

import "time"

// Attribute is a free-form key/value pair attached to a product
// (e.g. color=red, size=M).
type Attribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Product is a row in the products table.
type Product struct {
	ID          int64       `json:"id"`
	SKU         string      `json:"sku"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Brand       string      `json:"brand"`
	Category    string      `json:"category"`
	Price       float64     `json:"price"`
	Stock       int         `json:"stock"`
	Rating      float64     `json:"rating"`
	Tags        []string    `json:"tags"`
	Attributes  []Attribute `json:"attributes"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Input is the writable subset of Product used for create/update.
type Input struct {
	SKU         string      `json:"sku"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Brand       string      `json:"brand"`
	Category    string      `json:"category"`
	Price       float64     `json:"price"`
	Stock       int         `json:"stock"`
	Rating      float64     `json:"rating"`
	Tags        []string    `json:"tags"`
	Attributes  []Attribute `json:"attributes"`
}

// GenerateConfig controls deterministic product generation.
type GenerateConfig struct {
	Count      int      `json:"count"`
	Prefix     string   `json:"prefix"`
	Seed       int64    `json:"seed"`
	PriceMin   float64  `json:"price_min"`
	PriceMax   float64  `json:"price_max"`
	Brands     []string `json:"brands"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
	AttrMax    int      `json:"attr_max"`
	Truncate   bool     `json:"truncate"`
}
