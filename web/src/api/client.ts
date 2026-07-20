import type { HistoryPoint, Quote } from '../types/quote'
import type { Symbol } from '../types/symbol'

// Same-origin in prod (backend serves the built frontend), proxied by Vite
// in dev (see vite.config.ts) — so this never needs an absolute URL or CORS
// config on either side.
const API_BASE = '/api'

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function getJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`)
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(body.error ?? res.statusText, res.status)
  }
  return res.json() as Promise<T>
}

export function fetchQuotes(): Promise<Quote[]> {
  return getJSON<Quote[]>('/quotes')
}

export function fetchQuote(symbol: string): Promise<Quote> {
  return getJSON<Quote>(`/quotes/${encodeURIComponent(symbol)}`)
}

export type HistoryRange = '1mo' | '3mo' | '1y' | 'max'

export function fetchHistory(symbol: string, range?: HistoryRange): Promise<HistoryPoint[]> {
  const query = range ? `?range=${range}` : ''
  return getJSON<HistoryPoint[]>(`/history/${encodeURIComponent(symbol)}${query}`)
}

export function fetchSymbols(): Promise<Symbol[]> {
  return getJSON<Symbol[]>('/symbols')
}
