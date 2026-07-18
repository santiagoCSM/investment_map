package ports

import (
	"context"

	"investment-map/internal/domain"
)

// HistoryFetcher fetches a historical daily-close series for a symbol from
// an external data source (used for the initial Mongo backfill and, later,
// any on-demand history requests).
type HistoryFetcher interface {
	FetchHistory(ctx context.Context, symbol string, rng domain.HistoryRange) ([]domain.HistoryPoint, error)
}
