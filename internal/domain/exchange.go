package domain

import "time"

// Session is a single continuous trading window expressed in the
// exchange's own local wall-clock time ("HH:MM", 24h).
type Session struct {
	Open  string
	Close string
}

// MidDayBreak is a lunch recess within the regular session (Tokyo, Hong
// Kong, Shanghai), also in local wall-clock time.
type MidDayBreak struct {
	Start string
	End   string
}

// Exchange describes a stock exchange's trading calendar in terms of an
// IANA timezone identifier, letting the standard library resolve DST rules
// instead of hardcoding UTC offsets.
type Exchange struct {
	Code         string // MIC-style code, e.g. "XNYS", "XLON", "XTKS"
	Name         string
	IANATimeZone string // e.g. "America/New_York"
	TradingDays  []time.Weekday
	Session      Session
	MidDayBreak  *MidDayBreak // nil when the exchange has no lunch recess

	// PreMarket/PostMarket are only modeled where Yahoo actually surfaces
	// extended-hours data for the instruments we track there (the NYSE/
	// Nasdaq/ARCA-listed ETFs and ADRs) — nil everywhere else. Indices
	// don't have a meaningful pre/post session of their own, and verifying
	// extended hours for all 17 exchanges was out of scope for this pass.
	PreMarket  *Session
	PostMarket *Session

	// USDSTSync marks exchanges that manually republish their own local
	// trading hours twice a year purely to stay UTC-aligned with NYSE —
	// because their home country stopped observing its own DST while the
	// US still does (Mexico since 2022, Brazil since 2019). When true,
	// SessionDuringUSStandardTime is used instead of Session whenever
	// America/New_York is currently on standard time (EST, not EDT).
	//
	// This is an approximation: BMV and B3 each announce their actual
	// switch dates independently a few weeks ahead of the change, and
	// those dates track the US DST calendar closely but aren't guaranteed
	// to match it exactly. Good enough for the MVP; revisit with each
	// exchange's own published calendar if exact-day precision matters.
	USDSTSync                   bool
	SessionDuringUSStandardTime Session
}

var weekdaysMonFri = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}

// ExchangeCatalog holds every exchange referenced by Symbol.ExchangeCode in
// MasterCatalog, keyed by that same code. Trading hours were verified
// against at least two independent sources per exchange (see project notes)
// as of July 2026; see the USDSTSync doc comment for the one deliberate
// approximation.
var ExchangeCatalog = map[string]Exchange{
	"XNYS": {
		Code: "XNYS", Name: "New York Stock Exchange", IANATimeZone: "America/New_York",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:30", Close: "16:00"},
		PreMarket: &Session{Open: "04:00", Close: "09:30"}, PostMarket: &Session{Open: "16:00", Close: "20:00"},
	},
	"XNAS": {
		Code: "XNAS", Name: "Nasdaq", IANATimeZone: "America/New_York",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:30", Close: "16:00"},
		PreMarket: &Session{Open: "04:00", Close: "09:30"}, PostMarket: &Session{Open: "16:00", Close: "20:00"},
	},
	"ARCX": {
		Code: "ARCX", Name: "NYSE Arca", IANATimeZone: "America/New_York",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:30", Close: "16:00"},
		PreMarket: &Session{Open: "04:00", Close: "09:30"}, PostMarket: &Session{Open: "16:00", Close: "20:00"},
	},
	"XTSE": {
		Code: "XTSE", Name: "Toronto Stock Exchange", IANATimeZone: "America/Toronto",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:30", Close: "16:00"},
	},
	"XLON": {
		Code: "XLON", Name: "London Stock Exchange", IANATimeZone: "Europe/London",
		TradingDays: weekdaysMonFri, Session: Session{Open: "08:00", Close: "16:30"},
	},
	"XFRA": {
		// Xetra electronic trading; Germany has no "Europe/Frankfurt" zone
		// in IANA tzdata, "Europe/Berlin" is the correct identifier.
		Code: "XFRA", Name: "Deutsche Börse (Xetra)", IANATimeZone: "Europe/Berlin",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:00", Close: "17:30"},
	},
	"XPAR": {
		Code: "XPAR", Name: "Euronext Paris", IANATimeZone: "Europe/Paris",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:00", Close: "17:30"},
	},
	"XMAD": {
		Code: "XMAD", Name: "Bolsa de Madrid", IANATimeZone: "Europe/Madrid",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:00", Close: "17:30"},
	},
	"XMIL": {
		Code: "XMIL", Name: "Borsa Italiana (Euronext Milan)", IANATimeZone: "Europe/Rome",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:00", Close: "17:30"},
	},
	"XTKS": {
		// Close extended 15:00 → 15:30 on 2024-11-05, first extension
		// since 1954. Verify again if this data goes stale for years.
		Code: "XTKS", Name: "Tokyo Stock Exchange", IANATimeZone: "Asia/Tokyo",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:00", Close: "15:30"},
		MidDayBreak: &MidDayBreak{Start: "11:30", End: "12:30"},
	},
	"XHKG": {
		Code: "XHKG", Name: "Hong Kong Stock Exchange", IANATimeZone: "Asia/Hong_Kong",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:30", Close: "16:00"},
		MidDayBreak: &MidDayBreak{Start: "12:00", End: "13:00"},
	},
	"XSHG": {
		// SSE's last 3 minutes (14:57-15:00) are technically a closing
		// call auction rather than continuous trading; treated as regular
		// session close for MVP purposes.
		Code: "XSHG", Name: "Shanghai Stock Exchange", IANATimeZone: "Asia/Shanghai",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:30", Close: "15:00"},
		MidDayBreak: &MidDayBreak{Start: "11:30", End: "13:00"},
	},
	"XBOM": {
		Code: "XBOM", Name: "BSE (Bombay Stock Exchange)", IANATimeZone: "Asia/Kolkata",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:15", Close: "15:30"},
	},
	"XASX": {
		Code: "XASX", Name: "Australian Securities Exchange", IANATimeZone: "Australia/Sydney",
		TradingDays: weekdaysMonFri, Session: Session{Open: "10:00", Close: "16:00"},
	},
	"XKRX": {
		// Korea Exchange removed its midday break in 2000; confirmed
		// still absent as of this research pass, not a stale assumption.
		Code: "XKRX", Name: "Korea Exchange (Seoul)", IANATimeZone: "Asia/Seoul",
		TradingDays: weekdaysMonFri, Session: Session{Open: "09:00", Close: "15:30"},
	},
	"BVMF": {
		// B3 shifts close time to stay UTC-aligned with NYSE since Brazil
		// dropped its own DST in 2019 but the US didn't. Open is fixed at
		// 10:00 year-round either way.
		Code: "BVMF", Name: "B3 (Brasil Bolsa Balcão)", IANATimeZone: "America/Sao_Paulo",
		TradingDays: weekdaysMonFri, Session: Session{Open: "10:00", Close: "17:00"},
		USDSTSync: true, SessionDuringUSStandardTime: Session{Open: "10:00", Close: "18:00"},
	},
	"XMEX": {
		// Bolsa Mexicana de Valores shifts its whole session by an hour to
		// stay UTC-aligned with NYSE since Mexico dropped its own DST in
		// 2022 but the US didn't. Both 07:30-14:00 and 08:30-15:00 are
		// correct — they're the same underlying UTC window, just for
		// opposite halves of the year.
		Code: "XMEX", Name: "Bolsa Mexicana de Valores", IANATimeZone: "America/Mexico_City",
		TradingDays: weekdaysMonFri, Session: Session{Open: "07:30", Close: "14:00"},
		USDSTSync: true, SessionDuringUSStandardTime: Session{Open: "08:30", Close: "15:00"},
	},
}
