import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { useReservoirs } from '../hooks/useQueries'
import type { ReservoirSummary } from '../types'

export default function Reservoirs() {
  const { t } = useTranslation()
  const { data, isLoading } = useReservoirs(1, 100)
  const reservoirs = data?.data as ReservoirSummary[] | undefined

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">{t('nav.reservoirs')}</h1>
      {isLoading ? (
        <p>{t('loading')}</p>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="bg-gray-100">
              <tr>
                <th className="text-left px-4 py-2">{t('reservoir.basin')}</th>
                <th className="text-left px-4 py-2">{t('reservoir.province')}</th>
                <th className="text-right px-4 py-2">{t('reservoir.fillPercent')}</th>
                <th className="text-right px-4 py-2">{t('reservoir.capacity')}</th>
              </tr>
            </thead>
            <tbody>
              {reservoirs?.map((r) => (
                <tr key={r.id} className="border-t hover:bg-gray-50">
                  <td className="px-4 py-2">
                    <Link to={`/embalses/${encodeURIComponent(r.external_id)}`} className="text-blue-600 hover:underline">
                      {r.name}
                    </Link>
                  </td>
                  <td className="px-4 py-2">{r.basin_name}</td>
                  <td className="px-4 py-2 text-right">{r.province_name}</td>
                  <td className="px-4 py-2 text-right font-semibold">
                    {r.latest_fill_pct != null ? `${Math.round(r.latest_fill_pct)}%` : '-'}
                  </td>
                  <td className="px-4 py-2 text-right">
                    {r.capacity_hm3 != null ? `${r.capacity_hm3.toLocaleString()} hm³` : '-'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!reservoirs?.length && <p className="p-4 text-gray-500">No reservoirs found.</p>}
        </div>
      )}
    </div>
  )
}
