import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'

export default function NotFound() {
  const { t } = useTranslation()
  return (
    <div className="animate-fade-in text-center py-20">
      <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-[#f1f5f9] flex items-center justify-center text-[#cbd5e1]">
        <svg className="w-10 h-10" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h1 className="text-4xl font-bold mb-4 text-[#0f172a]">404</h1>
      <p className="text-[#475569] mb-6">{t('notFound')}</p>
      <Link to="/" className="gov-btn gov-btn-primary">
        Volver al inicio
      </Link>
    </div>
  )
}
