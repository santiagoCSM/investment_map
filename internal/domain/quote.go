package domain

import "time"

// MarketState represents whether an exchange is open, closed, or in a
// pre/post-market session at the moment a Quote was computed.
type MarketState string

const (
	MarketStateOpen       MarketState = "open"
	MarketStateClosed     MarketState = "closed"
	MarketStatePreMarket  MarketState = "pre-market"
	MarketStatePostMarket MarketState = "post-market"
	MarketStateHoliday    MarketState = "holiday"
	MarketStateUnknown    MarketState = "unknown"
)

// Quote is a point-in-time snapshot of a tracked symbol. JSON tags define
// the REST API's wire format directly, since the API returns domain types.
type Quote struct {
	Symbol        string      `json:"symbol"`
	Price         float64     `json:"price"`
	Change        float64     `json:"change"`
	ChangePercent float64     `json:"changePercent"`
	Currency      string      `json:"currency"`
	MarketState   MarketState `json:"marketState"`
	Timestamp     time.Time   `json:"timestamp"`
	LastUpdated   time.Time   `json:"lastUpdated"`
}

// HistoryRange is a coarse lookback window requested from a HistoryFetcher.
type HistoryRange string

const (
	HistoryRange1Month HistoryRange = "1mo"
	HistoryRange3Month HistoryRange = "3mo"
	HistoryRange1Year  HistoryRange = "1y"
	HistoryRangeMax    HistoryRange = "max"
)

// HistoryPoint is a single day's close in a symbol's historical series.
type HistoryPoint struct {
	Date  time.Time `json:"date"`
	Close float64   `json:"close"`
}
