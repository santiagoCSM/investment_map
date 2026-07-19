import * as d3 from 'd3'
import './Legend.css'

// Mirrors the WorldMap's color scale exactly (domain [-2, 0, 2], same
// interpolator) so the legend never drifts out of sync with what's
// actually rendered — duplicated here rather than imported from WorldMap
// since the scale is cheap to construct and this keeps Legend usable
// standalone.
const colorScale = d3.scaleDiverging<string>().domain([-2, 0, 2]).interpolator(d3.interpolateRdYlGn).clamp(true)

const STOPS = [-2, -1, 0, 1, 2]

export function Legend() {
  return (
    <div className="legend">
      <div className="legend__gradient">
        {STOPS.map((v) => (
          <div key={v} className="legend__swatch" style={{ background: colorScale(v) }} />
        ))}
      </div>
      <div className="legend__labels">
        <span>-2%</span>
        <span>0%</span>
        <span>+2%</span>
      </div>
      <div className="legend__note">
        <span className="legend__dot" style={{ background: 'var(--map-no-data, #2a2f3a)' }} />
        no data tracked
        <span className="legend__opacity-sample" />
        market closed (dimmed)
      </div>
    </div>
  )
}
