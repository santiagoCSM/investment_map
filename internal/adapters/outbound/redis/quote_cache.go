package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"investment-map/internal/domain"
)

func quoteKey(symbol string) string {
	return "quote:" + symbol
}

// SetQuote stores quote as JSON under quote:{symbol} with the given TTL.
// Workers call this once per symbol per refresh cycle.
func (a *Adapter) SetQuote(ctx context.Context, quote domain.Quote, ttl time.Duration) error {
	body, err := json.Marshal(quote)
	if err != nil {
		return fmt.Errorf("redis: marshal quote %s: %w", quote.Symbol, err)
	}
	if err := a.client.Set(ctx, quoteKey(quote.Symbol), body, ttl).Err(); err != nil {
		return fmt.Errorf("redis: set quote %s: %w", quote.Symbol, err)
	}
	return nil
}

// GetQuote reads the cached quote for symbol. Returns an error if the key
// is missing or has expired (TTL exceeded) — callers decide whether that
// means "not tracked" or "worker is behind".
func (a *Adapter) GetQuote(ctx context.Context, symbol string) (domain.Quote, error) {
	body, err := a.client.Get(ctx, quoteKey(symbol)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return domain.Quote{}, fmt.Errorf("redis: quote %s not cached", symbol)
	}
	if err != nil {
		return domain.Quote{}, fmt.Errorf("redis: get quote %s: %w", symbol, err)
	}

	var quote domain.Quote
	if err := json.Unmarshal(body, &quote); err != nil {
		return domain.Quote{}, fmt.Errorf("redis: unmarshal quote %s: %w", symbol, err)
	}
	return quote, nil
}

// GetQuotes reads cached quotes for a batch of symbols in one round trip.
// Symbols with no cached value (never fetched, or TTL expired) are silently
// omitted rather than failing the whole batch — the map view should still
// render the symbols that are healthy.
func (a *Adapter) GetQuotes(ctx context.Context, symbols []string) ([]domain.Quote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	keys := make([]string, len(symbols))
	for i, s := range symbols {
		keys[i] = quoteKey(s)
	}

	raw, err := a.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis: mget quotes: %w", err)
	}

	quotes := make([]domain.Quote, 0, len(raw))
	for i, v := range raw {
		if v == nil {
			continue // not cached or expired
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		var quote domain.Quote
		if err := json.Unmarshal([]byte(s), &quote); err != nil {
			return nil, fmt.Errorf("redis: unmarshal quote %s: %w", symbols[i], err)
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}
