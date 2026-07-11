package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"belajar-opensearch-golang/internal/product"
	"belajar-opensearch-golang/internal/search"

	"github.com/segmentio/kafka-go"
)

// Debezium envelope (schemas.enable=false, decimal.handling.mode=double).
type dbzEnvelope struct {
	Op     string  `json:"op"` // c,u,d,r
	After  *dbzRow `json:"after"`
	Before *dbzRow `json:"before"`
}

type dbzRow struct {
	ID          int64    `json:"id"`
	SKU         string   `json:"sku"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Brand       string   `json:"brand"`
	Category    string   `json:"category"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	Rating      float64  `json:"rating"`
	Tags        []string `json:"tags"`
	Attributes  string   `json:"attributes"` // JSONB arrives as a string
	CreatedAt   string   `json:"created_at"` // ISO8601 string from Debezium
	UpdatedAt   string   `json:"updated_at"` // ISO8601 string from Debezium
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func toDoc(r *dbzRow) map[string]any {
	var attrs []product.Attribute
	if r.Attributes != "" {
		_ = json.Unmarshal([]byte(r.Attributes), &attrs)
	}
	if attrs == nil {
		attrs = []product.Attribute{}
	}
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}
	return map[string]any{
		"id": r.ID, "sku": r.SKU, "name": r.Name, "description": r.Description,
		"brand": r.Brand, "category": r.Category, "price": r.Price, "stock": r.Stock,
		"rating": r.Rating, "tags": tags, "attributes": attrs,
		"created_at": r.CreatedAt, "updated_at": r.UpdatedAt,
	}
}

func main() {
	osClient := search.NewClient(env("OPENSEARCH_URL", "http://opensearch:9200"), env("OPENSEARCH_INDEX", "products"))
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{env("KAFKA_BROKERS", "kafka:9092")},
		Topic:   env("KAFKA_TOPIC", "catalog.public.products"),
		GroupID: env("KAFKA_GROUP", "indexer"),
	})
	defer reader.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Println("indexer started")
	for {
		// FetchMessage does NOT commit; we commit only after a successful
		// apply so a failed write is redelivered -> real at-least-once.
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("shutting down")
				return
			}
			log.Printf("read error: %v", err)
			continue
		}
		if len(m.Value) > 0 { // len==0 is a tombstone; nothing to apply
			var envelope dbzEnvelope
			if err := json.Unmarshal(m.Value, &envelope); err != nil {
				// unparseable payload will never succeed; commit and skip it
				// so it doesn't block the partition forever.
				log.Printf("bad message, skipping: %v", err)
			} else if err := apply(ctx, osClient, envelope); err != nil {
				log.Printf("apply error (op=%s): %v", envelope.Op, err)
				// offset NOT committed -> message is redelivered next loop.
				// ponytail: tight retry loop; add backoff/DLQ if a poison
				// message ever spins here.
				continue
			}
		}
		if err := reader.CommitMessages(ctx, m); err != nil {
			log.Printf("commit error: %v", err)
		}
	}
}

func apply(ctx context.Context, os *search.Client, e dbzEnvelope) error {
	switch e.Op {
	case "c", "u", "r": // create, update, snapshot read
		if e.After == nil {
			return nil
		}
		return os.IndexDoc(ctx, e.After.ID, toDoc(e.After))
	case "d":
		if e.Before == nil {
			return nil
		}
		return os.DeleteDoc(ctx, e.Before.ID)
	default:
		return nil
	}
}
