package ports

import (
	"context"

	"investment-map/internal/domain"
)

// QuoteFetcher fetches a live quote for a single symbol from an external
// data source. Implemented by outbound adapters (e.g. Yahoo Finance);
// swapping data providers means writing a new adapter against this port,
// not touching domain, workers, or the REST API.
type QuoteFetcher interface {
	FetchQuote(ctx context.Context, symbol string) (domain.Quote, error)
}
