package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"investment-map/internal/domain"
)

// fakeLiveCache and fakeHistoryStore are minimal in-memory implementations
// of the ports QueryService depends on, so these tests exercise real
// QueryService logic (tracked-symbol checks, nil->empty normalization)
// without needing Redis or Mongo running.

type fakeLiveCache struct {
	quotes map[string]domain.Quote
}

func newFakeLiveCache(quotes ...domain.Quote) *fakeLiveCache {
	m := make(map[string]domain.Quote, len(quotes))
	for _, q := range quotes {
		m[q.Symbol] = q
	}
	return &fakeLiveCache{quotes: m}
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

func newFakeHistoryStore(bySymbol map[string][]domain.HistoryPoint) *fakeHistoryStore {
	if bySymbol == nil {
		bySymbol = map[string][]domain.HistoryPoint{}
	}
	return &fakeHistoryStore{points: bySymbol}
}

func (f *fakeHistoryStore) Append(_ context.Context, symbol string, point domain.HistoryPoint) error {
	f.points[symbol] = append(f.points[symbol], point)
	return nil
}

func (f *fakeHistoryStore) Query(_ context.Context, symbol string, _ domain.HistoryRange) ([]domain.HistoryPoint, error) {
	return f.points[symbol], nil
}

func TestListQuotes_ReturnsAllTrackedQuotes(t *testing.T) {
	cache := newFakeLiveCache(
		domain.Quote{Symbol: "^GSPC", Price: 100},
		domain.Quote{Symbol: "^MXX", Price: 200},
	)
	svc := NewQueryService(cache, newFakeHistoryStore(nil), []string{"^GSPC", "^MXX"})

	got, err := svc.ListQuotes(context.Background())
	if err != nil {
		t.Fatalf("ListQuotes() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListQuotes() returned %d quotes, want 2", len(got))
	}
}

func TestGetQuote_TrackedSymbol_ReturnsQuote(t *testing.T) {
	cache := newFakeLiveCache(domain.Quote{Symbol: "^GSPC", Price: 100})
	svc := NewQueryService(cache, newFakeHistoryStore(nil), []string{"^GSPC"})

	got, err := svc.GetQuote(context.Background(), "^GSPC")
	if err != nil {
		t.Fatalf("GetQuote() error = %v", err)
	}
	if got.Price != 100 {
		t.Errorf("GetQuote() price = %v, want 100", got.Price)
	}
}

func TestGetQuote_UntrackedSymbol_ReturnsError(t *testing.T) {
	svc := NewQueryService(newFakeLiveCache(), newFakeHistoryStore(nil), []string{"^GSPC"})

	_, err := svc.GetQuote(context.Background(), "FAKE")
	if err == nil {
		t.Fatal("GetQuote() expected error for untracked symbol, got nil")
	}
}

func TestGetHistory_TrackedSymbol_ReturnsPoints(t *testing.T) {
	points := []domain.HistoryPoint{{Close: 1}, {Close: 2}}
	store := newFakeHistoryStore(map[string][]domain.HistoryPoint{"^GSPC": points})
	svc := NewQueryService(newFakeLiveCache(), store, []string{"^GSPC"})

	got, err := svc.GetHistory(context.Background(), "^GSPC", domain.HistoryRange1Month)
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("GetHistory() returned %d points, want 2", len(got))
	}
}

func TestGetHistory_UntrackedSymbol_ReturnsError(t *testing.T) {
	svc := NewQueryService(newFakeLiveCache(), newFakeHistoryStore(nil), []string{"^GSPC"})

	_, err := svc.GetHistory(context.Background(), "FAKE", domain.HistoryRange1Month)
	if err == nil {
		t.Fatal("GetHistory() expected error for untracked symbol, got nil")
	}
}

// Regression guard: the REST API must return `[]`, never `null`, for a
// symbol with no history yet — a nil slice would encode as `null` in JSON
// and break frontend code expecting an array to map/iterate over.
func TestGetHistory_NoPointsYet_ReturnsEmptySliceNotNil(t *testing.T) {
	svc := NewQueryService(newFakeLiveCache(), newFakeHistoryStore(nil), []string{"^GSPC"})

	got, err := svc.GetHistory(context.Background(), "^GSPC", domain.HistoryRange1Month)
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if got == nil {
		t.Error("GetHistory() returned nil, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Errorf("GetHistory() returned %d points, want 0", len(got))
	}
}
