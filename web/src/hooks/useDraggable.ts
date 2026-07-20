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
//
// `position` starts (and stays) null until the user's first drag — the
// panel's default on-screen spot is plain CSS (e.g. `top: 20px; right:
// 20px`), not something measured in JS on mount. An earlier version did
// measure it on mount via getBoundingClientRect(), and that one-time
// measurement raced with the browser's own initial layout pass often
// enough to land the panel off-screen. Reading the rect inside
// onHandlePointerDown instead is reliable because it only ever runs in
// response to a real user gesture, well after layout has settled.
export function useDraggable(): UseDraggableResult {
  const panelRef = useRef<HTMLElement | null>(null)
  const [position, setPosition] = useState<Position | null>(null)
  const dragOffset = useRef<{ dx: number; dy: number } | null>(null)

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
      const el = panelRef.current
      if (!el) return
      const current = position ?? el.getBoundingClientRect()
      if (!position) setPosition({ x: current.x, y: current.y })
      ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
      dragOffset.current = { dx: e.clientX - current.x, dy: e.clientY - current.y }
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
