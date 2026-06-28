import { useTranslation } from 'react-i18next'
import { useRankings } from '../hooks/useQueries'
import type { RankingItem } from '../types'

function getFillColor(pct: number): string {
  if (pct < 20) return '#dc2626'
  if (pct < 40) return '#ea580c'
  if (pct < 60) return '#ca8a04'
  if (pct < 80) return '#16a34a'
  return '#15803d'
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

export default function Basins() {
  const { t } = useTranslation()
  const { data, isLoading } = useRankings('fullest', 20)
  const basins = data?.data as RankingItem[] | undefined

  return (
    <div className="animate-fade-in">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-[#0f172a]">{t('nav.basins')}</h1>
        <p className="text-[#475569] text-sm mt-1">Clasificación por nivel de llenado</p>
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
        <div className="gov-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="gov-table">
              <thead>
                <tr>
                  <th>#</th>
                  <th>{t('reservoir.basin')}</th>
                  <th>{t('reservoir.fillPercent')}</th>
                </tr>
              </thead>
              <tbody>
                {basins?.map((b) => {
                  const pct = Math.round(b.value)
                  return (
                    <tr key={b.reservoir_id}>
                      <td className="font-medium text-[#475569]">{b.rank}</td>
                      <td className="font-medium text-[#0f172a]">{b.name}</td>
                      <td>
                        <div className="flex items-center gap-3">
                          <FillBar pct={pct} />
                          <span className="font-semibold text-sm">{pct}%</span>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
          {!basins?.length && (
            <div className="p-8 text-center text-[#94a3b8]">
              <p>No hay datos disponibles.</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
