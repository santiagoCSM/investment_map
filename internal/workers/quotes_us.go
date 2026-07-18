package workers

import (
	"log/slog"

	"investment-map/config"
	"investment-map/internal/app"
)

// NewQuotesUSWorker builds the worker refreshing major US indices
// (S&P 500, Nasdaq, Dow, TSX). Kept as its own constructor, rather than
// calling &Worker{...} inline, so this group's cadence or symbol list can
// be tuned independently once quotes_europe_asia/quotes_mexico exist.
func NewQuotesUSWorker(cfg config.WorkerConfig, refresher *app.RefreshService, log *slog.Logger) *Worker {
	return &Worker{
		Name:      cfg.Name,
		Symbols:   cfg.Symbols,
		Interval:  cfg.Interval,
		Refresher: refresher,
		Logger:    log,
	}
}
