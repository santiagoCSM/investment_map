package domain

import (
	"strconv"
	"strings"
	"time"
)

// ComputeMarketState determines whether ex is open, closed, in a midday
// recess, or in a pre/post-market session at instant t. It resolves DST
// entirely through ex.IANATimeZone via time.LoadLocation + .In() — no
// hardcoded UTC offsets — so the same code handles exchanges that observe
// DST (NYSE, LSE) and ones that don't (Mexico City since 2022) correctly
// as of whatever date it runs.
//
// Holiday calendars are out of scope for the MVP: a holiday reads as
// "closed" because it falls outside the session window, not because it's
// specifically recognized as a holiday.
func ComputeMarketState(ex Exchange, t time.Time) MarketState {
	loc, err := time.LoadLocation(ex.IANATimeZone)
	if err != nil {
		return MarketStateUnknown
	}
	local := t.In(loc)

	if !isTradingDay(ex, local.Weekday()) {
		return MarketStateClosed
	}

	session := resolveSession(ex, t)

	open, ok := atLocalTime(local, session.Open)
	if !ok {
		return MarketStateUnknown
	}
	close_, ok := atLocalTime(local, session.Close)
	if !ok {
		return MarketStateUnknown
	}

	if ex.MidDayBreak != nil {
		if inWindow, ok := withinLocalWindow(local, ex.MidDayBreak.Start, ex.MidDayBreak.End); ok && inWindow {
			return MarketStateClosed
		}
	}

	if !local.Before(open) && local.Before(close_) {
		return MarketStateOpen
	}

	if ex.PreMarket != nil {
		if inWindow, ok := withinLocalWindow(local, ex.PreMarket.Open, ex.PreMarket.Close); ok && inWindow {
			return MarketStatePreMarket
		}
	}

	if ex.PostMarket != nil {
		if inWindow, ok := withinLocalWindow(local, ex.PostMarket.Open, ex.PostMarket.Close); ok && inWindow {
			return MarketStatePostMarket
		}
	}

	return MarketStateClosed
}

func isTradingDay(ex Exchange, day time.Weekday) bool {
	for _, d := range ex.TradingDays {
		if d == day {
			return true
		}
	}
	return false
}

// resolveSession picks ex.Session, except for USDSTSync exchanges (BMV,
// B3): see the field's doc comment in exchange.go for why and for the
// approximation this relies on.
func resolveSession(ex Exchange, t time.Time) Session {
	if ex.USDSTSync && !isUSEasternDST(t) {
		return ex.SessionDuringUSStandardTime
	}
	return ex.Session
}

func isUSEasternDST(t time.Time) bool {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return false
	}
	name, _ := t.In(loc).Zone()
	return name == "EDT"
}

func withinLocalWindow(local time.Time, startHHMM, endHHMM string) (inWindow bool, ok bool) {
	start, ok1 := atLocalTime(local, startHHMM)
	end, ok2 := atLocalTime(local, endHHMM)
	if !ok1 || !ok2 {
		return false, false
	}
	return !local.Before(start) && local.Before(end), true
}

// atLocalTime combines local's own date with a "HH:MM" wall-clock time, in
// local's own location, so it can be compared directly against local.
func atLocalTime(local time.Time, hhmm string) (time.Time, bool) {
	hour, minute, ok := parseHHMM(hhmm)
	if !ok {
		return time.Time{}, false
	}
	return time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, local.Location()), true
}

func parseHHMM(hhmm string) (hour, minute int, ok bool) {
	parts := strings.SplitN(hhmm, ":", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}
	minute, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}
	return hour, minute, true
}
