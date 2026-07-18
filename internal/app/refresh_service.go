// Package app holds the use cases that orchestrate domain + ports without
// knowing which concrete adapter sits behind them.
package app

import (
	"context"
	"time"

	"investment-map/internal/domain"
	"investment-map/internal/ports"
)

// RefreshService fetches a live quote for one symbol and writes it into the
// live cache. It never overwrites the cache on a fetch error — the caller
// (a worker) just logs/reports the failure and the previously cached quote
// keeps serving until its TTL expires.
type RefreshService struct {
	fetcher ports.QuoteFetcher
	cache   ports.LiveCache
	ttl     time.Duration
}

func NewRefreshService(fetcher ports.QuoteFetcher, cache ports.LiveCache, ttl time.Duration) *RefreshService {
	return &RefreshService{fetcher: fetcher, cache: cache, ttl: ttl}
}

// RefreshSymbol fetches symbol's live quote and overwrites its MarketState
// with a value computed from the catalog's own exchange calendar. The
// fetcher (Yahoo) reports MarketStateUnknown for every quote — it has no
// authenticated way to ask Yahoo for real market state (see the yahoo
// package's FetchQuote doc comment) — so this is the one place that state
// actually gets resolved, using time.Now() rather than the quote's own
// (possibly stale) timestamp.
func (s *RefreshService) RefreshSymbol(ctx context.Context, symbol string) error {
	quote, err := s.fetcher.FetchQuote(ctx, symbol)
	if err != nil {
		return err
	}
	quote.MarketState = resolveMarketState(symbol, time.Now())
	return s.cache.SetQuote(ctx, quote, s.ttl)
}

func resolveMarketState(symbol string, at time.Time) domain.MarketState {
	catalogSymbol, ok := domain.FindSymbol(symbol)
	if !ok || catalogSymbol.ExchangeCode == "" {
		// Not in the catalog, or a forex pair with no exchange calendar
		// (e.g. USDMXN=X) — forex trades nearly continuously and modeling
		// its hours is out of scope for the MVP.
		return domain.MarketStateUnknown
	}
	exchange, ok := domain.ExchangeCatalog[catalogSymbol.ExchangeCode]
	if !ok {
		return domain.MarketStateUnknown
	}
	return domain.ComputeMarketState(exchange, at)
}
