package app

import (
	"context"
	"fmt"

	"investment-map/internal/domain"
	"investment-map/internal/ports"
)

// QueryService serves the current map state to the REST API by reading
// exclusively from the live cache and the history store — it never talks
// to Yahoo, so any number of requests never translate into extra upstream
// calls.
type QueryService struct {
	cache   ports.LiveCache
	history ports.HistoryStore
	symbols []string
}

func NewQueryService(cache ports.LiveCache, history ports.HistoryStore, symbols []string) *QueryService {
	return &QueryService{cache: cache, history: history, symbols: symbols}
}

// ListQuotes returns whatever is currently cached for the tracked catalog.
// Symbols not yet refreshed (or expired) are simply absent from the result.
func (s *QueryService) ListQuotes(ctx context.Context) ([]domain.Quote, error) {
	return s.cache.GetQuotes(ctx, s.symbols)
}

func (s *QueryService) GetQuote(ctx context.Context, symbol string) (domain.Quote, error) {
	if !s.isTracked(symbol) {
		return domain.Quote{}, fmt.Errorf("symbol %q is not tracked", symbol)
	}
	return s.cache.GetQuote(ctx, symbol)
}

// GetHistory returns symbol's daily-close series for the scrubber/trend
// layer. Always returns a non-nil slice (empty rather than null in the
// JSON response) even when the symbol hasn't been backfilled yet.
func (s *QueryService) GetHistory(ctx context.Context, symbol string, rng domain.HistoryRange) ([]domain.HistoryPoint, error) {
	if !s.isTracked(symbol) {
		return nil, fmt.Errorf("symbol %q is not tracked", symbol)
	}
	points, err := s.history.Query(ctx, symbol, rng)
	if err != nil {
		return nil, err
	}
	if points == nil {
		points = []domain.HistoryPoint{}
	}
	return points, nil
}

func (s *QueryService) isTracked(symbol string) bool {
	for _, s := range s.symbols {
		if s == symbol {
			return true
		}
	}
	return false
}
