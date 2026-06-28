import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import { useReservoirs } from '../hooks/useQueries'
import type { ReservoirSummary } from '../types'

function getFillColor(pct: number): string {
  if (pct < 20) return '#dc2626'
  if (pct < 40) return '#ea580c'
  if (pct < 60) return '#ca8a04'
  if (pct < 80) return '#16a34a'
  return '#15803d'
}

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

    const markers = document.querySelectorAll('.maplibregl-marker')
    markers.forEach((m) => m.remove())

    reservoirs.forEach((r) => {
      const lng = -3 + (Math.random() - 0.5) * 10
      const lat = 40 + (Math.random() - 0.5) * 8
      const fillPct = r.latest_fill_pct ?? 0
      const color = getFillColor(fillPct)

      const el = document.createElement('div')
      el.className = 'w-3.5 h-3.5 rounded-full border-2 border-white cursor-pointer shadow-md'
      el.style.backgroundColor = color
      el.title = r.name

      new maplibregl.Marker(el)
        .setLngLat([lng, lat])
        .addTo(mapRef.current!)
        .getElement()
        .addEventListener('click', () => setSelected(r))
    })
  }, [reservoirs])

  return (
    <div className="animate-fade-in">
      <div className="mb-4">
        <h1 className="text-2xl font-bold text-[#0f172a]">{t('map.title')}</h1>
        <p className="text-[#475569] text-sm mt-1">Haz clic en un marcador para ver detalles</p>
      </div>
      {isLoading && (
        <div className="flex items-center gap-2 text-[#94a3b8] mb-4">
          <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          {t('loading')}
        </div>
      )}
      <div className="relative">
        <div
          ref={mapContainer}
          className="w-full rounded-xl border border-[#e2e8f0] shadow-card"
          style={{ height: 500 }}
        />
        {selected && (
          <div className="absolute bottom-4 left-4 gov-card p-4 max-w-xs">
            <h3 className="font-semibold text-[#0f172a]">{selected.name}</h3>
            <p className="text-sm text-[#475569]">{selected.basin_name}</p>
            <p className="text-sm text-[#475569]">{selected.province_name}</p>
            {selected.latest_fill_pct != null && (
              <div className="flex items-center gap-2 mt-2">
                <span
                  className="w-2.5 h-2.5 rounded-full"
                  style={{ backgroundColor: getFillColor(selected.latest_fill_pct) }}
                />
                <span className="text-sm font-semibold text-[#0f172a]">
                  {Math.round(selected.latest_fill_pct)}% lleno
                </span>
              </div>
            )}
            <button
              onClick={() => setSelected(null)}
              className="mt-3 text-xs text-[#475569] hover:text-[#003366] underline"
            >
              Cerrar
            </button>
          </div>
        )}
      </div>
      <div className="mt-4 flex flex-wrap gap-3 text-xs">
        <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-full bg-[#dc2626]" /> &lt;20%</span>
        <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-full bg-[#ea580c]" /> 20-40%</span>
        <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-full bg-[#ca8a04]" /> 40-60%</span>
        <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-full bg-[#16a34a]" /> 60-80%</span>
        <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-full bg-[#15803d]" /> &gt;80%</span>
      </div>
    </div>
  )
}
