import { useEffect, useMemo, useRef, useState } from 'react'
import * as d3 from 'd3'
import type { Feature, Geometry } from 'geojson'
import { useWorldAtlas } from '../../hooks/useWorldAtlas'
import { NUMERIC_TO_ALPHA2 } from './countryCodes'
import { flagEmoji, formatRelativeTime, MARKET_STATE_LABEL } from '../../utils/format'
import type { Quote } from '../../types/quote'
import type { Symbol } from '../../types/symbol'
import './WorldMap.css'

interface WorldMapProps {
  quotes: Quote[]
  symbols: Symbol[]
}

interface TooltipState {
  x: number
  y: number
  alpha2: string
  countryName: string
  quote: Quote
  symbolName: string
}

interface Size {
  width: number
  height: number
}

const INITIAL_SCALE = 1.6

function initialTransform(size: Size): d3.ZoomTransform {
  return d3.zoomIdentity
    .translate((size.width * (1 - INITIAL_SCALE)) / 2, (size.height * (1 - INITIAL_SCALE)) / 2)
    .scale(INITIAL_SCALE)
}

export function WorldMap({ quotes, symbols }: WorldMapProps) {
  const { countries, loading, error } = useWorldAtlas()
  const containerRef = useRef<HTMLDivElement | null>(null)
  const svgRef = useRef<SVGSVGElement | null>(null)
  const [tooltip, setTooltip] = useState<TooltipState | null>(null)
  const [size, setSize] = useState<Size>({ width: 0, height: 0 })
  const [selectedCountry, setSelectedCountry] = useState<string | null>(null)

  const zoomTransformRef = useRef<d3.ZoomTransform | null>(null)
  const zoomBehaviorRef = useRef<d3.ZoomBehavior<SVGSVGElement, unknown> | null>(null)

  const quoteByCountryRef = useRef<Map<string, Quote>>(new Map())
  const symbolByTickerRef = useRef<Map<string, Symbol>>(new Map())

  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const observer = new ResizeObserver((entries) => {
      const { width, height } = entries[0].contentRect
      setSize({ width, height })
    })
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  const symbolByTicker = useMemo(() => {
    const m = new Map<string, Symbol>()
    for (const s of symbols) m.set(s.ticker, s)
    return m
  }, [symbols])

  // choropleth is driven by the index, matching the MVP's "index status"
  const quoteByCountry = useMemo(() => {
    const m = new Map<string, Quote>()
    for (const q of quotes) {
      const sym = symbolByTicker.get(q.symbol)
      if (sym && sym.assetClass === 'index') {
        m.set(sym.countryCode, q)
      }
    }
    return m
  }, [quotes, symbolByTicker])

  useEffect(() => {
    symbolByTickerRef.current = symbolByTicker
  }, [symbolByTicker])
  useEffect(() => {
    quoteByCountryRef.current = quoteByCountry
  }, [quoteByCountry])

  const colorScale = useMemo(
    () => d3.scaleDiverging<string>().domain([-2, 0, 2]).interpolator(d3.interpolateRdYlGn).clamp(true),
    [],
  )

  const fillFor = (alpha2: string | undefined) => {
    const quote = alpha2 ? quoteByCountryRef.current.get(alpha2) : undefined
    return quote ? colorScale(quote.changePercent) : 'var(--map-no-data)'
  }
  const opacityFor = (alpha2: string | undefined) => {
    const quote = alpha2 ? quoteByCountryRef.current.get(alpha2) : undefined
    if (!quote) return 1
    return quote.marketState === 'open' ? 1 : 0.6
  }

  // fill/opacity/selection are updated in place by the effect below so
  // existing DOM nodes can animate between states instead of popping in
  // fresh on every quote refresh.
  useEffect(() => {
    if (!countries || !svgRef.current || size.width === 0 || size.height === 0) return

    const svg = d3.select(svgRef.current)
    svg.selectAll('*').remove()

    const projection = d3.geoNaturalEarth1().fitSize([size.width, size.height], countries)
    const path = d3.geoPath(projection)

    const g = svg.append('g')

    g.selectAll('path')
      .data(countries.features)
      .join('path')
      .attr('class', 'country')
      .classed('country--selected', (d) => NUMERIC_TO_ALPHA2[String((d as Feature<Geometry>).id)] === selectedCountry)
      .attr('d', (d) => path(d as Feature<Geometry>))
      .attr('fill', (d) => fillFor(NUMERIC_TO_ALPHA2[String((d as Feature<Geometry>).id)]))
      .attr('opacity', (d) => opacityFor(NUMERIC_TO_ALPHA2[String((d as Feature<Geometry>).id)]))
      .on('mousemove', function (event: MouseEvent, d) {
        const feature = d as Feature<Geometry>
        const alpha2 = NUMERIC_TO_ALPHA2[String(feature.id)]
        const quote = alpha2 ? quoteByCountryRef.current.get(alpha2) : undefined
        if (!alpha2 || !quote) {
          setTooltip(null)
          return
        }
        const sym = symbolByTickerRef.current.get(quote.symbol)
        setTooltip({
          x: event.offsetX,
          y: event.offsetY,
          alpha2,
          countryName: (feature.properties as { name?: string } | undefined)?.name ?? alpha2,
          quote,
          symbolName: sym?.name ?? quote.symbol,
        })
      })
      .on('mouseleave', () => setTooltip(null))
      .on('click', function (event: MouseEvent, d) {
        event.stopPropagation()
        const alpha2 = NUMERIC_TO_ALPHA2[String((d as Feature<Geometry>).id)]
        if (!alpha2 || !quoteByCountryRef.current.has(alpha2)) return
        setSelectedCountry((prev) => (prev === alpha2 ? null : alpha2))
      })

    // Clicking open water clears the current selection. This can't be a
    // plain 'click' listener on the svg: d3-zoom's own mousedown/mouseup
    // handling swallows the native click event that would normally follow
    // a real (non-synthetic) mousedown+mouseup pair, so it never bubbles
    // here. Pointer events aren't part of d3-zoom's listener set, so they
    // arrive untouched — track the down/up distance ourselves to tell a
    // tap from the start of a drag.
    let pointerDownAt: [number, number] | null = null
    svg
      .on('pointerdown', (event: PointerEvent) => {
        pointerDownAt = [event.clientX, event.clientY]
      })
      .on('pointerup', (event: PointerEvent) => {
        const start = pointerDownAt
        pointerDownAt = null
        if (!start || event.target !== svgRef.current) return
        const dx = event.clientX - start[0]
        const dy = event.clientY - start[1]
        if (dx * dx + dy * dy > 16) return
        setSelectedCountry(null)
      })

    // Google-Maps-style navigation: drag to pan, wheel/pinch to zoom,
    // double-click to zoom in. Attached here (not a separate effect)
    // because the <g> this transforms is recreated whenever geometry
    // changes.
    const zoom = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([1, 12])
      // Keep the world on screen: panning stops at the map edges instead
      // of letting the user drag the world away into the void.
      .translateExtent([
        [0, 0],
        [size.width, size.height],
      ])
      .on('zoom', (event: d3.D3ZoomEvent<SVGSVGElement, unknown>) => {
        zoomTransformRef.current = event.transform
        g.attr('transform', event.transform.toString())
        setTooltip(null)
      })

    svg.call(zoom)
    zoomBehaviorRef.current = zoom

    // Start zoomed in a bit (centered), like a maps app — restore the
    // user's previous position instead if they had already navigated.
    svg.call(zoom.transform, zoomTransformRef.current ?? initialTransform(size))
    // selectedCountry deliberately omitted: this effect only needs to run
    // once per geometry change, and the initial class is set from
    // selectedCountry's value at that time via the classed() call above.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [countries, size])

  // Style pass: pushes fill/opacity/selection onto the existing path
  // elements with a real transition, so a changed % or a new selection
  // fades in instead of snapping.
  useEffect(() => {
    if (!svgRef.current) return
    const selection = d3.select(svgRef.current).selectAll<SVGPathElement, Feature<Geometry>>('path.country')
    if (selection.empty()) return

    selection.classed('country--selected', (d) => NUMERIC_TO_ALPHA2[String(d.id)] === selectedCountry)

    selection
      .transition()
      .duration(600)
      .ease(d3.easeCubicOut)
      .attr('fill', (d) => fillFor(NUMERIC_TO_ALPHA2[String(d.id)]))
      .attr('opacity', (d) => opacityFor(NUMERIC_TO_ALPHA2[String(d.id)]))
    // fillFor/opacityFor close over quoteByCountryRef (a ref), so they're
    // stable in effect; quoteByCountry itself is the real trigger.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [quoteByCountry, colorScale, selectedCountry])

  function handleResetView() {
    const svgEl = svgRef.current
    const zoom = zoomBehaviorRef.current
    if (!svgEl || !zoom || size.width === 0) return
    const svg = d3.select(svgEl)
    // zoom.transform() driven through a D3 transition doesn't reliably
    // dispatch the 'zoom' event on this element (leftover gesture state
    // from the most recent wheel/drag interaction seems to swallow it) —
    // so apply the transform through the same synchronous call used on
    // mount (proven reliable) and let a one-shot CSS transition on <g>
    // animate the resulting attribute change instead.
    const g = svg.select<SVGGElement>('g')
    g.style('transition', 'transform 650ms cubic-bezier(0.33, 1, 0.68, 1)')
    svg.call(zoom.transform, initialTransform(size))
    setTimeout(() => g.style('transition', null), 700)
  }

  return (
    <div ref={containerRef} className="absolute inset-0 overflow-hidden bg-[var(--map-ocean)]">
      {loading && (
        <p className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-sm text-[var(--panel-text)]">
          loading map…
        </p>
      )}
      {error && (
        <p className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-sm text-[var(--negative)]">
          error: {error}
        </p>
      )}
      <svg ref={svgRef} width={size.width} height={size.height} role="img" aria-label="World market status map" />

      <button
        type="button"
        onClick={handleResetView}
        className="absolute right-5 bottom-11 z-10 flex items-center gap-1.5 rounded-full border border-[var(--panel-border)] bg-[var(--panel-bg)] px-3 py-1.5 font-[var(--sans)] text-xs text-[var(--panel-text)] shadow-[0_8px_24px_rgba(0,0,0,0.25)] backdrop-blur-[14px] transition-transform duration-150 hover:scale-105 active:scale-95"
      >
        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M3 12a9 9 0 1 1 2.6 6.4M3 12V6m0 6h6" />
        </svg>
        Reset View
      </button>

      {tooltip && (
        <div
          className="map-tooltip pointer-events-none absolute z-10 min-w-[200px] rounded-xl border border-[var(--panel-border)] bg-[var(--panel-bg)] px-3.5 py-3 font-[var(--sans)] text-[var(--panel-text)] shadow-[0_12px_32px_rgba(0,0,0,0.35)] backdrop-blur-[14px]"
          style={{ left: tooltip.x + 16, top: tooltip.y + 16 }}
        >
          <div className="flex items-center gap-2">
            <span className="text-xl leading-none">{flagEmoji(tooltip.alpha2)}</span>
            <div className="min-w-0">
              <div className="truncate text-sm font-semibold text-[var(--panel-text-strong)]">{tooltip.countryName}</div>
              <div className="truncate text-[0.7rem] text-[var(--panel-text-muted)]">
                {tooltip.symbolName} · {tooltip.quote.symbol}
              </div>
            </div>
          </div>

          <div className="mt-2.5 flex items-baseline justify-between gap-3">
            <span className="font-[var(--mono)] text-base font-semibold text-[var(--panel-text-strong)]">
              {tooltip.quote.price.toLocaleString(undefined, { maximumFractionDigits: 2 })}{' '}
              <span className="text-xs font-normal text-[var(--panel-text-muted)]">{tooltip.quote.currency}</span>
            </span>
            <span
              className={`font-[var(--mono)] text-sm font-semibold ${tooltip.quote.changePercent >= 0 ? 'text-[var(--positive)]' : 'text-[var(--negative)]'}`}
            >
              {tooltip.quote.changePercent >= 0 ? '▲' : '▼'} {Math.abs(tooltip.quote.changePercent).toFixed(2)}%
            </span>
          </div>

          <div className="mt-2 flex items-center justify-between border-t border-t-[var(--panel-border)] pt-2 text-[0.68rem] text-[var(--panel-text-muted)]">
            <span className="inline-flex items-center gap-1.5">
              <span
                className={`inline-block h-1.5 w-1.5 rounded-full ${tooltip.quote.marketState === 'open' ? 'animate-pulse bg-[var(--positive)]' : 'bg-[var(--panel-text-muted)]'}`}
              />
              {MARKET_STATE_LABEL[tooltip.quote.marketState]}
            </span>
            <span>{formatRelativeTime(tooltip.quote.lastUpdated)}</span>
          </div>

          <div className="mt-1.5 text-center text-[0.62rem] tracking-wide text-[var(--panel-text-muted)]">
            click to pin
          </div>
        </div>
      )}
    </div>
  )
}
