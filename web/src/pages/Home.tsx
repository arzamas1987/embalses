import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { useRankings, useDataQuality } from '../hooks/useQueries'
import type { RankingItem, DataQualityReport } from '../types'

export default function Home() {
  const { t } = useTranslation()
  const { data: fullestData, isLoading: fLoading } = useRankings('fullest', 5)
  const { data: emptiestData, isLoading: eLoading } = useRankings('emptiest', 5)
  const { data: qualityData, isLoading: qLoading } = useDataQuality()

  const fullest = fullestData?.data as RankingItem[] | undefined
  const emptiest = emptiestData?.data as RankingItem[] | undefined
  const quality = qualityData?.data as DataQualityReport | undefined

  const avgFill = quality?.latest_reading_date
    ? Math.round((quality.reservoirs_with_readings / Math.max(quality.total_reservoirs, 1)) * 100)
    : 0

  return (
    <div>
      <h1 className="text-2xl font-bold mb-2">{t('home.title')}</h1>
      <p className="text-gray-600 mb-6">{t('home.subtitle')}</p>

      {/* KPIs */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">{t('home.kpi.totalReservoirs')}</div>
          <div className="text-2xl font-bold">{qLoading ? '-' : quality?.total_reservoirs ?? '-'}</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">{t('home.kpi.avgFill')}</div>
          <div className="text-2xl font-bold">{avgFill}%</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">{t('home.kpi.latestUpdate')}</div>
          <div className="text-lg font-bold">{quality?.latest_reading_date ?? '-'}</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">{t('home.kpi.totalStored')}</div>
          <div className="text-2xl font-bold">{qLoading ? '-' : quality?.reservoirs_with_readings ?? '-'}</div>
        </div>
      </div>

      {/* Rankings */}
      <div className="grid md:grid-cols-2 gap-6">
        <div className="bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-3">{t('home.rankings.fullest')}</h2>
          {fLoading ? (
            <p className="text-gray-500">{t('loading')}</p>
          ) : (
            <ul className="space-y-2">
              {fullest?.map((r) => (
                <li key={r.reservoir_id} className="flex justify-between">
                  <span className="truncate">{r.rank}. {r.name}</span>
                  <span className="font-semibold text-blue-700">{Math.round(r.value)}%</span>
                </li>
              ))}
            </ul>
          )}
          <Link to="/embalses" className="text-blue-600 text-sm mt-3 inline-block hover:underline">
            {t('home.rankings.viewAll')} &rarr;
          </Link>
        </div>

        <div className="bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-3">{t('home.rankings.emptiest')}</h2>
          {eLoading ? (
            <p className="text-gray-500">{t('loading')}</p>
          ) : (
            <ul className="space-y-2">
              {emptiest?.map((r) => (
                <li key={r.reservoir_id} className="flex justify-between">
                  <span className="truncate">{r.rank}. {r.name}</span>
                  <span className="font-semibold text-red-600">{Math.round(r.value)}%</span>
                </li>
              ))}
            </ul>
          )}
          <Link to="/embalses" className="text-blue-600 text-sm mt-3 inline-block hover:underline">
            {t('home.rankings.viewAll')} &rarr;
          </Link>
        </div>
      </div>
    </div>
  )
}
