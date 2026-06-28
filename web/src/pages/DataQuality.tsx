import { useTranslation } from 'react-i18next'
import { useDataQuality } from '../hooks/useQueries'
import type { DataQualityReport } from '../types'

const DatabaseIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
  </svg>
)

const CheckIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
  </svg>
)

const CalendarIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
  </svg>
)

const ClockIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
  </svg>
)

const WarningIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
  </svg>
)

const ShieldIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
  </svg>
)

export default function DataQuality() {
  const { t } = useTranslation()
  const { data, isLoading } = useDataQuality()
  const report = data?.data as DataQualityReport | undefined

  const cards = report ? [
    { icon: <DatabaseIcon />, label: t('dataQuality.totalReservoirs'), value: report.total_reservoirs, color: '#003366' },
    { icon: <CheckIcon />, label: t('dataQuality.withReadings'), value: report.reservoirs_with_readings, color: '#16a34a' },
    { icon: <CalendarIcon />, label: t('dataQuality.latestReading'), value: report.latest_reading_date ?? '-', color: '#006aa3', isSmall: true },
    { icon: <ClockIcon />, label: t('dataQuality.oldestReading'), value: report.oldest_reading_date ?? '-', color: '#004a74', isSmall: true },
    { icon: <WarningIcon />, label: t('dataQuality.provisional'), value: report.provisional_count, color: '#ca8a04' },
    { icon: <ShieldIcon />, label: t('dataQuality.official'), value: report.official_count, color: '#15803d' },
  ] : []

  return (
    <div className="animate-fade-in">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-[#0f172a]">{t('dataQuality.title')}</h1>
        <p className="text-[#475569] text-sm mt-1">Resumen del estado de los datos del sistema</p>
      </div>
      {isLoading ? (
        <div className="gov-card p-12 flex items-center justify-center text-[#94a3b8]">
          <svg className="animate-spin h-6 w-6 mr-3" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          {t('loading')}
        </div>
      ) : report ? (
        <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
          {cards.map((card, i) => (
            <div key={i} className="gov-card p-5">
              <div className="flex items-center justify-between mb-3">
                <div className="p-2 rounded-lg" style={{ backgroundColor: `${card.color}10`, color: card.color }}>
                  {card.icon}
                </div>
              </div>
              <div className={card.isSmall ? 'text-lg font-bold text-[#0f172a]' : 'kpi-value'}>
                {card.value}
              </div>
              <div className="kpi-label mt-1">{card.label}</div>
            </div>
          ))}
        </div>
      ) : (
        <div className="gov-card p-8 text-center text-[#94a3b8]">
          <p>No data available.</p>
        </div>
      )}
    </div>
  )
}
