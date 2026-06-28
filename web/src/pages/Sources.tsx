import { useTranslation } from 'react-i18next'
import { useSources } from '../hooks/useQueries'
import type { Source } from '../types'

const ExternalLinkIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
  </svg>
)

export default function SourcesPage() {
  const { t } = useTranslation()
  const { data, isLoading } = useSources()
  const sources = data?.data as Source[] | undefined

  return (
    <div className="animate-fade-in">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-[#0f172a]">{t('sources.title')}</h1>
        <p className="text-[#475569] text-sm mt-1">Fuentes oficiales de datos utilizadas en el sistema</p>
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
        <div className="space-y-4">
          {sources?.map((s) => (
            <div key={s.name} className="gov-card p-5">
              <h2 className="text-lg font-semibold text-[#0f172a]">{s.name}</h2>
              <div className="mt-3 space-y-2 text-sm">
                <p className="text-[#475569]">
                  <span className="font-semibold text-[#0f172a]">{t('sources.organism')}:</span> {s.organism}
                </p>
                <p className="text-[#475569]">
                  <span className="font-semibold text-[#0f172a]">{t('sources.licence')}:</span> {s.licence}
                </p>
                <p className="text-[#475569]">
                  <span className="font-semibold text-[#0f172a]">{t('sources.attribution')}:</span> {s.attribution}
                </p>
                {s.url && (
                  <a
                    href={s.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1.5 text-[#003366] hover:text-[#004a74] font-medium mt-1"
                  >
                    <ExternalLinkIcon />
                    {s.url}
                  </a>
                )}
              </div>
            </div>
          ))}
          {!sources?.length && (
            <div className="gov-card p-8 text-center text-[#94a3b8]">
              <p>No sources available.</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
