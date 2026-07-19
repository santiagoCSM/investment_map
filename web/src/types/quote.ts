// Mirrors internal/domain/quote.go's JSON shape exactly (see the `json:`
// tags there) — keep these two in sync by hand, there's no shared schema
// generation between the Go backend and this frontend.

export type MarketState =
  | 'open'
  | 'closed'
  | 'pre-market'
  | 'post-market'
  | 'holiday'
  | 'unknown'

export interface Quote {
  symbol: string
  price: number
  change: number
  changePercent: number
  currency: string
  marketState: MarketState
  timestamp: string // RFC3339, e.g. "2026-07-17T20:43:30Z"
  lastUpdated: string
}

export interface HistoryPoint {
  date: string
  close: number
}
