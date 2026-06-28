import { useTranslation } from 'react-i18next'
import { useParams, Link } from 'react-router-dom'
import { useReservoir, useReservoirReadings } from '../hooks/useQueries'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import type { ReservoirDetail, Reading } from '../types'

const BackIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
  </svg>
)

const CapacityIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
  </svg>
)

const VolumeIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0L12 2.69z" />
  </svg>
)

const PercentIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
    <path strokeLinecap="round" strokeLinejoin="round" d="M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z" />
  </svg>
)

const BasinIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
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

  const fillPct = reservoir?.latest_fill_pct
  const fillColor = fillPct != null ? getFillColor(fillPct) : '#94a3b8'

  const detailCards = [
    {
      icon: <CapacityIcon />,
      label: t('reservoir.capacity'),
      value: reservoir?.capacity_hm3 != null ? `${reservoir.capacity_hm3.toLocaleString()} hm³` : '-',
      color: '#003366',
    },
    {
      icon: <VolumeIcon />,
      label: t('reservoir.volume'),
      value: reservoir?.latest_volume_hm3 != null ? `${reservoir.latest_volume_hm3.toLocaleString()} hm³` : '-',
      color: '#006aa3',
    },
    {
      icon: <PercentIcon />,
      label: t('reservoir.fillPercent'),
      value: fillPct != null ? `${Math.round(fillPct)}%` : '-',
      color: fillColor,
      badge: fillPct != null ? getFillBadgeClass(fillPct) : undefined,
    },
    {
      icon: <BasinIcon />,
      label: t('reservoir.basin'),
      value: reservoir?.basin_name ?? '-',
      color: '#004a74',
    },
  ]

  return (
    <div className="animate-fade-in">
      <div className="mb-6">
        <Link
          to="/embalses"
          className="inline-flex items-center gap-1.5 text-sm text-[#475569] hover:text-[#003366] transition-colors mb-3"
        >
          <BackIcon />
          Volver al listado
        </Link>
        <h1 className="text-2xl sm:text-3xl font-bold text-[#0f172a]">
          {resLoading ? t('loading') : reservoir?.name ?? t('notFound')}
        </h1>
        {reservoir && (
          <div className="flex items-center gap-2 mt-2">
            {reservoir.province_name && (
              <span className="gov-badge gov-badge-blue">
                {reservoir.province_name}
              </span>
            )}
            {fillPct != null && (
              <span className={`gov-badge ${getFillBadgeClass(fillPct)}`}>
                {Math.round(fillPct)}% llenado
              </span>
            )}
          </div>
        )}
      </div>

      {reservoir && (
        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          {detailCards.map((card, i) => (
            <div key={i} className="gov-card p-5">
              <div className="flex items-center justify-between mb-3">
                <div className="p-2 rounded-lg" style={{ backgroundColor: `${card.color}10`, color: card.color }}>
                  {card.icon}
                </div>
                {card.badge && <span className={`gov-badge ${card.badge}`}>{card.value}</span>}
              </div>
              <div className="text-2xl font-bold text-[#0f172a]">{card.value}</div>
              <div className="text-xs font-semibold uppercase tracking-wider text-[#475569] mt-1">{card.label}</div>
            </div>
          ))}
        </div>
      )}

      <div className="gov-card-elevated p-5 sm:p-6">
        <div className="section-title">
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
          </svg>
          {t('reservoir.historicalChart')}
        </div>
        {readLoading ? (
          <div className="flex items-center justify-center py-16 text-[#94a3b8]">
            <svg className="animate-spin h-6 w-6 mr-3" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            {t('loading')}
          </div>
        ) : chartData.length > 0 ? (
          <div className="h-[400px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{ top: 10, right: 20, left: 10, bottom: 10 }}>
                <defs>
                  <linearGradient id="fillGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#003366" stopOpacity={0.1} />
                    <stop offset="100%" stopColor="#003366" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                <XAxis
                  dataKey="date"
                  tick={{ fontSize: 12, fill: '#475569' }}
                  tickLine={false}
                  axisLine={{ stroke: '#e2e8f0' }}
                />
                <YAxis
                  yAxisId="left"
                  tick={{ fontSize: 12, fill: '#475569' }}
                  tickLine={false}
                  axisLine={{ stroke: '#e2e8f0' }}
                  domain={[0, 100]}
                  label={{ value: '% Llenado', angle: -90, position: 'insideLeft', style: { fill: '#475569', fontSize: 12 } }}
                />
                <YAxis
                  yAxisId="right"
                  orientation="right"
                  tick={{ fontSize: 12, fill: '#475569' }}
                  tickLine={false}
                  axisLine={{ stroke: '#e2e8f0' }}
                  label={{ value: 'Volumen (hm³)', angle: 90, position: 'insideRight', style: { fill: '#475569', fontSize: 12 } }}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#fff',
                    border: '1px solid #e2e8f0',
                    borderRadius: '8px',
                    boxShadow: '0 4px 12px rgba(0,0,0,0.08)',
                    fontSize: 13,
                  }}
                  labelStyle={{ color: '#0f172a', fontWeight: 600, marginBottom: 4 }}
                />
                <Legend
                  wrapperStyle={{ fontSize: 13, paddingTop: 16 }}
                  iconType="circle"
                  iconSize={8}
                />
                <Line
                  yAxisId="left"
                  type="monotone"
                  dataKey="fill_pct"
                  name={t('reservoir.fill_pct')}
                  stroke="#003366"
                  strokeWidth={2.5}
                  dot={false}
                  activeDot={{ r: 4, fill: '#003366', stroke: '#fff', strokeWidth: 2 }}
                />
                <Line
                  yAxisId="right"
                  type="monotone"
                  dataKey="volume_hm3"
                  name={t('reservoir.volume_hm3')}
                  stroke="#16a34a"
                  strokeWidth={2.5}
                  dot={false}
                  activeDot={{ r: 4, fill: '#16a34a', stroke: '#fff', strokeWidth: 2 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <div className="text-center py-16 text-[#94a3b8]">
            <svg className="w-12 h-12 mx-auto mb-3 text-[#cbd5e1]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
            <p className="font-medium">No hay datos históricos disponibles</p>
          </div>
        )}
      </div>

      {readings && readings.length > 0 && (
        <div className="mt-6 gov-card p-5">
          <h3 className="text-lg font-semibold text-[#0f172a] mb-4">Últimas lecturas</h3>
          <div className="overflow-x-auto">
            <table className="gov-table">
              <thead>
                <tr>
                  <th>Fecha</th>
                  <th>Volumen (hm³)</th>
                  <th>% Llenado</th>
                  <th>Variación semanal</th>
                </tr>
              </thead>
              <tbody>
                {readings.slice(0, 10).map((r, i) => (
                  <tr key={i}>
                    <td>{r.observed_at}</td>
                    <td className="font-medium">{r.volume_hm3.toLocaleString()}</td>
                    <td>
                      <span className={`gov-badge ${getFillBadgeClass(r.fill_pct)}`}>
                        {Math.round(r.fill_pct)}%
                      </span>
                    </td>
                    <td>
                      {r.weekly_variation_hm3 != null ? (
                        <span className={`font-medium ${r.weekly_variation_hm3 >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                          {r.weekly_variation_hm3 >= 0 ? '+' : ''}{r.weekly_variation_hm3.toLocaleString()} hm³
                        </span>
                      ) : (
                        '-'
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
