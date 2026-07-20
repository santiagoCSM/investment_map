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
      className="data-panel fixed top-[60px] right-5 z-20 box-border max-h-[min(60vh,420px)] w-[220px] overflow-y-auto rounded-[12px] border border-[var(--panel-border)] bg-[var(--panel-bg)] p-[0.7rem_0.8rem_0.9rem] text-left font-[var(--sans)] text-[var(--panel-text)] shadow-[0_12px_40px_rgba(0,0,0,0.3)] backdrop-blur-[14px]"
      style={position ? { left: position.x, top: position.y, right: 'auto' } : undefined}
    >
      <header className="cursor-grab touch-none select-none pb-2.5 mb-0.5 border-b border-b-[var(--panel-border)]" onPointerDown={onHandlePointerDown}>
        <h1 className="text-[0.95rem] font-semibold tracking-tight text-[var(--panel-text-strong)] m-0">Mapa de Inversiones</h1>
        <p className="text-[0.68rem] text-[var(--panel-text-muted)] m-0 mt-0.5">Live global market status</p>
      </header>

      {loading && <p className="text-[var(--panel-text-muted)] text-sm">loading…</p>}
      {error && <p className="text-[var(--negative)] text-sm">{error}</p>}

      {!loading && !error && (
        <ul className="list-none mt-2 p-0 m-0 flex flex-col gap-px">
          {sorted.map((q) => (
            <li key={q.symbol} className="flex justify-between items-center gap-2 px-[0.35rem] py-[0.28rem] rounded-md text-xs hover:bg-[var(--panel-row-hover)]">
              <span className="text-[var(--panel-text)] whitespace-nowrap overflow-hidden text-ellipsis">{nameByTicker.get(q.symbol) ?? q.symbol}</span>
              <span className={`font-[var(--mono)] font-medium whitespace-nowrap flex-shrink-0 ${q.changePercent >= 0 ? 'text-[var(--positive)]' : 'text-[var(--negative)]'}`}>
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
