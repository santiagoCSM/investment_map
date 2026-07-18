// Package workers runs one independent goroutine per symbol group, each on
// its own time.Ticker, so a problem with one group (e.g. Yahoo rate-limiting
// Mexico-specific requests) never stalls or crashes the others.
package workers

import (
	"context"
	"log/slog"
	"time"
)

// symbolSpacing is the pause between individual symbol fetches within a
// single tick, so a worker never fires a burst of simultaneous requests at
// Yahoo.
const symbolSpacing = 300 * time.Millisecond

// WorkerError is sent on the manager's shared error channel whenever a
// symbol refresh fails, carrying enough context to know exactly which
// worker/symbol/when without digging through logs.
type WorkerError struct {
	Worker string
	Symbol string
	Err    error
	Time   time.Time
}

// Refresher does whatever "refresh one symbol" means for a given worker
// group. app.RefreshService (live quote → Redis) and app.HistoryService
// (daily close → Mongo) both implement it, so the same ticker/spacing/
// error-channel machinery below drives every worker group regardless of
// which store it's actually writing to.
type Refresher interface {
	RefreshSymbol(ctx context.Context, symbol string) error
}

// Worker refreshes a fixed group of symbols on its own ticker.
type Worker struct {
	Name      string
	Symbols   []string
	Interval  time.Duration
	Refresher Refresher
	Logger    *slog.Logger
}

// Run blocks until ctx is cancelled, refreshing Symbols once immediately
// and then again every Interval.
func (w *Worker) Run(ctx context.Context, errCh chan<- WorkerError) {
	log := w.Logger.With("worker", w.Name)
	log.Info("worker starting", "symbols", len(w.Symbols), "interval", w.Interval)

	w.tick(ctx, errCh, log)

	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("worker stopped")
			return
		case <-ticker.C:
			w.tick(ctx, errCh, log)
		}
	}
}

func (w *Worker) tick(ctx context.Context, errCh chan<- WorkerError, log *slog.Logger) {
	for i, symbol := range w.Symbols {
		if ctx.Err() != nil {
			return
		}
		if i > 0 {
			select {
			case <-time.After(symbolSpacing):
			case <-ctx.Done():
				return
			}
		}

		if err := w.Refresher.RefreshSymbol(ctx, symbol); err != nil {
			log.Error("refresh failed", "symbol", symbol, "err", err)
			select {
			case errCh <- WorkerError{Worker: w.Name, Symbol: symbol, Err: err, Time: time.Now()}:
			case <-ctx.Done():
				return
			}
			continue
		}
		log.Debug("refreshed", "symbol", symbol)
	}
}
