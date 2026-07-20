import { useEffect, useMemo, useRef, useState } from 'react'
import * as d3 from 'd3'
import type { Feature, Geometry } from 'geojson'
import { useWorldAtlas } from '../../hooks/useWorldAtlas'
import { NUMERIC_TO_ALPHA2 } from './countryCodes'
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
  quote: Quote
  symbolName: string
}

interface Size {
  width: number
  height: number
}

export function WorldMap({ quotes, symbols }: WorldMapProps) {
  const { countries, loading, error } = useWorldAtlas()
  const containerRef = useRef<HTMLDivElement | null>(null)
  const svgRef = useRef<SVGSVGElement | null>(null)
  const [tooltip, setTooltip] = useState<TooltipState | null>(null)
  const [size, setSize] = useState<Size>({ width: 0, height: 0 })
  // Survives redraws (new quotes, resize) so the user's zoom/pan position
  // isn't reset every time the choropleth re-renders.
  const zoomTransformRef = useRef<d3.ZoomTransform | null>(null)

  // The map is full-bleed now (the whole viewport is its container), so
  // unlike a fixed viewBox it needs to know the container's actual pixel
  // size to fit the projection for whatever screen shape it's given.
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

  // ticker -> Symbol, for joining quotes with country/assetClass metadata
  // that Quote itself doesn't carry.
  const symbolByTicker = useMemo(() => {
    const m = new Map<string, Symbol>()
    for (const s of symbols) m.set(s.ticker, s)
    return m
  }, [symbols])

  // countryCode -> the quote for that country's headline index. Other
  // asset classes (ETFs, ADRs, forex) exist per country too, but the
  // choropleth is driven by the index, matching the MVP's "index status"
  // goal — the others feed the capital-flow layer later, not this map.
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

  const colorScale = useMemo(
    () => d3.scaleDiverging<string>().domain([-2, 0, 2]).interpolator(d3.interpolateRdYlGn).clamp(true),
    [],
  )

  useEffect(() => {
    if (!countries || !svgRef.current || size.width === 0 || size.height === 0) return

    const svg = d3.select(svgRef.current)
    svg.selectAll('*').remove()

    const projection = d3.geoNaturalEarth1().fitSize([size.width, size.height], countries)
    const path = d3.geoPath(projection)

    const lookup = (feature: Feature<Geometry>) => {
      const alpha2 = NUMERIC_TO_ALPHA2[String(feature.id)]
      const quote = alpha2 ? quoteByCountry.get(alpha2) : undefined
      return { alpha2, quote }
    }

    const g = svg.append('g')

    g.selectAll('path')
      .data(countries.features)
      .join('path')
      .attr('class', 'country')
      .attr('d', (d) => path(d as Feature<Geometry>))
      .attr('fill', (d) => {
        const { quote } = lookup(d as Feature<Geometry>)
        return quote ? colorScale(quote.changePercent) : 'var(--map-no-data)'
      })
      .attr('opacity', (d) => {
        const { quote } = lookup(d as Feature<Geometry>)
        if (!quote) return 1
        return quote.marketState === 'open' ? 1 : 0.6
      })
      .on('mousemove', (event: MouseEvent, d) => {
        const { alpha2, quote } = lookup(d as Feature<Geometry>)
        if (!alpha2 || !quote) {
          setTooltip(null)
          return
        }
        const sym = symbolByTicker.get(quote.symbol)
        setTooltip({
          x: event.offsetX,
          y: event.offsetY,
          quote,
          symbolName: sym?.name ?? quote.symbol,
        })
      })
      .on('mouseleave', () => setTooltip(null))

    // Google-Maps-style navigation: drag to pan, wheel/pinch to zoom,
    // double-click to zoom in. Attached here (not a separate effect)
    // because the <g> this transforms is recreated on every redraw.
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

    // Start zoomed in a bit (centered), like a maps app — restore the
    // user's previous position instead if they had already navigated.
    const initialScale = 1.6
    const initialTransform =
      zoomTransformRef.current ??
      d3.zoomIdentity
        .translate((size.width * (1 - initialScale)) / 2, (size.height * (1 - initialScale)) / 2)
        .scale(initialScale)
    svg.call(zoom.transform, initialTransform)
  }, [countries, quoteByCountry, symbolByTicker, colorScale, size])

  return (
    <div ref={containerRef} className="absolute inset-0 overflow-hidden bg-[var(--map-ocean)]">
      {loading && <p className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-[var(--panel-text)] font-[var(--sans)] text-sm">loading map…</p>}
      {error && <p className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-[var(--negative)] font-[var(--sans)] text-sm">error: {error}</p>}
      <svg
        ref={svgRef}
        width={size.width}
        height={size.height}
        role="img"
        aria-label="World market status map"
      />
      {tooltip && (
        <div className="map-tooltip" style={{ left: tooltip.x + 12, top: tooltip.y + 12 }}>
          <strong>{tooltip.symbolName}</strong> <span className="muted">({tooltip.quote.symbol})</span>
          <div>
            {tooltip.quote.price.toLocaleString(undefined, { maximumFractionDigits: 2 })}{' '}
            {tooltip.quote.currency}
          </div>
          <div className={tooltip.quote.changePercent >= 0 ? 'positive' : 'negative'}>
            {tooltip.quote.changePercent >= 0 ? '+' : ''}
            {tooltip.quote.changePercent.toFixed(2)}%
          </div>
          <div className="muted">{tooltip.quote.marketState}</div>
        </div>
      )}
    </div>
  )
}
