import { useTranslation } from 'react-i18next'
import { useParams } from 'react-router-dom'
import { useReservoir, useReservoirReadings } from '../hooks/useQueries'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import type { ReservoirDetail, Reading } from '../types'

export default function ReservoirDetail() {
  const { t } = useTranslation()
  const { slug } = useParams<{ slug: string }>()
  const { data: resData, isLoading: resLoading } = useReservoir(slug ?? '')
  const { data: readData, isLoading: readLoading } = useReservoirReadings(slug ?? '')

  const reservoir = resData?.data as ReservoirDetail | undefined
  const readings = readData?.data as Reading[] | undefined

  const chartData = readings?.map((r) => ({
    date: r.observed_at,
    fill_pct: r.fill_pct,
    volume_hm3: r.volume_hm3,
  })) ?? []

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">
        {resLoading ? t('loading') : reservoir?.name ?? t('notFound')}
      </h1>

      {reservoir && (
        <div className="grid md:grid-cols-2 gap-4 mb-6">
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{t('reservoir.capacity')}</div>
            <div className="text-xl font-bold">
              {reservoir.capacity_hm3 != null ? `${reservoir.capacity_hm3.toLocaleString()} hm³` : '-'}
            </div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{t('reservoir.volume')}</div>
            <div className="text-xl font-bold">
              {reservoir.latest_volume_hm3 != null ? `${reservoir.latest_volume_hm3.toLocaleString()} hm³` : '-'}
            </div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{t('reservoir.fillPercent')}</div>
            <div className="text-xl font-bold">
              {reservoir.latest_fill_pct != null ? `${Math.round(reservoir.latest_fill_pct)}%` : '-'}
            </div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{t('reservoir.basin')}</div>
            <div className="text-xl font-bold">{reservoir.basin_name ?? '-'}</div>
          </div>
        </div>
      )}

      <h2 className="text-lg font-semibold mb-3">{t('reservoir.historicalChart')}</h2>
      {readLoading ? (
        <p>{t('loading')}</p>
      ) : chartData.length > 0 ? (
        <div className="bg-white rounded-lg shadow p-4" style={{ height: 400 }}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis yAxisId="left" />
              <YAxis yAxisId="right" orientation="right" />
              <Tooltip />
              <Legend />
              <Line yAxisId="left" type="monotone" dataKey="fill_pct" name={t('reservoir.fill_pct')} stroke="#2563eb" strokeWidth={2} dot={false} />
              <Line yAxisId="right" type="monotone" dataKey="volume_hm3" name={t('reservoir.volume_hm3')} stroke="#16a34a" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      ) : (
        <p className="text-gray-500">No historical data available.</p>
      )}
    </div>
  )
}
