import type { MarketState } from '../types/quote'

// Converts an ISO 3166-1 alpha-2 code to its flag emoji by shifting each
// letter into the Unicode Regional Indicator Symbol block — no image
// assets or network fetch needed for ~15 countries.
export function flagEmoji(alpha2: string): string {
  return alpha2
    .toUpperCase()
    .replace(/./g, (c) => String.fromCodePoint(127397 + c.charCodeAt(0)))
}

export function formatRelativeTime(iso: string, now: Date = new Date()): string {
  const diffSec = Math.max(0, Math.round((now.getTime() - new Date(iso).getTime()) / 1000))
  if (diffSec < 10) return 'just now'
  if (diffSec < 60) return `${diffSec}s ago`
  const diffMin = Math.round(diffSec / 60)
  if (diffMin < 60) return `${diffMin}m ago`
  const diffHr = Math.round(diffMin / 60)
  if (diffHr < 24) return `${diffHr}h ago`
  return `${Math.round(diffHr / 24)}d ago`
}

export const MARKET_STATE_LABEL: Record<MarketState, string> = {
  open: 'Open',
  closed: 'Closed',
  'pre-market': 'Pre-Market',
  'post-market': 'Post-Market',
  holiday: 'Holiday',
  unknown: 'Unknown',
}
