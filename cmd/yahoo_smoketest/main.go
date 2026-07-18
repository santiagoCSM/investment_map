// Command yahoo_smoketest is a throwaway program to confirm the Yahoo
// adapter can fetch a live quote and real historical data before anything
// else (Redis, Mongo, workers) gets wired up to it. Run with:
//
//	go run ./cmd/yahoo_smoketest
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"investment-map/internal/adapters/outbound/yahoo"
	"investment-map/internal/domain"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	adapter := yahoo.New()

	quote, err := adapter.FetchQuote(ctx, "^GSPC")
	if err != nil {
		log.Fatalf("FetchQuote failed: %v", err)
	}
	fmt.Printf("QUOTE ^GSPC: price=%.2f change=%.2f (%.2f%%) currency=%s timestamp=%s\n",
		quote.Price, quote.Change, quote.ChangePercent, quote.Currency, quote.Timestamp.Format(time.RFC3339))

	history, err := adapter.FetchHistory(ctx, "^GSPC", domain.HistoryRange1Month)
	if err != nil {
		log.Fatalf("FetchHistory failed: %v", err)
	}
	fmt.Printf("HISTORY ^GSPC: %d points\n", len(history))
	for i, p := range history {
		if i >= 5 {
			fmt.Printf("  ... and %d more\n", len(history)-5)
			break
		}
		fmt.Printf("  %s close=%.2f\n", p.Date.Format("2006-01-02"), p.Close)
	}

	// Confirm ^MXX (previously validated live by the user) still works too.
	mxx, err := adapter.FetchQuote(ctx, "^MXX")
	if err != nil {
		log.Fatalf("FetchQuote ^MXX failed: %v", err)
	}
	fmt.Printf("QUOTE ^MXX: price=%.2f change=%.2f (%.2f%%) currency=%s timestamp=%s\n",
		mxx.Price, mxx.Change, mxx.ChangePercent, mxx.Currency, mxx.Timestamp.Format(time.RFC3339))
}
