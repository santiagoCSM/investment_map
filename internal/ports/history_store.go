package ports

import (
	"context"

	"investment-map/internal/domain"
)

// HistoryStore persists and queries each symbol's historical daily-close
// series. Implemented by the Mongo adapter and fed by the daily backfill
// job and history_daily worker.
type HistoryStore interface {
	Append(ctx context.Context, symbol string, point domain.HistoryPoint) error
	Query(ctx context.Context, symbol string, rng domain.HistoryRange) ([]domain.HistoryPoint, error)
}
