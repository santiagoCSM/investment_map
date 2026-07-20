import { useEffect, useState } from 'react'

// Ticks so relative-time labels ("Updated 2m ago") stay fresh between
// data polls instead of freezing at whatever they said on last fetch.
export function useNow(intervalMs = 30_000): Date {
  const [now, setNow] = useState(() => new Date())

  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs])

  return now
}
