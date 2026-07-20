import * as d3 from 'd3'

// Mirrors the WorldMap's color scale exactly (domain [-2, 0, 2], same
// interpolator) so the legend never drifts out of sync with what's
// actually rendered — duplicated here rather than imported from WorldMap
// since the scale is cheap to construct and this keeps Legend usable
// standalone.
const colorScale = d3.scaleDiverging<string>().domain([-2, 0, 2]).interpolator(d3.interpolateRdYlGn).clamp(true)

const STOPS = [-2, -1, 0, 1, 2]

// Single compact row rather than stacked blocks — this is a footnote, not
// a legend card, so it shouldn't claim more vertical space than the map.
export function Legend() {
  return (
    <div className="absolute bottom-0 left-0 right-0 flex flex-row items-center justify-center flex-wrap gap-2 px-3 py-1.5 text-xs text-[var(--legend-text)] font-[var(--sans)] leading-none bg-[var(--legend-bg)] backdrop-blur-sm z-10">
      <span className="font-[var(--mono)] text-[var(--legend-text)]">-2%</span>
      <div className="flex w-[110px] h-[7px] rounded overflow-hidden">
        {STOPS.map((v) => (
          <div key={v} className="flex-1" style={{ background: colorScale(v) }} />
        ))}
      </div>
      <span className="font-[var(--mono)] text-[var(--legend-text)]">+2%</span>

      <span className="w-px h-2.5 bg-[var(--legend-sep)]" />

      <span className="inline-flex items-center gap-0.75 text-[var(--legend-text)]">
        <span className="inline-block w-[7px] h-[7px] rounded-sm" style={{ background: 'var(--map-no-data, #2a2f3a)' }} />
        no data
      </span>
      <span className="inline-flex items-center gap-0.75 text-[var(--legend-text)]">
        <span className="inline-block w-[7px] h-[7px] rounded-sm bg-[#4caf50] opacity-60" />
        closed
      </span>
    </div>
  )
}
