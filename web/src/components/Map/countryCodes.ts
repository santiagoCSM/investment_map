// Crosswalk from ISO 3166-1 numeric (the `id` on each feature in
// world-countries-50m.json, sourced from the `world-atlas` npm package)
// to ISO 3166-1 alpha-2 (Symbol.CountryCode in the Go backend's master
// catalog). Only the ~15 countries actually present in the catalog are
// listed — verified by name against the topology file, not guessed.
export const NUMERIC_TO_ALPHA2: Record<string, string> = {
  '840': 'US', // United States of America
  '124': 'CA', // Canada
  '826': 'GB', // United Kingdom
  '276': 'DE', // Germany
  '250': 'FR', // France
  '724': 'ES', // Spain
  '380': 'IT', // Italy
  '392': 'JP', // Japan
  '344': 'HK', // Hong Kong
  '156': 'CN', // China
  '356': 'IN', // India
  '036': 'AU', // Australia (also covers Ashmore and Cartier Is., which
  //             shares Australia's ISO code as a separate landmass in
  //             this topology — same alpha-2, so it colors consistently)
  '410': 'KR', // South Korea
  '076': 'BR', // Brazil
  '484': 'MX', // Mexico
}
