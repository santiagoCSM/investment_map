// Package redis implements ports.LiveCache: the "estado vivo" store that
// workers write to on every refresh cycle and the REST API reads from on
// every request. Yahoo is never queried per-request — only this cache is.
package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"investment-map/internal/ports"
)

type Adapter struct {
	client *goredis.Client
}

// New takes a full Redis connection URL, not a bare host:port — e.g.
// "redis://localhost:6379" for local dev, or Upstash's "rediss://
// default:PASSWORD@HOST:PORT" in production, where the "rediss" scheme is
// what actually enables TLS. A bare address has no way to carry that
// scheme, a password, or a db index, so a hardcoded host:port option
// wouldn't reach an authenticated/TLS Redis at all.
func New(url string) (*Adapter, error) {
	opts, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("redis: parse url: %w", err)
	}
	return &Adapter{client: goredis.NewClient(opts)}, nil
}

func (a *Adapter) Ping(ctx context.Context) error {
	if err := a.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: ping: %w", err)
	}
	return nil
}

func (a *Adapter) Close() error {
	return a.client.Close()
}

var _ ports.LiveCache = (*Adapter)(nil)
