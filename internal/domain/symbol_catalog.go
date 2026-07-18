package domain

// AssetClass categorizes a tracked symbol for grouping and display purposes.
type AssetClass string

const (
	AssetClassIndex AssetClass = "index"
	AssetClassETF   AssetClass = "etf"
	AssetClassADR   AssetClass = "adr"
	AssetClassForex AssetClass = "forex"
)

// Region is used to bucket symbols for the map and for worker grouping.
type Region string

const (
	RegionNorthAmerica Region = "north_america"
	RegionLatam        Region = "latam"
	RegionEurope       Region = "europe"
	RegionAsia         Region = "asia"
)

// Symbol is a single tracked instrument in the master catalog.
type Symbol struct {
	Ticker       string // Yahoo Finance ticker, e.g. "^GSPC", "EWW", "USDMXN=X"
	Name         string
	AssetClass   AssetClass
	Region       Region
	CountryCode  string // ISO 3166-1 alpha-2, used to key the choropleth map
	ExchangeCode string // references Exchange.Code, empty for forex pairs
}

// MasterCatalog is the full set of instruments tracked by the MVP.
// Worker groups (quotes_us, quotes_europe_asia, quotes_mexico) each pull
// their slice of this catalog by Region/AssetClass.
var MasterCatalog = []Symbol{
	// Major world indices
	{Ticker: "^GSPC", Name: "S&P 500", AssetClass: AssetClassIndex, Region: RegionNorthAmerica, CountryCode: "US", ExchangeCode: "XNYS"},
	{Ticker: "^IXIC", Name: "Nasdaq Composite", AssetClass: AssetClassIndex, Region: RegionNorthAmerica, CountryCode: "US", ExchangeCode: "XNAS"},
	{Ticker: "^DJI", Name: "Dow Jones Industrial Average", AssetClass: AssetClassIndex, Region: RegionNorthAmerica, CountryCode: "US", ExchangeCode: "XNYS"},
	{Ticker: "^GSPTSE", Name: "S&P/TSX Composite", AssetClass: AssetClassIndex, Region: RegionNorthAmerica, CountryCode: "CA", ExchangeCode: "XTSE"},
	{Ticker: "^FTSE", Name: "FTSE 100", AssetClass: AssetClassIndex, Region: RegionEurope, CountryCode: "GB", ExchangeCode: "XLON"},
	{Ticker: "^GDAXI", Name: "DAX", AssetClass: AssetClassIndex, Region: RegionEurope, CountryCode: "DE", ExchangeCode: "XFRA"},
	{Ticker: "^FCHI", Name: "CAC 40", AssetClass: AssetClassIndex, Region: RegionEurope, CountryCode: "FR", ExchangeCode: "XPAR"},
	{Ticker: "^IBEX", Name: "IBEX 35", AssetClass: AssetClassIndex, Region: RegionEurope, CountryCode: "ES", ExchangeCode: "XMAD"},
	{Ticker: "FTSEMIB.MI", Name: "FTSE MIB", AssetClass: AssetClassIndex, Region: RegionEurope, CountryCode: "IT", ExchangeCode: "XMIL"},
	{Ticker: "^N225", Name: "Nikkei 225", AssetClass: AssetClassIndex, Region: RegionAsia, CountryCode: "JP", ExchangeCode: "XTKS"},
	{Ticker: "^HSI", Name: "Hang Seng", AssetClass: AssetClassIndex, Region: RegionAsia, CountryCode: "HK", ExchangeCode: "XHKG"},
	{Ticker: "000001.SS", Name: "Shanghai Composite", AssetClass: AssetClassIndex, Region: RegionAsia, CountryCode: "CN", ExchangeCode: "XSHG"},
	{Ticker: "^BSESN", Name: "BSE Sensex", AssetClass: AssetClassIndex, Region: RegionAsia, CountryCode: "IN", ExchangeCode: "XBOM"},
	{Ticker: "^AXJO", Name: "S&P/ASX 200", AssetClass: AssetClassIndex, Region: RegionAsia, CountryCode: "AU", ExchangeCode: "XASX"},
	{Ticker: "^KS11", Name: "KOSPI", AssetClass: AssetClassIndex, Region: RegionAsia, CountryCode: "KR", ExchangeCode: "XKRX"},
	{Ticker: "^BVSP", Name: "Ibovespa", AssetClass: AssetClassIndex, Region: RegionLatam, CountryCode: "BR", ExchangeCode: "BVMF"},
	{Ticker: "^MXX", Name: "IPC Mexico", AssetClass: AssetClassIndex, Region: RegionLatam, CountryCode: "MX", ExchangeCode: "XMEX"},

	// Mexico proxies (ETF, ADRs, forex)
	{Ticker: "EWW", Name: "iShares MSCI Mexico ETF", AssetClass: AssetClassETF, Region: RegionLatam, CountryCode: "MX", ExchangeCode: "ARCX"},
	{Ticker: "AMX", Name: "America Movil ADR", AssetClass: AssetClassADR, Region: RegionLatam, CountryCode: "MX", ExchangeCode: "XNYS"},
	{Ticker: "FMX", Name: "FEMSA ADR", AssetClass: AssetClassADR, Region: RegionLatam, CountryCode: "MX", ExchangeCode: "XNYS"},
	{Ticker: "CX", Name: "Cemex ADR", AssetClass: AssetClassADR, Region: RegionLatam, CountryCode: "MX", ExchangeCode: "XNYS"},
	{Ticker: "KOF", Name: "Coca-Cola FEMSA ADR", AssetClass: AssetClassADR, Region: RegionLatam, CountryCode: "MX", ExchangeCode: "XNYS"},
	{Ticker: "TV", Name: "Grupo Televisa ADR", AssetClass: AssetClassADR, Region: RegionLatam, CountryCode: "MX", ExchangeCode: "XNYS"},
	{Ticker: "USDMXN=X", Name: "USD/MXN", AssetClass: AssetClassForex, Region: RegionLatam, CountryCode: "MX", ExchangeCode: ""},

	// Additional country ETFs (capital flow proxy)
	{Ticker: "EWZ", Name: "iShares MSCI Brazil ETF", AssetClass: AssetClassETF, Region: RegionLatam, CountryCode: "BR", ExchangeCode: "ARCX"},
}

var symbolIndex = buildSymbolIndex()

func buildSymbolIndex() map[string]Symbol {
	idx := make(map[string]Symbol, len(MasterCatalog))
	for _, s := range MasterCatalog {
		idx[s.Ticker] = s
	}
	return idx
}

// FindSymbol looks up a catalog entry by its Yahoo Finance ticker.
func FindSymbol(ticker string) (Symbol, bool) {
	s, ok := symbolIndex[ticker]
	return s, ok
}
