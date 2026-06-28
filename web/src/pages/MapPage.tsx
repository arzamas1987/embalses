import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import { useReservoirs } from '../hooks/useQueries'
import type { ReservoirSummary } from '../types'

export default function MapPage() {
  const { t } = useTranslation()
  const mapContainer = useRef<HTMLDivElement>(null)
  const mapRef = useRef<maplibregl.Map | null>(null)
  const [selected, setSelected] = useState<ReservoirSummary | null>(null)
  const { data, isLoading } = useReservoirs(1, 500)
  const reservoirs = data?.data as ReservoirSummary[] | undefined

  useEffect(() => {
    if (!mapContainer.current || mapRef.current) return

    const map = new maplibregl.Map({
      container: mapContainer.current,
      style: {
        version: 8,
        sources: {
          osm: {
            type: 'raster',
            tiles: ['https://a.tile.openstreetmap.org/{z}/{x}/{y}.png'],
            tileSize: 256,
            attribution: '&copy; OpenStreetMap contributors',
          },
        },
        layers: [
          {
            id: 'osm',
            type: 'raster',
            source: 'osm',
          },
        ],
      },
      center: [-3.7, 40.4],
      zoom: 6,
    })

    mapRef.current = map

    return () => {
      map.remove()
      mapRef.current = null
    }
  }, [])

  useEffect(() => {
    if (!mapRef.current || !reservoirs) return

    // Clear existing markers
    const markers = document.querySelectorAll('.maplibregl-marker')
    markers.forEach((m) => m.remove())

    reservoirs.forEach((r) => {
      // Use a simple mock coordinate since we don't have real coordinates yet
      // In production this would use actual lat/lng from the database
      const lng = -3 + (Math.random() - 0.5) * 10
      const lat = 40 + (Math.random() - 0.5) * 8

      const el = document.createElement('div')
      el.className = 'w-3 h-3 rounded-full bg-blue-600 border border-white cursor-pointer'
      el.title = r.name

      new maplibregl.Marker(el)
        .setLngLat([lng, lat])
        .addTo(mapRef.current!)
        .getElement()
        .addEventListener('click', () => setSelected(r))
    })
  }, [reservoirs])

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">{t('map.title')}</h1>
      {isLoading && <p className="mb-2">{t('loading')}</p>}
      <div className="relative">
        <div ref={mapContainer} className="w-full rounded-lg shadow" style={{ height: 500 }} />
        {selected && (
          <div className="absolute bottom-4 left-4 bg-white rounded-lg shadow p-3 max-w-xs">
            <h3 className="font-semibold">{selected.name}</h3>
            <p className="text-sm text-gray-600">{selected.basin_name}</p>
            <p className="text-sm text-gray-600">{selected.province_name}</p>
            {selected.latest_fill_pct != null && (
              <p className="text-sm font-semibold text-blue-700">
                {Math.round(selected.latest_fill_pct)}% lleno
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
