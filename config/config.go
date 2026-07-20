// Package config assembles runtime configuration (symbols per worker,
// intervals, connection strings) from environment variables with sensible
// local-dev defaults, so cmd/server has nothing to hardcode.
package config

import (
	"os"
	"time"

	"investment-map/internal/domain"
)

type WorkerConfig struct {
	Name     string
	Symbols  []string
	Interval time.Duration
}

// Config groups are explicit fields, not a generic slice, because each
// worker group gets its own constructor (workers.NewQuotesUSWorker, etc.)
// so its cadence or symbol list can later be tuned independently — e.g.
// slowing down quotes-mexico alone if Yahoo starts rate-limiting just
// those requests, without touching the others.
type Config struct {
	RedisURL      string
	MongoURI      string
	MongoDatabase string
	HTTPAddr      string
	StaticDir     string
	CacheTTL      time.Duration

	QuotesUS         WorkerConfig
	QuotesEuropeAsia WorkerConfig
	QuotesMexico     WorkerConfig
	QuotesBrazil     WorkerConfig
	HistoryDaily     WorkerConfig
}

// AllWorkers returns the live-quote (Redis-backed) worker groups — used to
// build the full tracked symbol list for the query service. HistoryDaily is
// deliberately excluded: it writes to Mongo, not Redis, so it isn't part of
// what QueryService serves.
func (c Config) AllWorkers() []WorkerConfig {
	return []WorkerConfig{c.QuotesUS, c.QuotesEuropeAsia, c.QuotesMexico, c.QuotesBrazil}
}

func Load() Config {
	interval := getDurationEnv("REFRESH_INTERVAL", 5*time.Minute)
	historyInterval := getDurationEnv("HISTORY_REFRESH_INTERVAL", 24*time.Hour)

	return Config{
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGO_DATABASE", "investment_map"),
		HTTPAddr:      httpAddr(),
		StaticDir:     getEnv("STATIC_DIR", "web/dist"),
		CacheTTL:      10 * time.Minute,

		QuotesUS:         WorkerConfig{Name: "quotes-us", Symbols: usIndexSymbols(), Interval: interval},
		QuotesEuropeAsia: WorkerConfig{Name: "quotes-europe-asia", Symbols: europeAsiaIndexSymbols(), Interval: interval},
		QuotesMexico:     WorkerConfig{Name: "quotes-mexico", Symbols: countrySymbols("MX"), Interval: interval},
		QuotesBrazil:     WorkerConfig{Name: "quotes-brazil", Symbols: countrySymbols("BR"), Interval: interval},
		HistoryDaily:     WorkerConfig{Name: "history-daily", Symbols: allCatalogSymbols(), Interval: historyInterval},
	}
}

// usIndexSymbols pulls the US/Canada-listed index tickers out of the master
// catalog. Mexican ADRs also trade on NYSE but belong to quotes-mexico
// (grouped by country of origin, not listing exchange).
func usIndexSymbols() []string {
	var symbols []string
	for _, s := range domain.MasterCatalog {
		if s.Region == domain.RegionNorthAmerica && s.AssetClass == domain.AssetClassIndex {
			symbols = append(symbols, s.Ticker)
		}
	}
	return symbols
}

func europeAsiaIndexSymbols() []string {
	var symbols []string
	for _, s := range domain.MasterCatalog {
		if (s.Region == domain.RegionEurope || s.Region == domain.RegionAsia) && s.AssetClass == domain.AssetClassIndex {
			symbols = append(symbols, s.Ticker)
		}
	}
	return symbols
}

// countrySymbols returns every catalog entry for a given ISO country code,
// across all asset classes (index, ETF, ADR, forex) — used for the
// single-country LatAm groups (Mexico: ^MXX, EWW, ADRs, USD/MXN. Brazil:
// ^BVSP, EWZ).
func countrySymbols(countryCode string) []string {
	var symbols []string
	for _, s := range domain.MasterCatalog {
		if s.CountryCode == countryCode {
			symbols = append(symbols, s.Ticker)
		}
	}
	return symbols
}

// allCatalogSymbols returns every tracked ticker. Unlike the quotes-*
// groups, history isn't split by region — one daily job covers everything.
func allCatalogSymbols() []string {
	symbols := make([]string, len(domain.MasterCatalog))
	for i, s := range domain.MasterCatalog {
		symbols[i] = s.Ticker
	}
	return symbols
}

// httpAddr checks PORT first — Railway (and Render/Heroku) assign a port
// dynamically and expect the app to bind to it via that env var, not a
// fixed one we pick ourselves. HTTP_ADDR remains for local dev / anywhere
// else that wants to set a full "host:port" explicitly.
func httpAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return getEnv("HTTP_ADDR", ":8080")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
