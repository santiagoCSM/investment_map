package mongo

import (
	"testing"
	"time"

	"investment-map/internal/domain"
)

func TestRangeCutoff(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		rng        domain.HistoryRange
		wantCutoff time.Time
		wantHas    bool
	}{
		{"1 month", domain.HistoryRange1Month, time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC), true},
		{"3 months", domain.HistoryRange3Month, time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC), true},
		{"1 year", domain.HistoryRange1Year, time.Date(2025, 7, 18, 12, 0, 0, 0, time.UTC), true},
		{"max has no cutoff", domain.HistoryRangeMax, time.Time{}, false},
		{"unknown range has no cutoff", domain.HistoryRange("bogus"), time.Time{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cutoff, has := rangeCutoff(tt.rng, now)
			if has != tt.wantHas {
				t.Fatalf("rangeCutoff(%q) has = %v, want %v", tt.rng, has, tt.wantHas)
			}
			if has && !cutoff.Equal(tt.wantCutoff) {
				t.Errorf("rangeCutoff(%q) = %v, want %v", tt.rng, cutoff, tt.wantCutoff)
			}
		})
	}
}

func TestTruncateToDay(t *testing.T) {
	in := time.Date(2026, 7, 18, 23, 59, 59, 999, time.UTC)
	want := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)

	if got := truncateToDay(in); !got.Equal(want) {
		t.Errorf("truncateToDay(%v) = %v, want %v", in, got, want)
	}
}
