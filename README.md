# Mapa de Inversiones

An interactive world map showing the live status of major global financial
markets — stock indices, currency pairs, and country ETFs — with layers for
long-term trend and an estimated capital-flow proxy.

This is a portfolio project built to demonstrate full-stack engineering
practice: a Go backend with a hexagonal (ports & adapters) architecture and
independently-scheduled background workers, and a React + D3 frontend
consuming its REST API. It is not intended to be monetized or operated as a
product — the priority throughout has been clean, well-architected,
testable code over feature breadth.

## Table of contents

- [Goals and scope](#goals-and-scope)
- [Tech stack](#tech-stack)
- [Architecture](#architecture)
- [Data source and caching strategy](#data-source-and-caching-strategy)
- [Symbol catalog](#symbol-catalog)
- [Market hours and timezone handling](#market-hours-and-timezone-handling)
- [Background workers](#background-workers)
- [REST API](#rest-api)
- [Testing](#testing)
- [Getting started](#getting-started)
- [Project status](#project-status)
- [Known limitations and accepted risks](#known-limitations-and-accepted-risks)

## Goals and scope

The MVP tracks roughly 25 instruments across the world's major markets,
refreshed on a 5-minute cadence (this is explicitly **not** a real-time /
tick-by-tick product — polling a REST API every few minutes is sufficient
for the goal of showing market *status*, not intraday trading data).

In scope for the MVP, in priority order:

1. A world map showing the live status of ~15-20 major stock indices, plus
   per-exchange open/closed/pre-market/post-market state computed from real
   trading calendars and IANA timezones (not hardcoded UTC offsets).
2. A basic historical scrubber: daily percentage change for each index over
   the last 30-90 days.
3. A long-term trend layer: 1-year performance, to visualize sustained
   growth vs. stagnation rather than just today's move.
4. A capital-flow proxy (lowest priority, may not ship in the MVP): relative
   performance across markets, country-ETF movement, and currency movement,
   combined into a rough signal. This is a self-constructed heuristic, not
   real institutional capital-flow data.
5. A detail panel on symbol click: price, volume, and basic fundamentals
   (P/E, market cap) where Yahoo exposes them easily — nice-to-have, not
   critical.

Explicitly out of scope for the MVP: exhaustive coverage of the Mexican
stock exchange (BMV/BIVA) beyond a handful of proxy assets, and any
per-exchange holiday calendar (a holiday currently just reads as "closed"
because it falls outside the regular session window, not because it's
specifically recognized as a holiday).

Mexico gets slightly more attention than other single countries as a
deliberate regional focus: the IPC index, the `EWW` country ETF, USD/MXN,
and ADRs of the five largest Mexican companies trading on the NYSE (América
Móvil, FEMSA, Cemex, Coca-Cola FEMSA, Grupo Televisa).

## Tech stack

**Backend** — Go 1.26, hexagonal architecture, no web framework (standard
library `net/http` with Go 1.22+'s method/path-param routing).

- [`github.com/redis/go-redis/v9`](https://github.com/redis/go-redis) — live quote cache
- [`go.mongodb.org/mongo-driver/v2`](https://github.com/mongodb/mongo-go-driver) — historical data store
- No HTTP client library — `net/http` directly against Yahoo Finance's
  unofficial JSON endpoints (the same ones the Python `yfinance` library
  wraps), no API key or registration required

**Frontend** — React 19 + TypeScript, Vite 8, D3 7 for the map projection
and choropleth rendering (no `react-simple-maps` or similar wrapper — the
map is built directly against D3 for full control over the projection and
coloring logic).

**Infrastructure** — Redis 7 (live state cache) and MongoDB 7 (historical
time series) via Docker Compose for local development.

**Data source** — Yahoo Finance's unofficial `v8/finance/chart` JSON
endpoint (`query2.finance.yahoo.com`). This is an accepted risk: it's
undocumented and Yahoo can change or block it without notice. The
hexagonal architecture exists specifically so that swapping to an official,
API-key-based provider (Finnhub, Twelve Data, Alpha Vantage) later is a
matter of writing one new adapter, not touching domain logic, workers, or
the REST API.

## Architecture

The backend follows hexagonal architecture (ports and adapters): domain
logic has zero knowledge of Yahoo, Redis, or MongoDB — it only depends on
interfaces ("ports"), and each concrete technology lives in its own
adapter implementing those interfaces.

```
/cmd
  /server                    entrypoint: wires everything, starts the API + workers
  /yahoo_smoketest           throwaway CLI to verify the Yahoo adapter in isolation
  /store_smoketest           throwaway CLI to verify Redis + Mongo end-to-end
  /market_status_smoketest   throwaway CLI checking exchange-hours edge cases

/internal
  /domain                    core entities, zero external dependencies
    quote.go                  Quote, HistoryPoint, HistoryRange
    exchange.go                Exchange, Session, MidDayBreak, ExchangeCatalog
    symbol_catalog.go           Symbol, AssetClass, Region, MasterCatalog
    market_status.go             ComputeMarketState (IANA-timezone-based)

  /ports                     interfaces the domain/app layer depends on
    quote_fetcher.go           FetchQuote(symbol) (Quote, error)
    history_fetcher.go          FetchHistory(symbol, range) ([]Point, error)
    live_cache.go                 Set/Get live quotes
    history_store.go               Append/Query historical points

  /adapters
    /inbound/http             REST handlers — translates HTTP <-> domain
    /outbound/yahoo            QuoteFetcher + HistoryFetcher implementation
    /outbound/redis             LiveCache implementation
    /outbound/mongo              HistoryStore implementation

  /app                       use cases, orchestrate domain + ports
    refresh_service.go         fetch one symbol's live quote, cache it
    query_service.go            serve the map's current state to the API
    history_service.go           keep Mongo's history series up to date

  /workers                   one independent goroutine per symbol group
    worker.go                  generic ticker/spacing/error-reporting machinery
    manager.go                   fans out workers, collects errors, graceful shutdown
    quotes_us.go, quotes_europe_asia.go,
    quotes_mexico.go, quotes_brazil.go,
    history_daily.go             five concrete worker groups (see below)

/pkg                        infrastructure-agnostic utilities
  httpclient/                 HTTP client with browser User-Agent + single retry
  logger/                      structured logging (log/slog)

/config                     environment-based runtime configuration

/web                        React + TypeScript + D3 frontend (Vite)
```

**Why one worker per group instead of a single loop:** each group runs on
its own `time.Ticker`, so a problem with one group — say, Yahoo starts
rate-limiting requests for Mexican symbols specifically — never stalls or
crashes the others, and that group's cadence can be tuned independently.
Each worker reports failures on a shared `WorkerError` channel that a
central manager collects and logs, with graceful shutdown via
`context.Context` so a `SIGINT`/`SIGTERM` cleanly stops every worker before
the process exits.

## Data source and caching strategy

The Go backend is the only thing that ever talks to Yahoo Finance. The
frontend never calls Yahoo directly — it always reads through the backend's
own REST API, which itself always reads from Redis, never from Yahoo. That
means the number of concurrent users has zero effect on how often Yahoo
gets hit: only the scheduled workers do that, once per symbol per refresh
cycle.

- **Redis ("live state")**: each worker's tick writes one `quote:{symbol}`
  key per symbol as JSON, with a TTL slightly longer than the refresh
  interval as a buffer. The REST API always reads from here.
- **MongoDB (history)**: one document per symbol, holding an array of
  `{date, close}` points. Fed by a daily job that also handles its own
  initial backfill automatically (see below) — no separate manual seeding
  step is required.
- **Polite polling**: requests to Yahoo within a single worker's tick are
  spaced ~300ms apart rather than fired all at once; a failed request gets
  one retry, and if that still fails the previously cached value keeps
  serving rather than the worker insisting or blocking. Every quote carries
  its own `lastUpdated` timestamp so staleness is always visible.

### The backfill self-triggering pattern

Rather than requiring a separate one-time backfill script, the daily
history worker's refresh logic checks whether Mongo already has any points
for a symbol before deciding what to fetch: if it's empty, it performs a
one-time full historical backfill; if it already has data, it just appends
the latest daily close. This makes the initial seed happen automatically
the first time the worker ever sees a new symbol, and keeps the daily job
itself cheap afterward.

One real bug surfaced and fixed during this work is worth documenting:
Yahoo's chart endpoint silently *coarsens* `interval=1d` data to quarterly
bars once the requested range spans multiple decades — asking for `max`
range with daily interval doesn't error, it just quietly returns far less
resolution than requested. The backfill was changed to request a bounded
10-year window instead, which is confirmed to return true daily
granularity and comfortably covers every range the API exposes.

## Symbol catalog

The MVP catalog is defined once in `internal/domain/symbol_catalog.go` and
covers, by asset class:

- **Indices (17)**: S&P 500, Nasdaq, Dow Jones, S&P/TSX (Canada), FTSE 100,
  DAX, CAC 40, IBEX 35, FTSE MIB, Nikkei 225, Hang Seng, Shanghai
  Composite, BSE Sensex, S&P/ASX 200, KOSPI, Ibovespa, and IPC México.
- **Mexico proxies**: `EWW` (iShares MSCI Mexico ETF), USD/MXN, and NYSE
  ADRs of América Móvil (`AMX`), FEMSA (`FMX`), Cemex (`CX`), Coca-Cola
  FEMSA (`KOF`), and Grupo Televisa (`TV`).
- **Additional country ETF**: `EWZ` (iShares MSCI Brazil ETF), alongside the
  Ibovespa index, as a capital-flow proxy for Brazil.

Every symbol also carries a region, country code, and (where applicable) an
exchange code that ties it to the exchange trading-hours catalog described
below.

## Market hours and timezone handling

Every tracked exchange (17 in total) has its trading hours, IANA timezone
identifier, trading days, and midday recess (where applicable — Tokyo,
Hong Kong, and Shanghai all have a lunch break; Seoul's was removed in 2000
and confirmed still absent) recorded in `internal/domain/exchange.go`, each
verified against at least two independent sources.

Market state is computed with `time.LoadLocation` + `.In()` against each
exchange's own IANA identifier — never a hardcoded UTC offset — so daylight
saving time is resolved correctly and automatically for every exchange,
including ones in the opposite hemisphere (e.g. Sydney is on its own
standard/daylight schedule, offset from the Northern Hemisphere's).

One deliberately handled edge case: Mexico (since 2022) and Brazil (since
2019) no longer observe their own daylight saving time, while the US still
does. Rather than let their markets drift out of sync with Wall Street in
UTC terms, the Bolsa Mexicana de Valores and B3 (Brazil) both manually
republish their own local trading hours twice a year, shifting by exactly
one hour, purely to stay UTC-aligned with NYSE. The exchange catalog models
this explicitly (`USDSTSync` + an alternate session), using whether
`America/New_York` is currently on daylight time as the signal for which
of the two published schedules currently applies.

Where Yahoo's endpoints do expose genuine pre-market/after-hours sessions
(NYSE, Nasdaq, NYSE Arca — the exchanges backing the ETFs and ADRs tracked
here), those are modeled too; other exchanges are treated as a simple
open/closed state, since indices generally don't have a meaningful extended
session of their own and verifying extended hours for all 17 exchanges was
out of scope for this pass.

## Background workers

Five independent worker groups run concurrently, each on its own ticker:

| Worker | Symbols | Store | Default interval |
|---|---|---|---|
| `quotes-us` | 4 (US/Canada indices) | Redis | 5 min |
| `quotes-europe-asia` | 11 (European + Asian indices) | Redis | 5 min |
| `quotes-mexico` | 8 (IPC, EWW, ADRs, USD/MXN) | Redis | 5 min |
| `quotes-brazil` | 2 (Ibovespa, EWZ) | Redis | 5 min |
| `history-daily` | 25 (entire catalog) | MongoDB | 24 h |

The four quote groups partition the full catalog with no overlap and no
gaps. `history-daily` isn't split by region, since its cadence isn't
time-sensitive the way live quotes are.

## REST API

| Method & path | Description |
|---|---|
| `GET /healthz` | Liveness check |
| `GET /api/quotes` | Every tracked symbol's current cached quote |
| `GET /api/quotes/{symbol}` | A single symbol's current quote (404 if not tracked) |
| `GET /api/history/{symbol}?range=1mo\|3mo\|1y\|max` | Daily-close series (default `3mo`; 400 on an invalid range, 404 if not tracked) |

All responses are JSON with camelCase field names. A quote looks like:

```json
{
  "symbol": "^GSPC",
  "price": 7457.69,
  "change": -76.08,
  "changePercent": -1.01,
  "currency": "USD",
  "marketState": "closed",
  "timestamp": "2026-07-17T20:43:30Z",
  "lastUpdated": "2026-07-18T03:51:27Z"
}
```

## Testing

`go test ./...` (or `make test`) runs a hermetic suite — no Redis, Mongo,
or live Yahoo access required, since it runs against in-memory fakes of the
`LiveCache`/`HistoryStore` ports. Coverage includes:

- All 4 REST endpoints, exercised through the real router and handlers
  (`net/http/httptest`) against fake ports.
- `QueryService` unit tests, including a regression guard ensuring the
  history endpoint always serializes an empty result as `[]`, never `null`.
- A regression test locking in the Yahoo `range=max` coarsening bug fix
  described above — it fails loudly if the backfill's Yahoo range mapping
  ever regresses back to `"max"`.
- Pure-logic tests for the history store's date-window filtering.

Manual/throwaway coverage exists alongside the automated suite for things
that need real infrastructure or live data: `make smoketest-yahoo`,
`make smoketest-store` (Redis + Mongo round trip), and
`make smoketest-market-status` (11 hand-built market-hours edge cases plus
a live snapshot of every exchange in the catalog). Not yet covered by
either: worker scheduling/shutdown behavior and `ComputeMarketState` itself
have no `go test` unit tests, only the manual smoketest.

## Getting started

Prerequisites: Go 1.26+, Node 20+, Docker.

```bash
# start Redis + MongoDB
docker compose up -d

# run the backend (builds and execs a real binary — see the Makefile
# comment on why this matters more than it sounds like it should)
make run

# frontend, in a separate terminal
cd web
npm install
npm run dev
```

Other useful targets: `make build`, `make test`, `make fmt`, `make vet`,
`make smoketest-yahoo`, `make smoketest-store`, `make smoketest-market-status`.

Configuration is entirely environment-variable driven (see
`config/config.go`): `REDIS_ADDR`, `MONGO_URI`, `MONGO_DATABASE`,
`HTTP_ADDR`, `REFRESH_INTERVAL`, `HISTORY_REFRESH_INTERVAL`, all with
sensible local-dev defaults.

## Project status

Backend: feature-complete for the MVP's core (steps 1-5 of the original
plan) — symbol catalog, Yahoo adapter, Redis/Mongo persistence, all five
workers, market-hours calculation, and the quotes + history REST API.

Frontend: scaffolded (Vite + React + TypeScript + D3, API client, polling
hook) and verified against the real backend end-to-end. The actual D3
choropleth map, historical scrubber, trend layer, and detail panel are not
yet built.

Not started: the capital-flow proxy layer (lowest MVP priority by design).

## Known limitations and accepted risks

- **Unofficial data source.** Yahoo Finance's `v8/finance/chart` endpoint
  is undocumented and unauthenticated; it can change or be blocked without
  notice. Accepted for this project's scope — the hexagonal architecture
  keeps the cost of switching providers contained to a single new adapter.
- **No market-state cross-check against Yahoo.** Yahoo's own live
  `marketState` field lives behind an endpoint that now requires a
  cookie+crumb exchange this project doesn't implement, so market open/
  closed state is computed entirely from this project's own exchange-hours
  model rather than verified against Yahoo's.
- **No holiday calendars.** A market holiday currently just reads as
  "closed" via the regular trading-hours window, not as a distinctly
  recognized holiday state.
- **Forex hours aren't modeled.** USD/MXN's market state is reported as
  `unknown` rather than a guess, since forex trades nearly continuously and
  modeling its actual hours was out of scope.
- **BMV/B3 seasonal schedule sync is an approximation.** Both exchanges
  announce their own local schedule-shift dates independently each year;
  this project infers the shift from whether the US is currently observing
  daylight saving time, which tracks closely but isn't guaranteed to match
  exactly on the days immediately around the actual switch.
