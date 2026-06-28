import { useTranslation } from 'react-i18next'
import { useSources } from '../hooks/useQueries'
import type { Source } from '../types'

export default function SourcesPage() {
  const { t } = useTranslation()
  const { data, isLoading } = useSources()
  const sources = data?.data as Source[] | undefined

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">{t('sources.title')}</h1>
      {isLoading ? (
        <p>{t('loading')}</p>
      ) : (
        <div className="space-y-4">
          {sources?.map((s) => (
            <div key={s.name} className="bg-white rounded-lg shadow p-4">
              <h2 className="text-lg font-semibold">{s.name}</h2>
              <p className="text-sm text-gray-600 mt-1">
                <span className="font-medium">{t('sources.organism')}:</span> {s.organism}
              </p>
              <p className="text-sm text-gray-600">
                <span className="font-medium">{t('sources.licence')}:</span> {s.licence}
              </p>
              <p className="text-sm text-gray-600">
                <span className="font-medium">{t('sources.attribution')}:</span> {s.attribution}
              </p>
              {s.url && (
                <a href={s.url} target="_blank" rel="noopener noreferrer" className="text-blue-600 text-sm hover:underline mt-2 inline-block">
                  {s.url}
                </a>
              )}
            </div>
          ))}
          {!sources?.length && <p className="text-gray-500">No sources available.</p>}
        </div>
      )}
    </div>
  )
}
