package yahoo

// chartResponse mirrors the JSON shape of Yahoo Finance's unofficial
// v8/finance/chart endpoint. Only the fields the adapter actually uses are
// declared; unknown fields are ignored by encoding/json.
type chartResponse struct {
	Chart struct {
		Result []chartResult `json:"result"`
		Error  *chartError   `json:"error"`
	} `json:"chart"`
}

type chartError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type chartResult struct {
	Meta       chartMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators chartIndicators `json:"indicators"`
}

type chartMeta struct {
	Currency             string  `json:"currency"`
	Symbol               string  `json:"symbol"`
	ExchangeTimezoneName string  `json:"exchangeTimezoneName"`
	RegularMarketPrice   float64 `json:"regularMarketPrice"`
	ChartPreviousClose   float64 `json:"chartPreviousClose"`
	RegularMarketTime    int64   `json:"regularMarketTime"`
}

type chartIndicators struct {
	Quote []chartQuote `json:"quote"`
}

// chartQuote holds the daily OHLCV arrays. Entries are pointers because
// Yahoo emits `null` for non-trading days inside a range.
type chartQuote struct {
	Close []*float64 `json:"close"`
}
