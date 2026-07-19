package yahoo

import (
	"context"
	"fmt"
	"time"

	"investment-map/internal/domain"
)

// FetchHistory implements ports.HistoryFetcher, used for the initial Mongo
// backfill and the 30-90 day scrubber / 1-year trend layer.
func (a *Adapter) FetchHistory(ctx context.Context, symbol string, rng domain.HistoryRange) ([]domain.HistoryPoint, error) {
	yahooRange, err := toYahooRange(rng)
	if err != nil {
		return nil, err
	}

	result, err := a.fetchChart(ctx, symbol, "1d", yahooRange)
	if err != nil {
		return nil, err
	}

	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("yahoo: %s history returned no quote series", symbol)
	}
	closes := result.Indicators.Quote[0].Close

	points := make([]domain.HistoryPoint, 0, len(result.Timestamp))
	for i, ts := range result.Timestamp {
		if i >= len(closes) || closes[i] == nil {
			continue // non-trading day inside the range
		}
		points = append(points, domain.HistoryPoint{
			Date:  time.Unix(ts, 0).UTC(),
			Close: *closes[i],
		})
	}
	return points, nil
}

func toYahooRange(rng domain.HistoryRange) (string, error) {
	switch rng {
	case domain.HistoryRange1Month:
		return "1mo", nil
	case domain.HistoryRange3Month:
		return "3mo", nil
	case domain.HistoryRange1Year:
		return "1y", nil
	case domain.HistoryRangeMax:
		return "max", nil
	case domain.HistoryRange10Year:
		return "10y", nil
	default:
		return "", fmt.Errorf("yahoo: unsupported history range %q", rng)
	}
}
