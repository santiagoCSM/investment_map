import { useEffect, useRef, useState } from 'react'

// Eases a displayed number toward `target` over `durationMs` instead of
// snapping — the motion itself communicates "this just updated" for
// headline stats, without extra chrome around the number.
export function useAnimatedNumber(target: number, durationMs = 700): number {
  const [value, setValue] = useState(target)
  const valueRef = useRef(target)

  useEffect(() => {
    valueRef.current = value
  }, [value])

  useEffect(() => {
    const from = valueRef.current
    if (from === target) return

    const start = performance.now()
    let raf = 0

    function tick(now: number) {
      const t = Math.min(1, (now - start) / durationMs)
      const eased = 1 - (1 - t) ** 3
      const next = from + (target - from) * eased
      valueRef.current = next
      setValue(next)
      if (t < 1) raf = requestAnimationFrame(tick)
    }

    raf = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(raf)
  }, [target, durationMs])

  return value
}
