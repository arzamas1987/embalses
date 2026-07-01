import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import type { ReservoirSummary } from '../types'

function getFillColor(pct: number): string {
  if (pct < 20) return '#dc2626'
  if (pct < 40) return '#ea580c'
  if (pct < 60) return '#ca8a04'
  if (pct < 80) return '#16a34a'
  return '#15803d'
}

interface ReservoirMapProps {
  reservoirs: ReservoirSummary[]
  height?: number
  center?: [number, number]
  zoom?: number
  fitToBounds?: boolean
}

export default function ReservoirMap({
  reservoirs,
  height = 360,
  center = [-3.7, 40.4],
  zoom = 6,
  fitToBounds = false,
}: ReservoirMapProps) {
  const mapContainer = useRef<HTMLDivElement>(null)
  const mapRef = useRef<maplibregl.Map | null>(null)
  const popupRef = useRef<maplibregl.Popup | null>(null)

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
      center,
      zoom,
    })

    mapRef.current = map

    return () => {
      map.remove()
      mapRef.current = null
    }
  }, [center, zoom])

  useEffect(() => {
    const map = mapRef.current
    if (!map) return

    const markers = document.querySelectorAll('.maplibregl-marker')
    markers.forEach((m) => m.remove())
    if (popupRef.current) {
      popupRef.current.remove()
      popupRef.current = null
    }

    const valid = reservoirs.filter((r) => {
      const lng = r.longitude ?? 0
      const lat = r.latitude ?? 0
      return !(lng === 0 && lat === 0)
    })

    if (valid.length === 0) return

    valid.forEach((r) => {
      const lng = r.longitude ?? 0
      const lat = r.latitude ?? 0
      const fillPct = r.latest_fill_pct ?? 0
      const color = getFillColor(fillPct)

      const el = document.createElement('div')
      el.className = 'w-3.5 h-3.5 rounded-full border-2 border-white cursor-pointer shadow-md'
      el.style.backgroundColor = color
      el.title = r.name

      const marker = new maplibregl.Marker(el).setLngLat([lng, lat]).addTo(map)

      marker.getElement().addEventListener('click', (e) => {
        e.stopPropagation()
        if (popupRef.current) popupRef.current.remove()

        const popup = new maplibregl.Popup({ offset: 8 })
          .setLngLat([lng, lat])
          .setHTML(`
            <div class="p-1 min-w-[160px]">
              <div class="font-semibold text-[#0f172a] text-sm">${r.name}</div>
              <div class="text-xs text-[#475569]">${r.basin_name ?? ''}</div>
              <div class="flex items-center gap-2 mt-2">
                <span class="w-2.5 h-2.5 rounded-full" style="background-color:${color}"></span>
                <span class="text-sm font-semibold text-[#0f172a]">${Math.round(fillPct)}% lleno</span>
              </div>
              ${r.slug ? `<a href="/embalses/${encodeURIComponent(r.slug)}" class="text-xs font-medium text-[#003366] hover:underline mt-2 inline-block">Ver detalle →</a>` : ''}
            </div>
          `)
          .addTo(map)

        popupRef.current = popup
      })
    })

    if (fitToBounds && valid.length > 0) {
      const bounds = new maplibregl.LngLatBounds()
      valid.forEach((r) => bounds.extend([r.longitude ?? 0, r.latitude ?? 0]))
      map.fitBounds(bounds, { padding: 40, maxZoom: 10 })
    }
  }, [reservoirs, fitToBounds])

  return (
    <div
      ref={mapContainer}
      className="w-full rounded-xl border border-[#e2e8f0] shadow-card"
      style={{ height }}
    />
  )
}
