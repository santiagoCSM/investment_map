import type { Quote } from '../types/quote'

// Same-origin in prod (backend serves the built frontend), proxied by Vite
// in dev (see vite.config.ts) — so this never needs an absolute URL or CORS
// config on either side.
const API_BASE = '/api'

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message)
    this.name = 'ApiError'
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
