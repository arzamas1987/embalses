import { useTranslation } from 'react-i18next'
import { useParams, Link } from 'react-router-dom'
import { useBasinDetail } from '../hooks/useQueries'
import ReservoirMap from '../components/ReservoirMap'
import type { ReservoirSummary } from '../types'

const BackIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
  </svg>
)

const CapacityIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
  </svg>
)

const VolumeIcon = () => (
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
    <div className="gov-progress-bar w-24">
      <div
        className="gov-progress-fill"
        style={{ width: `${pct}%`, backgroundColor: color }}
      />
    </div>
  )
}

export default function BasinDetail() {
  const { t } = useTranslation()
  const { slug } = useParams<{ slug: string }>()
  const { data, isLoading } = useBasinDetail(slug ?? '')
  const basin = data?.data as { id: number; name: string; code?: string; reservoir_count: number; total_capacity_hm3: number; total_volume_hm3: number; avg_fill_pct: number; latest_observed_at?: string; reservoirs: ReservoirSummary[] } | undefined

  const pct = basin ? Math.round(basin.avg_fill_pct) : 0

  return (
    <div className="animate-fade-in">
      <div className="mb-6">
        <Link
          to="/cuencas"
          className="inline-flex items-center gap-1.5 text-sm text-[#475569] hover:text-[#003366] transition-colors mb-3"
        >
          <BackIcon />
          Volver a cuencas
        </Link>
        <h1 className="text-2xl sm:text-3xl font-bold text-[#0f172a]">
          {isLoading ? t('loading') : basin?.name ?? 'Cuenca no encontrada'}
        </h1>
        {basin && (
          <div className="flex items-center gap-2 mt-2">
            <span className="gov-badge gov-badge-blue">
              {basin.reservoir_count} embalses
            </span>
            <span className={`gov-badge ${getFillBadgeClass(pct)}`}>
              {pct}% llenado medio
            </span>
          </div>
        )}
      </div>

      {basin && (
        <>
          <div className="mb-6">
            <h2 className="text-lg font-bold text-[#0f172a] mb-3">Mapa de embalses</h2>
            <ReservoirMap reservoirs={basin.reservoirs} height={320} fitToBounds />
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <div className="gov-card p-5">
              <div className="flex items-center justify-between mb-3">
                <div className="p-2 rounded-lg" style={{ backgroundColor: '#00336610', color: '#003366' }}>
                  <CapacityIcon />
                </div>
              </div>
              <div className="text-2xl font-bold text-[#0f172a]">{basin.total_capacity_hm3.toLocaleString()} hm³</div>
              <div className="text-xs font-semibold uppercase tracking-wider text-[#475569] mt-1">Capacidad total</div>
            </div>

            <div className="gov-card p-5">
              <div className="flex items-center justify-between mb-3">
                <div className="p-2 rounded-lg" style={{ backgroundColor: '#006aa310', color: '#006aa3' }}>
                  <VolumeIcon />
                </div>
              </div>
              <div className="text-2xl font-bold text-[#0f172a]">{basin.total_volume_hm3.toLocaleString()} hm³</div>
              <div className="text-xs font-semibold uppercase tracking-wider text-[#475569] mt-1">Volumen actual</div>
            </div>

            <div className="gov-card p-5">
              <div className="flex items-center justify-between mb-3">
                <div className="p-2 rounded-lg" style={{ backgroundColor: `${getFillColor(pct)}10`, color: getFillColor(pct) }}>
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
                    <path strokeLinecap="round" strokeLinejoin="round" d="M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z" />
                  </svg>
                </div>
              </div>
              <div className="text-2xl font-bold text-[#0f172a]">{pct}%</div>
              <div className="text-xs font-semibold uppercase tracking-wider text-[#475569] mt-1">Llenado medio</div>
            </div>

            <div className="gov-card p-5">
              <div className="flex items-center justify-between mb-3">
                <div className="p-2 rounded-lg" style={{ backgroundColor: '#004a7410', color: '#004a74' }}>
                  <ReservoirIcon />
                </div>
              </div>
              <div className="text-2xl font-bold text-[#0f172a]">{basin.reservoir_count}</div>
              <div className="text-xs font-semibold uppercase tracking-wider text-[#475569] mt-1">Embalses</div>
            </div>
          </div>

          <div className="gov-card overflow-hidden">
            <div className="p-5 border-b border-[#e2e8f0]">
              <h2 className="text-lg font-bold text-[#0f172a]">Embalses de la cuenca</h2>
              {basin.latest_observed_at && (
                <p className="text-sm text-[#475569] mt-1">Última lectura: {basin.latest_observed_at}</p>
              )}
            </div>
            <div className="overflow-x-auto">
              <table className="gov-table">
                <thead>
                  <tr>
                    <th>Embalse</th>
                    <th>Capacidad</th>
                    <th>% Llenado</th>
                  </tr>
                </thead>
                <tbody>
                  {basin.reservoirs.map((r) => {
                    const rpct = r.latest_fill_pct != null ? Math.round(r.latest_fill_pct) : null
                    return (
                      <tr key={r.id}>
                        <td>
                          {r.slug ? (
                            <Link
                              to={`/embalses/${encodeURIComponent(r.slug)}`}
                              className="font-medium text-[#003366] hover:text-[#004a74] hover:underline"
                            >
                              {r.name}
                            </Link>
                          ) : (
                            <span className="font-medium text-[#0f172a]">{r.name}</span>
                          )}
                        </td>
                        <td className="text-right font-medium">
                          {r.capacity_hm3 != null ? `${r.capacity_hm3.toLocaleString()} hm³` : '-'}
                        </td>
                        <td>
                          {rpct != null ? (
                            <div className="flex items-center gap-3">
                              <FillBar pct={rpct} />
                              <span className={`gov-badge ${getFillBadgeClass(rpct)}`}>{rpct}%</span>
                            </div>
                          ) : (
                            '-'
                          )}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
            {!basin.reservoirs.length && (
              <div className="p-8 text-center text-[#94a3b8]">
                <p>No hay embalses registrados en esta cuenca.</p>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
