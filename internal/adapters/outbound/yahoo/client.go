// Package yahoo implements ports.QuoteFetcher and ports.HistoryFetcher
// against Yahoo Finance's unofficial v8/finance/chart JSON endpoint — the
// same one used internally by Python's yfinance library. No API key or
// registration is required.
//
// This is an accepted risk for the portfolio phase of the project: Yahoo
// can change or block the endpoint without notice. The hexagonal
// architecture exists so that swapping to an official provider (Finnhub,
// Twelve Data, Alpha Vantage) later means writing a new adapter against
// ports.QuoteFetcher/HistoryFetcher, not touching domain, workers, or the
// REST API.
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"investment-map/internal/ports"
	"investment-map/pkg/httpclient"
)

const defaultBaseURL = "https://query2.finance.yahoo.com"

type Adapter struct {
	http    *httpclient.Client
	baseURL string
}

var (
	_ ports.QuoteFetcher   = (*Adapter)(nil)
	_ ports.HistoryFetcher = (*Adapter)(nil)
)

func New() *Adapter {
	return &Adapter{
		http:    httpclient.New(),
		baseURL: defaultBaseURL,
	}
}

func (a *Adapter) fetchChart(ctx context.Context, symbol, interval, rng string) (chartResult, error) {
	u := fmt.Sprintf("%s/v8/finance/chart/%s?interval=%s&range=%s",
		a.baseURL, url.PathEscape(symbol), interval, rng)

	body, err := a.http.Get(ctx, u)
	if err != nil {
		return chartResult{}, fmt.Errorf("yahoo: fetch %s: %w", symbol, err)
	}

	var parsed chartResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return chartResult{}, fmt.Errorf("yahoo: decode %s: %w", symbol, err)
	}
	if parsed.Chart.Error != nil {
		return chartResult{}, fmt.Errorf("yahoo: %s returned error: %s: %s",
			symbol, parsed.Chart.Error.Code, parsed.Chart.Error.Description)
	}
	if len(parsed.Chart.Result) == 0 {
		return chartResult{}, fmt.Errorf("yahoo: %s returned empty result", symbol)
	}
	return parsed.Chart.Result[0], nil
}
