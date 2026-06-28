import { useTranslation } from 'react-i18next'

export default function Settings() {
  const { t, i18n } = useTranslation()

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng)
  }

  return (
    <div className="animate-fade-in">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-[#0f172a]">{t('settings.title')}</h1>
        <p className="text-[#475569] text-sm mt-1">Personaliza tu experiencia</p>
      </div>
      <div className="gov-card p-5 max-w-md">
        <h2 className="text-lg font-semibold text-[#0f172a] mb-4">{t('settings.language')}</h2>
        <div className="flex gap-3">
          <button
            onClick={() => changeLanguage('es')}
            className={`gov-btn ${
              i18n.language === 'es'
                ? 'bg-[#003366] text-white hover:bg-[#004a74]'
                : 'bg-white text-[#475569] border border-[#e2e8f0] hover:bg-[#f8fafc]'
            }`}
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
            </svg>
            {t('settings.spanish')}
          </button>
          <button
            onClick={() => changeLanguage('en')}
            className={`gov-btn ${
              i18n.language === 'en'
                ? 'bg-[#003366] text-white hover:bg-[#004a74]'
                : 'bg-white text-[#475569] border border-[#e2e8f0] hover:bg-[#f8fafc]'
            }`}
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
            </svg>
            {t('settings.english')}
          </button>
        </div>
      </div>
    </div>
  )
}
