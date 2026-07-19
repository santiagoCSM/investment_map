import { useEffect, useState } from 'react'
import { fetchSymbols } from '../api/client'
import type { Symbol } from '../types/symbol'

interface UseSymbolCatalogResult {
  symbols: Symbol[]
  loading: boolean
  error: string | null
}

// Reference data (ticker -> country/region/assetClass) rarely changes at
// runtime, so this fetches once on mount rather than polling like
// useQuotes does.
export function useSymbolCatalog(): UseSymbolCatalogResult {
  const [symbols, setSymbols] = useState<Symbol[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    fetchSymbols()
      .then((data) => {
        if (!cancelled) setSymbols(data)
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'failed to load symbol catalog')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [])

  return { symbols, loading, error }
}
