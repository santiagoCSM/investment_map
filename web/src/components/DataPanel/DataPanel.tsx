import { useMemo } from 'react'
import { useDraggable } from '../../hooks/useDraggable'
import type { Quote } from '../../types/quote'
import type { Symbol } from '../../types/symbol'
import './DataPanel.css'

interface DataPanelProps {
  quotes: Quote[]
  symbols: Symbol[]
  loading: boolean
  error: string | null
}

// Floating, draggable watchlist (Plague Inc.-esque: the map is the
// product, this is just an overlay). Deliberately small and move-able by
// its header so it never has to sit somewhere it's blocking the map — the
// legend lives separately, fixed at the bottom of the screen, since it's
// not something you'd ever want to drag around.
export function DataPanel({ quotes, symbols, loading, error }: DataPanelProps) {
  const { panelRef, position, onHandlePointerDown } = useDraggable()

  const nameByTicker = useMemo(() => {
    const m = new Map<string, string>()
    for (const s of symbols) m.set(s.ticker, s.name)
    return m
  }, [symbols])

  // Winners on top, losers at the bottom — the ranking itself is
  // information, not just a list of numbers.
  const sorted = useMemo(() => [...quotes].sort((a, b) => b.changePercent - a.changePercent), [quotes])

  return (
    <aside
      ref={panelRef as React.RefObject<HTMLElement>}
      className="data-panel"
      style={position ? { left: position.x, top: position.y, right: 'auto' } : undefined}
    >
      <header className="data-panel__header" onPointerDown={onHandlePointerDown}>
        <h1>Mapa de Inversiones</h1>
        <p className="data-panel__subtitle">Live global market status</p>
      </header>

      {loading && <p className="status">loading…</p>}
      {error && <p className="status status--error">{error}</p>}

      {!loading && !error && (
        <ul className="quote-list">
          {sorted.map((q) => (
            <li key={q.symbol} className="quote-list__row">
              <span className="quote-list__name">{nameByTicker.get(q.symbol) ?? q.symbol}</span>
              <span className={`quote-list__change ${q.changePercent >= 0 ? 'positive' : 'negative'}`}>
                {q.changePercent >= 0 ? '+' : ''}
                {q.changePercent.toFixed(2)}%
              </span>
            </li>
          ))}
        </ul>
      )}
    </aside>
  )
}
