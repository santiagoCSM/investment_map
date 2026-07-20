import { useMemo } from 'react'
import { useNow } from '../../hooks/useNow'
import { useAnimatedNumber } from '../../hooks/useAnimatedNumber'
import { formatRelativeTime } from '../../utils/format'
import type { Quote } from '../../types/quote'
import type { Symbol } from '../../types/symbol'

interface GlobalSummaryBarProps {
  quotes: Quote[]
  symbols: Symbol[]
}

// A thin strip of derived stats — nothing here is new data, it's all
// computed from the same quotes the map already renders. Sits above the
// map as a real layout row (not a floating overlay) so it reads as the
// app's status line rather than another panel competing for attention.
export function GlobalSummaryBar({ quotes, symbols }: GlobalSummaryBarProps) {
  const now = useNow()

  const indexQuotes = useMemo(() => {
    const indexTickers = new Set(symbols.filter((s) => s.assetClass === 'index').map((s) => s.ticker))
    return quotes.filter((q) => indexTickers.has(q.symbol))
  }, [quotes, symbols])

  const stats = useMemo(() => {
    if (indexQuotes.length === 0) return null
    const open = indexQuotes.filter((q) => q.marketState === 'open').length
    const positive = indexQuotes.filter((q) => q.changePercent > 0).length
    const negative = indexQuotes.filter((q) => q.changePercent < 0).length
    const avg = indexQuotes.reduce((sum, q) => sum + q.changePercent, 0) / indexQuotes.length
    const lastUpdated = indexQuotes.reduce(
      (latest, q) => (new Date(q.lastUpdated) > new Date(latest) ? q.lastUpdated : latest),
      indexQuotes[0].lastUpdated,
    )
    return { total: indexQuotes.length, open, positive, negative, avg, lastUpdated }
  }, [indexQuotes])

  const animatedAvg = useAnimatedNumber(stats?.avg ?? 0, 800)

  return (
    <div className="relative z-30 flex h-11 flex-shrink-0 items-center gap-5 overflow-x-auto border-b border-[var(--panel-border)] bg-[var(--panel-bg)] px-4 font-[var(--sans)] backdrop-blur-[14px]">
      {stats ? (
        <>
          <Stat label="Markets Open" value={`${stats.open}/${stats.total}`} />
          <Stat label="Positive" value={String(stats.positive)} tone="positive" />
          <Stat label="Negative" value={String(stats.negative)} tone="negative" />
          <Stat
            label="Global Avg"
            value={`${animatedAvg >= 0 ? '+' : ''}${animatedAvg.toFixed(2)}%`}
            tone={animatedAvg >= 0 ? 'positive' : 'negative'}
          />
          <div className="ml-auto flex flex-shrink-0 items-center gap-1.5 whitespace-nowrap text-xs text-[var(--panel-text-muted)]">
            <span className="inline-block h-1.5 w-1.5 rounded-full bg-[var(--positive)]" />
            Updated {formatRelativeTime(stats.lastUpdated, now)}
          </div>
        </>
      ) : (
        <span className="text-xs text-[var(--panel-text-muted)]">Loading global markets…</span>
      )}
    </div>
  )
}

function Stat({ label, value, tone }: { label: string; value: string; tone?: 'positive' | 'negative' }) {
  const color =
    tone === 'positive'
      ? 'text-[var(--positive)]'
      : tone === 'negative'
        ? 'text-[var(--negative)]'
        : 'text-[var(--panel-text-strong)]'
  return (
    <div className="flex flex-shrink-0 items-baseline gap-1.5 whitespace-nowrap">
      <span className={`font-[var(--mono)] text-sm font-semibold transition-colors duration-500 ${color}`}>
        {value}
      </span>
      <span className="text-[0.65rem] tracking-wide text-[var(--panel-text-muted)] uppercase">{label}</span>
    </div>
  )
}
