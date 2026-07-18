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

func New(addr string) *Adapter {
	return &Adapter{
		client: goredis.NewClient(&goredis.Options{Addr: addr}),
	}
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
