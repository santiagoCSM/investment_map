package workers

import (
	"log/slog"

	"investment-map/config"
	"investment-map/internal/app"
)

// NewHistoryDailyWorker builds the once-a-day job that keeps Mongo's
// history collection current for every catalog symbol. Unlike the
// quotes-* groups it isn't split by region: history isn't time-sensitive
// enough to need per-region cadence tuning, so one worker covers the whole
// catalog.
func NewHistoryDailyWorker(cfg config.WorkerConfig, history *app.HistoryService, log *slog.Logger) *Worker {
	return &Worker{
		Name:      cfg.Name,
		Symbols:   cfg.Symbols,
		Interval:  cfg.Interval,
		Refresher: history,
		Logger:    log,
	}
}
