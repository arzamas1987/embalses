import { useTranslation } from 'react-i18next'

export default function Settings() {
  const { t, i18n } = useTranslation()

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng)
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">{t('settings.title')}</h1>
      <div className="bg-white rounded-lg shadow p-4 max-w-md">
        <h2 className="text-lg font-semibold mb-3">{t('settings.language')}</h2>
        <div className="flex gap-4">
          <button
            onClick={() => changeLanguage('es')}
            className={`px-4 py-2 rounded border ${
              i18n.language === 'es' ? 'bg-blue-600 text-white border-blue-600' : 'bg-white text-gray-700 border-gray-300'
            }`}
          >
            {t('settings.spanish')}
          </button>
          <button
            onClick={() => changeLanguage('en')}
            className={`px-4 py-2 rounded border ${
              i18n.language === 'en' ? 'bg-blue-600 text-white border-blue-600' : 'bg-white text-gray-700 border-gray-300'
            }`}
          >
            {t('settings.english')}
          </button>
        </div>
      </div>
    </div>
  )
}
