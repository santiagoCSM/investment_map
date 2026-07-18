// Package httpclient provides a minimal HTTP client with a browser
// User-Agent and a single retry, intended for outbound adapters that talk
// to rate-sensitive or unofficial APIs.
package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

type Client struct {
	http      *http.Client
	userAgent string
	retryWait time.Duration
}

func New() *Client {
	return &Client{
		http:      &http.Client{Timeout: 10 * time.Second},
		userAgent: defaultUserAgent,
		retryWait: 500 * time.Millisecond,
	}
}

// Get performs a GET request, retrying once after a short backoff if the
// first attempt errors or returns a non-2xx status. Callers that want a
// stale-cache fallback instead of an error should catch the returned error
// and read their own cache.
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	body, err := c.get(ctx, url)
	if err == nil {
		return body, nil
	}

	select {
	case <-time.After(c.retryWait):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return c.get(ctx, url)
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("httpclient: GET %s: status %d: %s", url, resp.StatusCode, string(body))
	}
	return body, nil
}
