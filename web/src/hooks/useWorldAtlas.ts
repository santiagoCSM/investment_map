import { useEffect, useState } from 'react'
import { feature } from 'topojson-client'
import type { Topology, GeometryObject } from 'topojson-specification'
import type { FeatureCollection, Geometry } from 'geojson'

const ATLAS_URL = '/world-countries-50m.json'

interface UseWorldAtlasResult {
  countries: FeatureCollection<Geometry> | null
  loading: boolean
  error: string | null
}

// Loads the world topology once and converts it from TopoJSON (compact,
// shared-boundary format) to GeoJSON features D3's path generator can
// render directly. The file itself lives in /public — see countryCodes.ts
// for how each feature's numeric `id` maps back to the backend's country
// codes.
export function useWorldAtlas(): UseWorldAtlasResult {
  const [countries, setCountries] = useState<FeatureCollection<Geometry> | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    fetch(ATLAS_URL)
      .then((res) => {
        if (!res.ok) throw new Error(`failed to load world atlas: ${res.status}`)
        return res.json()
      })
      .then((topology: Topology) => {
        if (cancelled) return
        const object = topology.objects.countries as GeometryObject
        const geojson = feature(topology, object) as unknown as FeatureCollection<Geometry>
        setCountries(geojson)
      })
      .catch((err) => {
        if (cancelled) return
        setError(err instanceof Error ? err.message : 'failed to load world atlas')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [])

  return { countries, loading, error }
}
