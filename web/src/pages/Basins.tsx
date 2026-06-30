import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { useBasinSummary } from '../hooks/useQueries'
import type { BasinSummary } from '../types'

const WaterIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0L12 2.69z" />
  </svg>
)

const ReservoirIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
  </svg>
)

function getFillColor(pct: number): string {
  if (pct < 20) return '#dc2626'
  if (pct < 40) return '#ea580c'
  if (pct < 60) return '#ca8a04'
  if (pct < 80) return '#16a34a'
  return '#15803d'
}

function getFillBadgeClass(pct: number): string {
  if (pct < 20) return 'gov-badge-red'
  if (pct < 40) return 'gov-badge-orange'
  if (pct < 60) return 'gov-badge-yellow'
  if (pct < 80) return 'gov-badge-green-light'
  return 'gov-badge-green-dark'
}

function FillBar({ pct }: { pct: number }) {
  const color = getFillColor(pct)
  return (
    <div className="gov-progress-bar w-full">
      <div
        className="gov-progress-fill"
        style={{ width: `${pct}%`, backgroundColor: color }}
      />
    </div>
  )
}

export default function Basins() {
  const { t } = useTranslation()
  const { data, isLoading } = useBasinSummary()
  const basins = data?.data as BasinSummary[] | undefined

  return (
    <div className="animate-fade-in">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-[#0f172a]">{t('nav.basins')}</h1>
        <p className="text-[#475569] text-sm mt-1">
          Estado actualizado por cuenca hidrográfica
        </p>
      </div>

      {isLoading ? (
        <div className="gov-card p-12 flex items-center justify-center text-[#94a3b8]">
          <svg className="animate-spin h-6 w-6 mr-3" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          {t('loading')}
        </div>
      ) : (
        <div className="grid md:grid-cols-2 xl:grid-cols-3 gap-5">
          {basins?.map((b) => {
            const pct = Math.round(b.avg_fill_pct)
            return (
              <Link
                key={b.id}
                to={`/cuencas/${encodeURIComponent(b.name)}`}
                className="gov-card p-5 hover:shadow-md transition-shadow group"
              >
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h2 className="font-bold text-[#0f172a] group-hover:text-[#003366] transition-colors">
                      {b.name}
                    </h2>
                    <div className="flex items-center gap-4 mt-2 text-xs text-[#475569]">
                      <span className="inline-flex items-center gap-1">
                        <ReservoirIcon />
                        {b.reservoir_count} embalses
                      </span>
                      <span className="inline-flex items-center gap-1">
                        <WaterIcon />
                        {b.total_volume_hm3.toLocaleString()} / {b.total_capacity_hm3.toLocaleString()} hm³
                      </span>
                    </div>
                  </div>
                  <span className={`gov-badge ${getFillBadgeClass(pct)}`}>
                    {pct}%
                  </span>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-[#475569]">Nivel medio de llenado</span>
                    <span className="font-semibold text-[#0f172a]">{pct}%</span>
                  </div>
                  <FillBar pct={pct} />
                </div>

                {b.latest_observed_at && (
                  <div className="mt-4 pt-3 border-t border-[#e2e8f0] text-xs text-[#94a3b8]">
                    Última lectura: {b.latest_observed_at}
                  </div>
                )}
              </Link>
            )
          })}
        </div>
      )}

      {!isLoading && !basins?.length && (
        <div className="gov-card p-8 text-center text-[#94a3b8]">
          <p>No hay datos disponibles.</p>
        </div>
      )}
    </div>
  )
}
