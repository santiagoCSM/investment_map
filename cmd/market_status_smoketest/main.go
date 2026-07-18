// Command market_status_smoketest checks domain.ComputeMarketState against
// known edge cases (regular session, midday break, US pre/post-market, and
// the BMV/B3 seasonal UTC-sync shift) plus a live snapshot of every
// exchange in the catalog. Run with:
//
//	go run ./cmd/market_status_smoketest
package main

import (
	"fmt"
	"log"
	"time"

	"investment-map/internal/domain"
)

type testCase struct {
	desc     string
	exchange string
	when     time.Time
	want     domain.MarketState
}

func main() {
	loc := func(name string) *time.Location {
		l, err := time.LoadLocation(name)
		if err != nil {
			log.Fatalf("load location %s: %v", name, err)
		}
		return l
	}

	nyLoc := loc("America/New_York")
	tkLoc := loc("Asia/Tokyo")
	mxLoc := loc("America/Mexico_City")

	cases := []testCase{
		{"NYSE Tue 10:00 ET (regular session)", "XNYS", time.Date(2026, 1, 13, 10, 0, 0, 0, nyLoc), domain.MarketStateOpen},
		{"NYSE Sat 10:00 ET (weekend)", "XNYS", time.Date(2026, 1, 17, 10, 0, 0, 0, nyLoc), domain.MarketStateClosed},
		{"NYSE Tue 07:00 ET (pre-market)", "XNYS", time.Date(2026, 1, 13, 7, 0, 0, 0, nyLoc), domain.MarketStatePreMarket},
		{"NYSE Tue 18:00 ET (post-market)", "XNYS", time.Date(2026, 1, 13, 18, 0, 0, 0, nyLoc), domain.MarketStatePostMarket},
		{"NYSE Tue 02:00 ET (fully closed, no session)", "XNYS", time.Date(2026, 1, 13, 2, 0, 0, 0, nyLoc), domain.MarketStateClosed},
		{"Tokyo Wed 10:00 JST (morning session)", "XTKS", time.Date(2026, 1, 14, 10, 0, 0, 0, tkLoc), domain.MarketStateOpen},
		{"Tokyo Wed 12:00 JST (lunch recess)", "XTKS", time.Date(2026, 1, 14, 12, 0, 0, 0, tkLoc), domain.MarketStateClosed},
		{"Tokyo Wed 13:00 JST (afternoon session)", "XTKS", time.Date(2026, 1, 14, 13, 0, 0, 0, tkLoc), domain.MarketStateOpen},
		{"BMV Wed 08:00 CDMX in January (US standard time -> 08:30-15:00 session, not open yet)", "XMEX", time.Date(2026, 1, 14, 8, 0, 0, 0, mxLoc), domain.MarketStateClosed},
		{"BMV Wed 08:00 CDMX in July (US DST -> 07:30-14:00 session, open)", "XMEX", time.Date(2026, 7, 15, 8, 0, 0, 0, mxLoc), domain.MarketStateOpen},
		{"BMV Wed 08:45 CDMX in January (US standard time session, now open)", "XMEX", time.Date(2026, 1, 14, 8, 45, 0, 0, mxLoc), domain.MarketStateOpen},
	}

	fmt.Println("=== Edge case checks ===")
	failures := 0
	for _, tc := range cases {
		ex, ok := domain.ExchangeCatalog[tc.exchange]
		if !ok {
			log.Fatalf("unknown exchange %s in test case %q", tc.exchange, tc.desc)
		}
		got := domain.ComputeMarketState(ex, tc.when)
		status := "OK"
		if got != tc.want {
			status = "FAIL"
			failures++
		}
		fmt.Printf("[%s] %-70s want=%-12s got=%-12s\n", status, tc.desc, tc.want, got)
	}

	fmt.Println("\n=== Live snapshot (time.Now()) for every catalog exchange ===")
	now := time.Now()
	seen := map[string]bool{}
	for _, sym := range domain.MasterCatalog {
		if sym.ExchangeCode == "" || seen[sym.ExchangeCode] {
			continue
		}
		seen[sym.ExchangeCode] = true
		ex, ok := domain.ExchangeCatalog[sym.ExchangeCode]
		if !ok {
			fmt.Printf("%-6s %-35s NO EXCHANGE DATA\n", sym.ExchangeCode, sym.Name)
			continue
		}
		state := domain.ComputeMarketState(ex, now)
		loc, _ := time.LoadLocation(ex.IANATimeZone)
		fmt.Printf("%-6s %-35s local=%s state=%s\n", ex.Code, ex.Name, now.In(loc).Format("2006-01-02 15:04 MST"), state)
	}

	if failures > 0 {
		log.Fatalf("\n%d/%d edge cases FAILED", failures, len(cases))
	}
	fmt.Printf("\nall %d edge cases passed\n", len(cases))
}
