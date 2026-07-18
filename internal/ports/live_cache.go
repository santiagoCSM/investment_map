package ports

import (
	"context"
	"time"

	"investment-map/internal/domain"
)

// LiveCache stores the most recent quote for each symbol (the "estado
// vivo"). Implemented by the Redis adapter; the REST API only ever reads
// through this port, never talking to Yahoo directly.
type LiveCache interface {
	SetQuote(ctx context.Context, quote domain.Quote, ttl time.Duration) error
	GetQuote(ctx context.Context, symbol string) (domain.Quote, error)
	GetQuotes(ctx context.Context, symbols []string) ([]domain.Quote, error)
}
