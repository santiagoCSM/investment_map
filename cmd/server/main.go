// Command server starts the REST API and launches the background workers
// that keep Redis populated with live quotes.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"investment-map/config"
	inboundhttp "investment-map/internal/adapters/inbound/http"
	"investment-map/internal/adapters/outbound/mongo"
	"investment-map/internal/adapters/outbound/redis"
	"investment-map/internal/adapters/outbound/yahoo"
	"investment-map/internal/app"
	"investment-map/internal/workers"
	"investment-map/pkg/logger"
)

func main() {
	log := logger.New()
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cache := redis.New(cfg.RedisAddr)
	defer cache.Close()
	if err := cache.Ping(ctx); err != nil {
		log.Error("redis unreachable, aborting startup", "err", err)
		os.Exit(1)
	}

	store, err := mongo.New(ctx, cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		log.Error("mongo unreachable, aborting startup", "err", err)
		os.Exit(1)
	}
	defer store.Close(context.Background())
	if err := store.EnsureIndexes(ctx); err != nil {
		log.Error("mongo ensure indexes failed, aborting startup", "err", err)
		os.Exit(1)
	}

	fetcher := yahoo.New()
	refresher := app.NewRefreshService(fetcher, cache, cfg.CacheTTL)
	history := app.NewHistoryService(fetcher, store)

	workerList := []*workers.Worker{
		workers.NewQuotesUSWorker(cfg.QuotesUS, refresher, log),
		workers.NewQuotesEuropeAsiaWorker(cfg.QuotesEuropeAsia, refresher, log),
		workers.NewQuotesMexicoWorker(cfg.QuotesMexico, refresher, log),
		workers.NewQuotesBrazilWorker(cfg.QuotesBrazil, refresher, log),
		workers.NewHistoryDailyWorker(cfg.HistoryDaily, history, log),
	}
	manager := workers.NewManager(log, workerList...)

	var allSymbols []string
	for _, wc := range cfg.AllWorkers() {
		allSymbols = append(allSymbols, wc.Symbols...)
	}

	managerDone := make(chan struct{})
	go func() {
		manager.Run(ctx)
		close(managerDone)
	}()

	query := app.NewQueryService(cache, allSymbols)
	router := inboundhttp.NewRouter(query)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: router}

	go func() {
		log.Info("http server listening", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server error", "err", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown error", "err", err)
	}

	select {
	case <-managerDone:
		log.Info("workers stopped cleanly")
	case <-time.After(10 * time.Second):
		log.Warn("workers did not stop in time")
	}

	log.Info("shutdown complete")
}
