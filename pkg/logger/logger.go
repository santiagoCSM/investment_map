// Package logger provides the structured logger shared by workers and the
// server, so every log line can be filtered by which worker emitted it
// (e.g. worker=quotes-us) when something fails.
package logger

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
