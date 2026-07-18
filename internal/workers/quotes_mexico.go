package workers

import (
	"log/slog"

	"investment-map/config"
	"investment-map/internal/app"
)

// NewQuotesMexicoWorker builds the worker refreshing Mexico's proxy assets:
// EWW, the major NYSE-listed Mexican ADRs (AMX, FMX, CX, KOF, TV), USD/MXN,
// and the IPC index (^MXX).
func NewQuotesMexicoWorker(cfg config.WorkerConfig, refresher *app.RefreshService, log *slog.Logger) *Worker {
	return &Worker{
		Name:      cfg.Name,
		Symbols:   cfg.Symbols,
		Interval:  cfg.Interval,
		Refresher: refresher,
		Logger:    log,
	}
}
