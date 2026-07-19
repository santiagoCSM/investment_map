package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"investment-map/internal/app"
	"investment-map/internal/domain"
)

// fakeLiveCache/fakeHistoryStore mirror the ones in
// internal/app/query_service_test.go — small enough that a shared
// test-only package isn't worth the indirection, and keeping them local
// makes each test package's fakes explicit at a glance.

type fakeLiveCache struct {
	quotes map[string]domain.Quote
}

func (f *fakeLiveCache) SetQuote(_ context.Context, quote domain.Quote, _ time.Duration) error {
	f.quotes[quote.Symbol] = quote
	return nil
}

func (f *fakeLiveCache) GetQuote(_ context.Context, symbol string) (domain.Quote, error) {
	q, ok := f.quotes[symbol]
	if !ok {
		return domain.Quote{}, errors.New("not cached")
	}
	return q, nil
}

func (f *fakeLiveCache) GetQuotes(_ context.Context, symbols []string) ([]domain.Quote, error) {
	var out []domain.Quote
	for _, s := range symbols {
		if q, ok := f.quotes[s]; ok {
			out = append(out, q)
		}
	}
	return out, nil
}

type fakeHistoryStore struct {
	points map[string][]domain.HistoryPoint
}

func (f *fakeHistoryStore) Append(_ context.Context, symbol string, point domain.HistoryPoint) error {
	f.points[symbol] = append(f.points[symbol], point)
	return nil
}

func (f *fakeHistoryStore) Query(_ context.Context, symbol string, _ domain.HistoryRange) ([]domain.HistoryPoint, error) {
	return f.points[symbol], nil
}

// newTestRouter wires a real app.QueryService (real routing/handler/JSON
// logic) against fake ports, so these tests exercise the full HTTP request
// path without needing Redis or Mongo running.
func newTestRouter() http.Handler {
	cache := &fakeLiveCache{quotes: map[string]domain.Quote{
		"^GSPC": {Symbol: "^GSPC", Price: 7457.69, Currency: "USD", MarketState: domain.MarketStateClosed},
	}}
	store := &fakeHistoryStore{points: map[string][]domain.HistoryPoint{
		"^GSPC": {{Close: 100}, {Close: 200}},
	}}
	query := app.NewQueryService(cache, store, []string{"^GSPC"})
	return NewRouter(query)
}

func doRequest(t *testing.T, router http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestHealthEndpoint(t *testing.T) {
	rec := doRequest(t, newTestRouter(), "/healthz")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestListQuotesEndpoint(t *testing.T) {
	rec := doRequest(t, newTestRouter(), "/api/quotes")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var quotes []domain.Quote
	if err := json.Unmarshal(rec.Body.Bytes(), &quotes); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(quotes) != 1 {
		t.Fatalf("got %d quotes, want 1", len(quotes))
	}
	if quotes[0].Symbol != "^GSPC" {
		t.Errorf("quote symbol = %q, want %q", quotes[0].Symbol, "^GSPC")
	}
}

func TestGetQuoteEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"tracked symbol", "/api/quotes/%5EGSPC", http.StatusOK},
		{"untracked symbol", "/api/quotes/FAKE", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, newTestRouter(), tt.path)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestGetHistoryEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"default range", "/api/history/%5EGSPC", http.StatusOK},
		{"explicit 1mo", "/api/history/%5EGSPC?range=1mo", http.StatusOK},
		{"explicit 3mo", "/api/history/%5EGSPC?range=3mo", http.StatusOK},
		{"explicit 1y", "/api/history/%5EGSPC?range=1y", http.StatusOK},
		{"explicit max", "/api/history/%5EGSPC?range=max", http.StatusOK},
		{"invalid range", "/api/history/%5EGSPC?range=bogus", http.StatusBadRequest},
		{"untracked symbol", "/api/history/FAKE", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, newTestRouter(), tt.path)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestGetHistoryEndpoint_HappyPathReturnsPoints(t *testing.T) {
	rec := doRequest(t, newTestRouter(), "/api/history/%5EGSPC")

	var points []domain.HistoryPoint
	if err := json.Unmarshal(rec.Body.Bytes(), &points); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("got %d points, want 2", len(points))
	}
}

// Regression guard: a symbol with no history yet must serialize as `[]`,
// never `null` — a nil slice would break frontend code expecting an array.
func TestGetHistoryEndpoint_NoHistoryYet_ReturnsEmptyArrayNotNull(t *testing.T) {
	cache := &fakeLiveCache{quotes: map[string]domain.Quote{}}
	store := &fakeHistoryStore{points: map[string][]domain.HistoryPoint{}}
	query := app.NewQueryService(cache, store, []string{"^GSPC"})
	router := NewRouter(query)

	rec := doRequest(t, router, "/api/history/%5EGSPC")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); got != "[]\n" {
		t.Errorf("body = %q, want %q", got, "[]\n")
	}
}
