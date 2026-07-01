import { useTranslation } from 'react-i18next'
import { useAllReservoirs } from '../hooks/useQueries'
import ReservoirMap from '../components/ReservoirMap'
import type { ReservoirSummary } from '../types'

export default function MapPage() {
  const { t } = useTranslation()
  const { data, isLoading } = useAllReservoirs()
  const reservoirs = data?.data as ReservoirSummary[] | undefined

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
      <ReservoirMap reservoirs={reservoirs ?? []} height={500} />
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
