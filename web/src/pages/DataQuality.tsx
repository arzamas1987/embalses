import { useTranslation } from 'react-i18next'
import { useDataQuality } from '../hooks/useQueries'
import type { DataQualityReport } from '../types'

export default function DataQuality() {
  const { t } = useTranslation()
  const { data, isLoading } = useDataQuality()
  const report = data?.data as DataQualityReport | undefined

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">{t('dataQuality.title')}</h1>
      {isLoading ? (
        <p>{t('loading')}</p>
      ) : report ? (
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{t('dataQuality.totalReservoirs')}</div>
            <div className="text-2xl font-bold">{report.total_reservoirs}</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{t('dataQuality.withReadings')}</div>
            <div className="text-2xl font-bold">{report.reservoirs_with_readings}</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{t('dataQuality.latestReading')}</div>
            <div className="text-lg font-bold">{report.latest_reading_date ?? '-'}</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{t('dataQuality.oldestReading')}</div>
            <div className="text-lg font-bold">{report.oldest_reading_date ?? '-'}</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{t('dataQuality.provisional')}</div>
            <div className="text-2xl font-bold">{report.provisional_count}</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{t('dataQuality.official')}</div>
            <div className="text-2xl font-bold">{report.official_count}</div>
          </div>
        </div>
      ) : (
        <p className="text-gray-500">No data available.</p>
      )}
    </div>
  )
}
