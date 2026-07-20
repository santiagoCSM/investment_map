import { useCallback, useEffect, useRef, useState, type PointerEvent as ReactPointerEvent } from 'react'

interface Position {
  x: number
  y: number
}

interface UseDraggableResult {
  panelRef: React.RefObject<HTMLElement | null>
  position: Position | null
  onHandlePointerDown: (e: ReactPointerEvent) => void
}

// Drag-by-handle for a fixed-position panel: attach `onHandlePointerDown`
// to whatever element should act as the drag grip (a header bar, not the
// whole panel, so text inside stays selectable/clickable). Uses pointer
// capture rather than document-level listeners, so the drag keeps tracking
// even if the cursor moves faster than the panel.
export function useDraggable(initialCorner: 'top-right' | 'top-left' = 'top-right'): UseDraggableResult {
  const panelRef = useRef<HTMLElement | null>(null)
  const [position, setPosition] = useState<Position | null>(null)
  const dragOffset = useRef<{ dx: number; dy: number } | null>(null)

  // Position starts null and gets set on mount once we can measure the
  // panel — avoids hardcoding a width that has to stay in sync with CSS.
  useEffect(() => {
    const el = panelRef.current
    if (!el) return
    const rect = el.getBoundingClientRect()
    const margin = 20
    const x = initialCorner === 'top-right' ? window.innerWidth - rect.width - margin : margin
    setPosition({ x, y: margin })
  }, [initialCorner])

  const clamp = useCallback((x: number, y: number) => {
    const el = panelRef.current
    const width = el?.offsetWidth ?? 0
    const height = el?.offsetHeight ?? 0
    return {
      x: Math.min(Math.max(0, x), window.innerWidth - width),
      y: Math.min(Math.max(0, y), window.innerHeight - height),
    }
  }, [])

  const onHandlePointerDown = useCallback(
    (e: ReactPointerEvent) => {
      if (!position) return
      ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
      dragOffset.current = { dx: e.clientX - position.x, dy: e.clientY - position.y }
    },
    [position],
  )

  useEffect(() => {
    function onPointerMove(e: PointerEvent) {
      if (!dragOffset.current) return
      const next = clamp(e.clientX - dragOffset.current.dx, e.clientY - dragOffset.current.dy)
      setPosition(next)
    }
    function onPointerUp() {
      dragOffset.current = null
    }
    window.addEventListener('pointermove', onPointerMove)
    window.addEventListener('pointerup', onPointerUp)
    return () => {
      window.removeEventListener('pointermove', onPointerMove)
      window.removeEventListener('pointerup', onPointerUp)
    }
  }, [clamp])

  return { panelRef, position, onHandlePointerDown }
}
