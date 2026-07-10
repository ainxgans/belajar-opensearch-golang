package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"belajar-opensearch-golang/internal/export"
	"belajar-opensearch-golang/internal/httpx"
	"belajar-opensearch-golang/internal/product"
	"belajar-opensearch-golang/internal/search"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// waitDB pings db until it answers or gives up after ~60s.
func waitDB(db *sql.DB) error {
	var err error
	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return err
}

func main() {
	db, err := sql.Open("pgx", env("DATABASE_URL", "postgres://catalog:catalog@localhost:5432/catalog?sslmode=disable"))
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := waitDB(db); err != nil {
		log.Fatalf("db not ready: %v", err)
	}

	repo := product.NewRepo(db)
	osClient := search.NewClient(env("OPENSEARCH_URL", "http://localhost:9200"), env("OPENSEARCH_INDEX", "products"))
	enableGenerator := env("ENABLE_GENERATOR", "false") == "true"

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /api/products/search", handleSearch(osClient))
	mux.HandleFunc("GET /api/products/autocomplete", handleAutocomplete(osClient))
	mux.HandleFunc("GET /api/products/export.xlsx", handleExport(osClient))
	mux.HandleFunc("POST /api/products/generate", handleGenerate(repo, enableGenerator))
	mux.HandleFunc("GET /api/products/{id}", handleGet(repo))
	mux.HandleFunc("POST /api/products", handleCreate(repo))
	mux.HandleFunc("PUT /api/products/{id}", handleUpdate(repo))
	mux.HandleFunc("DELETE /api/products/{id}", handleDelete(repo))

	addr := env("API_ADDR", ":8080")
	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, cors(mux)); err != nil {
		log.Fatal(err)
	}
}

// cors allows all origins, echoing requested method/headers, and
// short-circuits preflight OPTIONS requests.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// parseParams converts query params into search.Params.
func parseParams(r *http.Request) search.Params {
	q := r.URL.Query()
	p := search.Params{
		Q:        q.Get("q"),
		Brand:    q.Get("brand"),
		Category: q.Get("category"),
		Tags:     httpx.QueryStrings(r, "tags"),
		Sort:     q.Get("sort"),
		Page:     httpx.QueryInt(r, "page", 1),
		Size:     httpx.QueryInt(r, "size", 20),
	}
	if v, ok := httpx.QueryFloat(r, "price_min", 0); ok {
		p.PriceMin = &v
	}
	if v, ok := httpx.QueryFloat(r, "price_max", 0); ok {
		p.PriceMax = &v
	}
	if v, ok := httpx.QueryFloat(r, "rating_min", 0); ok {
		p.RatingMin = &v
	}
	attrs := map[string]string{}
	for k, vals := range q {
		if len(k) > 5 && k[:5] == "attr." && len(vals) > 0 {
			attrs[k[5:]] = vals[0]
		}
	}
	if len(attrs) > 0 {
		p.Attrs = attrs
	}
	return p
}

func runSearch(ctx context.Context, c *search.Client, p search.Params) (search.SearchResult, error) {
	body, err := json.Marshal(search.BuildSearch(p))
	if err != nil {
		return search.SearchResult{}, err
	}
	raw, err := c.SearchRaw(ctx, body)
	if err != nil {
		return search.SearchResult{}, err
	}
	return search.ParseSearch(raw)
}

func handleSearch(c *search.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := runSearch(r.Context(), c, parseParams(r))
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, res)
	}
}

func handleAutocomplete(c *search.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		body, err := json.Marshal(search.BuildAutocomplete(q))
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		raw, err := c.SearchRaw(r.Context(), body)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		suggestions, err := search.ParseSuggestions(raw)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"suggestions": suggestions})
	}
}

func handleExport(c *search.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := parseParams(r)
		p.Size = 100
		p.Page = 1

		var all []product.Product
		for {
			res, err := runSearch(r.Context(), c, p)
			if err != nil {
				httpx.Error(w, http.StatusInternalServerError, err.Error())
				return
			}
			all = append(all, res.Items...)
			if len(res.Items) < p.Size || int64(len(all)) >= res.Total {
				break
			}
			p.Page++
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", `attachment; filename="products.xlsx"`)
		if err := export.WriteXLSX(w, all); err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
}

const maxGenerateCount = 10000

func handleGenerate(repo *product.Repo, enabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !enabled {
			httpx.Error(w, http.StatusForbidden, "generator disabled")
			return
		}
		var cfg product.GenerateConfig
		if err := httpx.DecodeJSON(r, &cfg); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid body")
			return
		}
		if cfg.Count > maxGenerateCount {
			httpx.Error(w, http.StatusBadRequest, "count exceeds 10000")
			return
		}
		if cfg.Truncate {
			if err := repo.Truncate(r.Context()); err != nil {
				httpx.Error(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		products := product.Generate(cfg)
		n, err := repo.BulkInsert(r.Context(), products, 500)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]int{"inserted": n})
	}
}

func handleGet(repo *product.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := httpx.IDParam(r, "id")
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		p, err := repo.Get(r.Context(), id)
		if errors.Is(err, product.ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, p)
	}
}

func validateInput(in product.Input) string {
	switch {
	case in.SKU == "":
		return "sku is required"
	case in.Name == "":
		return "name is required"
	case in.Brand == "":
		return "brand is required"
	case in.Category == "":
		return "category is required"
	default:
		return ""
	}
}

func handleCreate(repo *product.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in product.Input
		if err := httpx.DecodeJSON(r, &in); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid body")
			return
		}
		if msg := validateInput(in); msg != "" {
			httpx.Error(w, http.StatusBadRequest, msg)
			return
		}
		p, err := repo.Create(r.Context(), in)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusCreated, p)
	}
}

func handleUpdate(repo *product.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := httpx.IDParam(r, "id")
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		var in product.Input
		if err := httpx.DecodeJSON(r, &in); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid body")
			return
		}
		if msg := validateInput(in); msg != "" {
			httpx.Error(w, http.StatusBadRequest, msg)
			return
		}
		p, err := repo.Update(r.Context(), id, in)
		if errors.Is(err, product.ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, p)
	}
}

func handleDelete(repo *product.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := httpx.IDParam(r, "id")
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		err = repo.Delete(r.Context(), id)
		if errors.Is(err, product.ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
