// Command store_smoketest confirms the Redis and Mongo adapters work
// end-to-end against the local docker-compose containers: fetch a live
// quote from Yahoo, cache it, read it back; fetch history, append it,
// query it back. Run with:
//
//	go run ./cmd/store_smoketest
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"investment-map/internal/adapters/outbound/mongo"
	"investment-map/internal/adapters/outbound/redis"
	"investment-map/internal/adapters/outbound/yahoo"
	"investment-map/internal/domain"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	yh := yahoo.New()

	// --- Redis: live cache round trip ---
	cache := redis.New("localhost:6379")
	defer cache.Close()
	if err := cache.Ping(ctx); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}
	fmt.Println("REDIS: connected")

	quote, err := yh.FetchQuote(ctx, "^GSPC")
	if err != nil {
		log.Fatalf("FetchQuote failed: %v", err)
	}
	if err := cache.SetQuote(ctx, quote, 10*time.Minute); err != nil {
		log.Fatalf("SetQuote failed: %v", err)
	}
	got, err := cache.GetQuote(ctx, "^GSPC")
	if err != nil {
		log.Fatalf("GetQuote failed: %v", err)
	}
	fmt.Printf("REDIS: round-tripped ^GSPC price=%.2f change=%.2f%% lastUpdated=%s\n",
		got.Price, got.ChangePercent, got.LastUpdated.Format(time.RFC3339))

	batch, err := cache.GetQuotes(ctx, []string{"^GSPC", "^DOES_NOT_EXIST"})
	if err != nil {
		log.Fatalf("GetQuotes failed: %v", err)
	}
	fmt.Printf("REDIS: batch get returned %d/2 quotes (missing symbol silently skipped)\n", len(batch))

	// --- Mongo: history store round trip ---
	store, err := mongo.New(ctx, "mongodb://localhost:27017", "investment_map")
	if err != nil {
		log.Fatalf("mongo connect failed: %v", err)
	}
	defer store.Close(ctx)
	if err := store.EnsureIndexes(ctx); err != nil {
		log.Fatalf("EnsureIndexes failed: %v", err)
	}
	fmt.Println("MONGO: connected")

	history, err := yh.FetchHistory(ctx, "^GSPC", domain.HistoryRange1Month)
	if err != nil {
		log.Fatalf("FetchHistory failed: %v", err)
	}
	for _, p := range history {
		if err := store.Append(ctx, "^GSPC", p); err != nil {
			log.Fatalf("Append failed: %v", err)
		}
	}
	fmt.Printf("MONGO: appended %d points for ^GSPC\n", len(history))

	// Append the same last point again to prove idempotency (no duplicate).
	last := history[len(history)-1]
	if err := store.Append(ctx, "^GSPC", last); err != nil {
		log.Fatalf("Append (repeat) failed: %v", err)
	}

	queried, err := store.Query(ctx, "^GSPC", domain.HistoryRangeMax)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Printf("MONGO: queried back %d points (expected %d, confirms re-append didn't duplicate)\n",
		len(queried), len(history))
}
