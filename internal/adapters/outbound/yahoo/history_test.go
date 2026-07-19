package yahoo

import (
	"testing"

	"investment-map/internal/domain"
)

// Regression guard for a real bug: Yahoo's v8/chart endpoint silently
// coarsens interval=1d data to quarterly bars once range=max spans
// decades (confirmed live: 168 points for ^GSPC back to 1985 with ~90-day
// gaps, vs. 2514 points with real 1-day gaps for range=10y). The backfill
// must keep requesting "10y" from Yahoo, never "max" — this test fails
// loudly if that mapping ever gets "simplified" back to max.
func TestToYahooRange(t *testing.T) {
	tests := []struct {
		name    string
		rng     domain.HistoryRange
		want    string
		wantErr bool
	}{
		{"1 month", domain.HistoryRange1Month, "1mo", false},
		{"3 months", domain.HistoryRange3Month, "3mo", false},
		{"1 year", domain.HistoryRange1Year, "1y", false},
		{"max", domain.HistoryRangeMax, "max", false},
		{"10 year (backfill range)", domain.HistoryRange10Year, "10y", false},
		{"unsupported", domain.HistoryRange("bogus"), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toYahooRange(tt.rng)
			if (err != nil) != tt.wantErr {
				t.Fatalf("toYahooRange(%q) error = %v, wantErr %v", tt.rng, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("toYahooRange(%q) = %q, want %q", tt.rng, got, tt.want)
			}
		})
	}
}
