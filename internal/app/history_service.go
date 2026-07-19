package app

import (
	"context"
	"fmt"

	"investment-map/internal/domain"
	"investment-map/internal/ports"
)

// HistoryService keeps each symbol's Mongo history series up to date.
// RefreshSymbol implements workers.Refresher, so the same worker/ticker
// machinery that drives live-quote refreshes also drives this daily job.
type HistoryService struct {
	fetcher ports.HistoryFetcher
	store   ports.HistoryStore
}

func NewHistoryService(fetcher ports.HistoryFetcher, store ports.HistoryStore) *HistoryService {
	return &HistoryService{fetcher: fetcher, store: store}
}

// RefreshSymbol brings symbol's history up to date. The first time it ever
// runs for a symbol (no points in Mongo yet) it performs the one-time full
// backfill from Yahoo; every run after that just appends the most recent
// daily close, which is all a once-a-day job needs. This makes the backfill
// self-triggering instead of requiring a separate manual step.
func (s *HistoryService) RefreshSymbol(ctx context.Context, symbol string) error {
	existing, err := s.store.Query(ctx, symbol, domain.HistoryRangeMax)
	if err != nil {
		return fmt.Errorf("history: check existing points for %s: %w", symbol, err)
	}

	if len(existing) == 0 {
		return s.backfill(ctx, symbol)
	}
	return s.appendLatestClose(ctx, symbol)
}

func (s *HistoryService) backfill(ctx context.Context, symbol string) error {
	// Deliberately HistoryRange10Year, not HistoryRangeMax: Yahoo silently
	// coarsens interval=1d data to quarterly bars once range=max spans
	// decades, which produced unusable sparse history when this fetched
	// "max" directly (confirmed live: 168 points back to 1985, mostly
	// 3-month gaps). 10y is confirmed to return true daily granularity and
	// comfortably covers every range this API exposes (up to 1y).
	points, err := s.fetcher.FetchHistory(ctx, symbol, domain.HistoryRange10Year)
	if err != nil {
		return fmt.Errorf("history: backfill %s: %w", symbol, err)
	}
	for _, p := range points {
		if err := s.store.Append(ctx, symbol, p); err != nil {
			return fmt.Errorf("history: backfill %s: append %s: %w", symbol, p.Date.Format("2006-01-02"), err)
		}
	}
	return nil
}

func (s *HistoryService) appendLatestClose(ctx context.Context, symbol string) error {
	points, err := s.fetcher.FetchHistory(ctx, symbol, domain.HistoryRange1Month)
	if err != nil {
		return fmt.Errorf("history: refresh daily close %s: %w", symbol, err)
	}
	if len(points) == 0 {
		return fmt.Errorf("history: refresh daily close %s: no points returned", symbol)
	}
	return s.store.Append(ctx, symbol, points[len(points)-1])
}
