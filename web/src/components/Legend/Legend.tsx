import * as d3 from 'd3'
import './Legend.css'

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
    <div className="legend">
      <span className="legend__scale-label">-2%</span>
      <div className="legend__gradient">
        {STOPS.map((v) => (
          <div key={v} className="legend__swatch" style={{ background: colorScale(v) }} />
        ))}
      </div>
      <span className="legend__scale-label">+2%</span>

      <span className="legend__sep" />

      <span className="legend__note">
        <span className="legend__dot" style={{ background: 'var(--map-no-data, #2a2f3a)' }} />
        no data
      </span>
      <span className="legend__note">
        <span className="legend__opacity-sample" />
        closed
      </span>
    </div>
  )
}
