// Mirrors internal/domain/symbol_catalog.go's JSON shape. Static reference
// data from GET /api/symbols — join against Quote by `ticker`/`symbol` for
// anything that needs country/region/asset-class (the map, filters).

export type AssetClass = 'index' | 'etf' | 'adr' | 'forex'

export type Region = 'north_america' | 'latam' | 'europe' | 'asia'

export interface Symbol {
  ticker: string
  name: string
  assetClass: AssetClass
  region: Region
  countryCode: string // ISO 3166-1 alpha-2
  exchangeCode: string // empty for forex pairs
}
