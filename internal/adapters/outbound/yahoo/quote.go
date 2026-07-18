package yahoo

import (
	"context"
	"math"
	"time"

	"investment-map/internal/domain"
)

// FetchQuote implements ports.QuoteFetcher. Market state is left as
// MarketStateUnknown here: Yahoo's own live market-state field lives behind
// the v7/finance/quote endpoint, which now requires a cookie+crumb exchange
// we don't perform, so market state is resolved separately by
// domain.ComputeMarketState using the exchange's IANA timezone.
func (a *Adapter) FetchQuote(ctx context.Context, symbol string) (domain.Quote, error) {
	result, err := a.fetchChart(ctx, symbol, "1d", "1d")
	if err != nil {
		return domain.Quote{}, err
	}

	meta := result.Meta
	change := meta.RegularMarketPrice - meta.ChartPreviousClose
	var changePercent float64
	if meta.ChartPreviousClose != 0 {
		changePercent = (change / meta.ChartPreviousClose) * 100
	}

	return domain.Quote{
		Symbol:        symbol,
		Price:         meta.RegularMarketPrice,
		Change:        round2(change),
		ChangePercent: round2(changePercent),
		Currency:      meta.Currency,
		MarketState:   domain.MarketStateUnknown,
		Timestamp:     time.Unix(meta.RegularMarketTime, 0).UTC(),
		LastUpdated:   time.Now().UTC(),
	}, nil
}

// round2 rounds to 2 decimal places, matching Yahoo's own priceHint for the
// symbols in the catalog, to avoid float64 subtraction noise (e.g.
// -76.08000000000084) leaking into the API response.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
