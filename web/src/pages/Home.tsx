import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { useRankings, useDataQuality } from '../hooks/useQueries'
import type { RankingItem, DataQualityReport } from '../types'

const WaterIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0L12 2.69z" />
  </svg>
)

const ChartIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
  </svg>
)

const CalendarIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
  </svg>
)

const DatabaseIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
  </svg>
)

const ArrowUpIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
  </svg>
)

const ArrowDownIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
  </svg>
)

function getFillColor(pct: number): string {
  if (pct < 20) return '#dc2626'
  if (pct < 40) return '#ea580c'
  if (pct < 60) return '#ca8a04'
  if (pct < 80) return '#16a34a'
  return '#15803d'
}

function getFillBadgeClass(pct: number): string {
  if (pct < 20) return 'gov-badge-red'
  if (pct < 40) return 'gov-badge-orange'
  if (pct < 60) return 'gov-badge-yellow'
  if (pct < 80) return 'gov-badge-green-light'
  return 'gov-badge-green-dark'
}

function FillBar({ pct }: { pct: number }) {
  const color = getFillColor(pct)
  return (
    <div className="gov-progress-bar w-full max-w-[140px]">
      <div
        className="gov-progress-fill"
        style={{ width: `${pct}%`, backgroundColor: color }}
      />
    </div>
  )
}

export default function Home() {
  const { t } = useTranslation()
  const { data: fullestData, isLoading: fLoading } = useRankings('fullest', 5)
  const { data: emptiestData, isLoading: eLoading } = useRankings('emptiest', 5)
  const { data: qualityData, isLoading: qLoading } = useDataQuality()

  const fullest = fullestData?.data as RankingItem[] | undefined
  const emptiest = emptiestData?.data as RankingItem[] | undefined
  const quality = qualityData?.data as DataQualityReport | undefined

  const avgFill = quality?.latest_reading_date
    ? Math.round((quality.reservoirs_with_readings / Math.max(quality.total_reservoirs, 1)) * 100)
    : 0

  const kpiCards = [
    {
      icon: <DatabaseIcon />,
      label: t('home.kpi.totalReservoirs'),
      value: qLoading ? '-' : quality?.total_reservoirs ?? '-',
      color: '#003366',
    },
    {
      icon: <ChartIcon />,
      label: t('home.kpi.avgFill'),
      value: `${avgFill}%`,
      color: '#006aa3',
    },
    {
      icon: <CalendarIcon />,
      label: t('home.kpi.latestUpdate'),
      value: quality?.latest_reading_date ?? '-',
      color: '#004a74',
      isSmall: true,
    },
    {
      icon: <WaterIcon />,
      label: t('home.kpi.totalStored'),
      value: qLoading ? '-' : quality?.reservoirs_with_readings ?? '-',
      color: '#16a34a',
    },
  ]

  return (
    <div className="animate-fade-in">
      {/* Hero section */}
      <div className="hero-gradient rounded-2xl p-8 sm:p-10 mb-8 text-white shadow-elevated">
        <div className="max-w-3xl">
          <div className="flex items-center gap-2 mb-4">
            <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-white/15 text-white/90 text-xs font-medium border border-white/10">
              <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse" />
              Datos actualizados
            </span>
          </div>
          <h1 className="text-3xl sm:text-4xl font-bold mb-3 text-white leading-tight">
            {t('home.title')}
          </h1>
          <p className="text-white/80 text-lg leading-relaxed mb-6 max-w-2xl">
            {t('home.subtitle')}. Consulta el estado de los embalses españoles,
            compara niveles de llenado y accede a datos históricos.
          </p>
          <div className="flex flex-wrap gap-3">
            <Link to="/embalses" className="gov-btn gov-btn-primary bg-white text-[#003366] hover:bg-white/90">
              <WaterIcon />
              Ver embalses
            </Link>
            <Link to="/comparar" className="gov-btn gov-btn-outline border-white/30 text-white hover:bg-white/10 hover:text-white">
              <ChartIcon />
              Comparar
            </Link>
          </div>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {kpiCards.map((kpi, i) => (
          <div key={i} className="gov-card p-5 flex flex-col">
            <div className="flex items-center justify-between mb-3">
              <div className="p-2 rounded-lg" style={{ backgroundColor: `${kpi.color}10`, color: kpi.color }}>
                {kpi.icon}
              </div>
            </div>
            <div className={kpi.isSmall ? 'text-lg font-bold text-[#0f172a]' : 'kpi-value'}>
              {kpi.value}
            </div>
            <div className="kpi-label mt-1">{kpi.label}</div>
          </div>
        ))}
      </div>

      {/* Rankings */}
      <div className="grid lg:grid-cols-2 gap-6">
        <div className="gov-card p-5">
          <div className="section-title">
            <div className="p-1.5 rounded-md bg-green-50 text-green-600">
              <ArrowUpIcon />
            </div>
            {t('home.rankings.fullest')}
          </div>
          {fLoading ? (
            <div className="flex items-center justify-center py-8 text-[#94a3b8]">
              <svg className="animate-spin h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              {t('loading')}
            </div>
          ) : (
            <div className="space-y-3">
              {fullest?.map((r) => {
                const pct = Math.round(r.value)
                return (
                  <div key={r.reservoir_id} className="flex items-center gap-4 p-3 rounded-lg hover:bg-[#f8fafc] transition-colors">
                    <div className="flex-shrink-0 w-8 h-8 rounded-full bg-[#f1f5f9] flex items-center justify-center text-sm font-bold text-[#475569]">
                      {r.rank}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-[#0f172a] text-sm truncate mb-1">
                        <Link to={`/embalses/${encodeURIComponent(r.name.toLowerCase().replace(/\s+/g, '-'))}`} className="hover:text-[#003366]">
                          {r.name}
                        </Link>
                      </div>
                      <FillBar pct={pct} />
                    </div>
                    <div className="flex-shrink-0 flex flex-col items-end">
                      <span className={`gov-badge ${getFillBadgeClass(pct)}`}>
                        {pct}%
                      </span>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
          <Link to="/embalses" className="inline-flex items-center gap-1 mt-4 text-sm font-medium text-[#003366] hover:text-[#004a74] transition-colors">
            {t('home.rankings.viewAll')}
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </Link>
        </div>

        <div className="gov-card p-5">
          <div className="section-title">
            <div className="p-1.5 rounded-md bg-red-50 text-red-600">
              <ArrowDownIcon />
            </div>
            {t('home.rankings.emptiest')}
          </div>
          {eLoading ? (
            <div className="flex items-center justify-center py-8 text-[#94a3b8]">
              <svg className="animate-spin h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              {t('loading')}
            </div>
          ) : (
            <div className="space-y-3">
              {emptiest?.map((r) => {
                const pct = Math.round(r.value)
                return (
                  <div key={r.reservoir_id} className="flex items-center gap-4 p-3 rounded-lg hover:bg-[#f8fafc] transition-colors">
                    <div className="flex-shrink-0 w-8 h-8 rounded-full bg-[#f1f5f9] flex items-center justify-center text-sm font-bold text-[#475569]">
                      {r.rank}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-[#0f172a] text-sm truncate mb-1">
                        <Link to={`/embalses/${encodeURIComponent(r.name.toLowerCase().replace(/\s+/g, '-'))}`} className="hover:text-[#003366]">
                          {r.name}
                        </Link>
                      </div>
                      <FillBar pct={pct} />
                    </div>
                    <div className="flex-shrink-0 flex flex-col items-end">
                      <span className={`gov-badge ${getFillBadgeClass(pct)}`}>
                        {pct}%
                      </span>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
          <Link to="/embalses" className="inline-flex items-center gap-1 mt-4 text-sm font-medium text-[#003366] hover:text-[#004a74] transition-colors">
            {t('home.rankings.viewAll')}
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </Link>
        </div>
      </div>
    </div>
  )
}
