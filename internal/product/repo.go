package product

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

// ErrNotFound is returned when a lookup by id matches no row.
var ErrNotFound = errors.New("product: not found")

const selectCols = "id,sku,name,description,brand,category,price,stock,rating,tags,attributes,created_at,updated_at"

// Repo is the PostgreSQL-backed product repository.
type Repo struct {
	db *sql.DB
}

// NewRepo builds a Repo around an existing *sql.DB.
func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// parsePGTextArray parses a Postgres text[] literal like `{a,"b,c","d ""e"""}`
// into a []string. The pgx stdlib driver (database/sql) hands array columns
// back as their raw text form rather than decoding them into a Go slice or a
// pgtype.Scanner (that decoding only happens via pgx's native, non-sql.DB
// interface), so it has to be parsed here instead of scanned directly.
//
// Postgres quotes any element containing a comma, brace, quote, backslash or
// whitespace; inside quotes, " and \ are backslash-escaped. A naive
// strings.Split(",") corrupts such elements, so we walk the literal by hand.
func parsePGTextArray(s string) []string {
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		return nil
	}
	var out []string
	var cur strings.Builder
	inQuotes := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s): // escaped char, take next verbatim
			i++
			cur.WriteByte(s[i])
		case c == '"':
			inQuotes = !inQuotes
		case c == ',' && !inQuotes:
			out = append(out, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	out = append(out, cur.String())
	return out
}

func scanProduct(row interface{ Scan(...any) error }) (Product, error) {
	var p Product
	var tagsRaw string
	var attrsRaw []byte
	err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.Brand, &p.Category,
		&p.Price, &p.Stock, &p.Rating, &tagsRaw, &attrsRaw, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return Product{}, err
	}
	p.Tags = parsePGTextArray(tagsRaw)
	if len(attrsRaw) > 0 {
		if err := json.Unmarshal(attrsRaw, &p.Attributes); err != nil {
			return Product{}, err
		}
	}
	return p, nil
}

func attrsJSON(attrs []Attribute) ([]byte, error) {
	if attrs == nil {
		attrs = []Attribute{}
	}
	return json.Marshal(attrs)
}

// Create inserts a new product and returns the stored row.
func (r *Repo) Create(ctx context.Context, in Input) (Product, error) {
	attrs, err := attrsJSON(in.Attributes)
	if err != nil {
		return Product{}, err
	}
	tags := in.Tags
	if tags == nil {
		tags = []string{}
	}
	const q = `INSERT INTO products (sku,name,description,brand,category,price,stock,rating,tags,attributes)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING ` + selectCols
	row := r.db.QueryRowContext(ctx, q, in.SKU, in.Name, in.Description, in.Brand, in.Category,
		in.Price, in.Stock, in.Rating, tags, attrs)
	return scanProduct(row)
}

// Update overwrites all writable fields of the product with the given id.
func (r *Repo) Update(ctx context.Context, id int64, in Input) (Product, error) {
	attrs, err := attrsJSON(in.Attributes)
	if err != nil {
		return Product{}, err
	}
	tags := in.Tags
	if tags == nil {
		tags = []string{}
	}
	const q = `UPDATE products SET
sku=$1, name=$2, description=$3, brand=$4, category=$5,
price=$6, stock=$7, rating=$8, tags=$9, attributes=$10, updated_at=now()
WHERE id=$11
RETURNING ` + selectCols
	row := r.db.QueryRowContext(ctx, q, in.SKU, in.Name, in.Description, in.Brand, in.Category,
		in.Price, in.Stock, in.Rating, tags, attrs, id)
	p, err := scanProduct(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	return p, err
}

// Delete removes the product with the given id.
func (r *Repo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Get fetches a single product by id.
func (r *Repo) Get(ctx context.Context, id int64) (Product, error) {
	const q = `SELECT ` + selectCols + ` FROM products WHERE id=$1`
	row := r.db.QueryRowContext(ctx, q, id)
	p, err := scanProduct(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	return p, err
}

// BulkInsert writes products in batches, each batch committed as its own
// transaction. It is idempotent: rows are keyed by sku via
// ON CONFLICT (sku) DO UPDATE, so re-running the same slice is a no-op after
// the first successful run. Returns the number of rows written.
func (r *Repo) BulkInsert(ctx context.Context, products []Product, batch int) (int, error) {
	if batch <= 0 {
		batch = len(products)
	}
	// ponytail: id is GENERATED ALWAYS AS IDENTITY, so Postgres rejects an
	// explicit value for it ("cannot insert a non-DEFAULT value into column
	// id"). Generated products also don't carry a real id yet, so key the
	// upsert on sku (unique, deterministic per seed+prefix+index) instead.
	const q = `INSERT INTO products (sku,name,description,brand,category,price,stock,rating,tags,attributes,created_at,updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (sku) DO UPDATE SET
name=EXCLUDED.name, description=EXCLUDED.description,
brand=EXCLUDED.brand, category=EXCLUDED.category, price=EXCLUDED.price,
stock=EXCLUDED.stock, rating=EXCLUDED.rating, tags=EXCLUDED.tags,
attributes=EXCLUDED.attributes, updated_at=EXCLUDED.updated_at`

	written := 0
	for start := 0; start < len(products); start += batch {
		end := min(start+batch, len(products))
		if err := r.insertBatch(ctx, q, products[start:end]); err != nil {
			return written, err
		}
		written += end - start
	}
	return written, nil
}

func (r *Repo) insertBatch(ctx context.Context, q string, batch []Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, q)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range batch {
		attrs, err := attrsJSON(p.Attributes)
		if err != nil {
			return err
		}
		tags := p.Tags
		if tags == nil {
			tags = []string{}
		}
		if _, err := stmt.ExecContext(ctx, p.SKU, p.Name, p.Description, p.Brand, p.Category,
			p.Price, p.Stock, p.Rating, tags, attrs, p.CreatedAt, p.UpdatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Truncate removes all rows from the products table.
func (r *Repo) Truncate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `TRUNCATE products RESTART IDENTITY`)
	return err
}
