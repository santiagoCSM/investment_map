import { useEffect, useRef, useState } from 'react'
import { fetchQuotes } from '../api/client'
import type { Quote } from '../types/quote'

// The backend only refreshes Redis every 5min (see config.REFRESH_INTERVAL),
// so polling much faster than that just re-reads the same cached values.
// 60s keeps the UI feeling responsive without hammering the API for
// nothing.
const DEFAULT_POLL_MS = 60_000

interface UseQuotesResult {
  quotes: Quote[]
  loading: boolean
  error: string | null
}

export function useQuotes(pollMs: number = DEFAULT_POLL_MS): UseQuotesResult {
  const [quotes, setQuotes] = useState<Quote[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const isFirstLoad = useRef(true)

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const data = await fetchQuotes()
        if (cancelled) return
        setQuotes(data)
        setError(null)
      } catch (err) {
        if (cancelled) return
        setError(err instanceof Error ? err.message : 'failed to load quotes')
      } finally {
        if (!cancelled) {
          isFirstLoad.current = false
          setLoading(false)
        }
      }
    }

    load()
    const id = setInterval(load, pollMs)
    return () => {
      cancelled = true
      clearInterval(id)
    }
  }, [pollMs])

  return { quotes, loading, error }
}
