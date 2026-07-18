package workers

import (
	"log/slog"

	"investment-map/config"
	"investment-map/internal/app"
)

// NewQuotesBrazilWorker builds the worker refreshing Brazil's proxy assets:
// the Ibovespa index (^BVSP) and the iShares MSCI Brazil ETF (EWZ). Split
// out as its own group, separate from Mexico, even though both are LatAm,
// so each country's cadence can be tuned independently later.
func NewQuotesBrazilWorker(cfg config.WorkerConfig, refresher *app.RefreshService, log *slog.Logger) *Worker {
	return &Worker{
		Name:      cfg.Name,
		Symbols:   cfg.Symbols,
		Interval:  cfg.Interval,
		Refresher: refresher,
		Logger:    log,
	}
}
