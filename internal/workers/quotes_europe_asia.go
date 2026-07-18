package workers

import (
	"log/slog"

	"investment-map/config"
	"investment-map/internal/app"
)

// NewQuotesEuropeAsiaWorker builds the worker refreshing major European and
// Asian indices (FTSE, DAX, CAC 40, IBEX, FTSE MIB, Nikkei, Hang Seng,
// Shanghai, Sensex, ASX, KOSPI).
func NewQuotesEuropeAsiaWorker(cfg config.WorkerConfig, refresher *app.RefreshService, log *slog.Logger) *Worker {
	return &Worker{
		Name:      cfg.Name,
		Symbols:   cfg.Symbols,
		Interval:  cfg.Interval,
		Refresher: refresher,
		Logger:    log,
	}
}
