import { useTranslation } from 'react-i18next'
import { useRankings } from '../hooks/useQueries'
import type { RankingItem } from '../types'

export default function Basins() {
  const { t } = useTranslation()
  const { data, isLoading } = useRankings('fullest', 20)
  const basins = data?.data as RankingItem[] | undefined

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">{t('nav.basins')}</h1>
      {isLoading ? (
        <p>{t('loading')}</p>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="bg-gray-100">
              <tr>
                <th className="text-left px-4 py-2">#</th>
                <th className="text-left px-4 py-2">{t('reservoir.basin')}</th>
                <th className="text-right px-4 py-2">{t('reservoir.fillPercent')}</th>
              </tr>
            </thead>
            <tbody>
              {basins?.map((b) => (
                <tr key={b.reservoir_id} className="border-t">
                  <td className="px-4 py-2">{b.rank}</td>
                  <td className="px-4 py-2">{b.name}</td>
                  <td className="px-4 py-2 text-right font-semibold">{Math.round(b.value)}%</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
