package workers

import (
	"context"
	"log/slog"
	"sync"
)

// Manager starts every worker as its own goroutine, collects their errors
// on a shared channel, and shuts them all down cleanly when ctx is
// cancelled — a failure or slowdown in one worker never blocks another.
type Manager struct {
	workers []*Worker
	errCh   chan WorkerError
	logger  *slog.Logger
}

func NewManager(logger *slog.Logger, workers ...*Worker) *Manager {
	return &Manager{
		workers: workers,
		errCh:   make(chan WorkerError, 32),
		logger:  logger,
	}
}

// Run starts all workers and blocks until ctx is cancelled and every
// worker has returned, logging each WorkerError as it arrives.
func (m *Manager) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for _, w := range m.workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			w.Run(ctx, m.errCh)
		}(w)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case werr := <-m.errCh:
			m.logger.Error("worker reported error",
				"worker", werr.Worker, "symbol", werr.Symbol, "err", werr.Err, "at", werr.Time)
		case <-done:
			return
		}
	}
}
