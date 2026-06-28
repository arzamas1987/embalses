import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useReservoirs } from '../hooks/useQueries'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import type { ReservoirSummary } from '../types'

const RemoveIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
  </svg>
)

const CompareIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
  </svg>
)

const colors = ['#003366', '#16a34a', '#dc2626', '#9333ea', '#ea580c']

// Mock comparison data — in production this would fetch real aligned time-series
const mockChartData = [
  { date: '2024-01', r0: 45, r1: 62, r2: 38, r3: 55, r4: 71 },
  { date: '2024-02', r0: 48, r1: 60, r2: 40, r3: 53, r4: 69 },
  { date: '2024-03', r0: 52, r1: 58, r2: 42, r3: 51, r4: 67 },
  { date: '2024-04', r0: 55, r1: 55, r2: 45, r3: 48, r4: 64 },
  { date: '2024-05', r0: 58, r1: 52, r2: 48, r3: 45, r4: 62 },
  { date: '2024-06', r0: 60, r1: 50, r2: 50, r3: 42, r4: 60 },
]

export default function Comparator() {
  const { t } = useTranslation()
  const { data } = useReservoirs(1, 100)
  const reservoirs = data?.data as ReservoirSummary[] | undefined

  const [selected, setSelected] = useState<ReservoirSummary[]>([])
  const [metric, setMetric] = useState<'fill_pct' | 'volume_hm3'>('fill_pct')

  const addReservoir = (r: ReservoirSummary) => {
    if (selected.length >= 5) return
    if (selected.find((s) => s.id === r.id)) return
    setSelected([...selected, r])
  }

  const removeReservoir = (id: number) => {
    setSelected(selected.filter((s) => s.id !== id))
  }

  const chartData = mockChartData.map(d => {
    const result: Record<string, number | string> = { date: d.date }
    selected.forEach((_, i) => {
      result[`r${i}`] = metric === 'fill_pct' ? d[`r${i}` as keyof typeof d] : Math.round((d[`r${i}` as keyof typeof d] as number) * 12.5)
    })
    return result
  })

  return (
    <div className="animate-fade-in">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-[#0f172a]">{t('comparator.title')}</h1>
        <p className="text-[#475569] text-sm mt-1">
          Selecciona hasta 5 embalses para comparar su evolución histórica
        </p>
      </div>

      <div className="grid lg:grid-cols-3 gap-6 mb-6">
        <div className="lg:col-span-2 gov-card p-5">
          <div className="section-title">
            <CompareIcon />
            {t('comparator.selectReservoirs')}
          </div>
          <p className="text-sm text-[#475569] mb-3">
            {t('comparator.maxReservoirs')}
          </p>

          <div className="relative mb-4">
            <select
              className="gov-select w-full border border-[#e2e8f0] rounded-lg px-4 py-2.5 text-sm bg-white focus:outline-none focus:ring-2 focus:ring-[#003366]/20 focus:border-[#003366]"
              onChange={(e) => {
                const r = reservoirs?.find((x) => x.id === Number(e.target.value))
                if (r) addReservoir(r)
                e.target.value = ''
              }}
              value=""
            >
              <option value="">Añadir embalse...</option>
              {reservoirs
                ?.filter((r) => !selected.find((s) => s.id === r.id))
                .map((r) => (
                  <option key={r.id} value={r.id}>
                    {r.name} ({r.basin_name})
                  </option>
                ))}
            </select>
          </div>

          {selected.length === 0 ? (
            <div className="text-center py-8 text-[#94a3b8] border-2 border-dashed border-[#e2e8f0] rounded-lg">
              <svg className="w-10 h-10 mx-auto mb-2 text-[#cbd5e1]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              <p className="text-sm font-medium">Añade embalses para comparar</p>
            </div>
          ) : (
            <div className="space-y-2">
              {selected.map((s, i) => (
                <div
                  key={s.id}
                  className="flex items-center justify-between p-3 rounded-lg border border-[#e2e8f0] bg-white hover:shadow-sm transition-shadow"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className="w-3 h-3 rounded-full flex-shrink-0"
                      style={{ backgroundColor: colors[i % colors.length] }}
                    />
                    <div>
                      <div className="font-medium text-[#0f172a] text-sm">{s.name}</div>
                      <div className="text-xs text-[#475569]">{s.basin_name}</div>
                    </div>
                  </div>
                  <button
                    onClick={() => removeReservoir(s.id)}
                    className="p-1.5 rounded-md text-[#475569] hover:text-red-600 hover:bg-red-50 transition-colors"
                    title={t('comparator.remove')}
                  >
                    <RemoveIcon />
                  </button>
                </div>
              ))}
            </div>
          )}

          <div className="mt-4 flex gap-2">
            <button
              onClick={() => setSelected([])}
              className="gov-btn gov-btn-outline text-sm"
              disabled={selected.length === 0}
            >
              Limpiar selección
            </button>
          </div>
        </div>

        <div className="gov-card p-5">
          <div className="section-title">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {t('comparator.metric')}
          </div>
          <div className="flex flex-col gap-2">
            <button
              onClick={() => setMetric('fill_pct')}
              className={`gov-btn text-sm justify-start ${
                metric === 'fill_pct'
                  ? 'bg-[#003366] text-white hover:bg-[#004a74]'
                  : 'bg-white text-[#475569] border border-[#e2e8f0] hover:bg-[#f8fafc]'
              }`}
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
                <path strokeLinecap="round" strokeLinejoin="round" d="M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z" />
              </svg>
              % Llenado
            </button>
            <button
              onClick={() => setMetric('volume_hm3')}
              className={`gov-btn text-sm justify-start ${
                metric === 'volume_hm3'
                  ? 'bg-[#003366] text-white hover:bg-[#004a74]'
                  : 'bg-white text-[#475569] border border-[#e2e8f0] hover:bg-[#f8fafc]'
              }`}
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0L12 2.69z" />
              </svg>
              Volumen (hm³)
            </button>
          </div>
        </div>
      </div>

      {selected.length > 0 && (
        <div className="gov-card-elevated p-5 sm:p-6">
          <div className="section-title">
            <CompareIcon />
            Comparación de {metric === 'fill_pct' ? 'porcentaje de llenado' : 'volumen'}
          </div>
          <div className="h-[400px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{ top: 10, right: 20, left: 10, bottom: 10 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                <XAxis
                  dataKey="date"
                  tick={{ fontSize: 12, fill: '#475569' }}
                  tickLine={false}
                  axisLine={{ stroke: '#e2e8f0' }}
                />
                <YAxis
                  tick={{ fontSize: 12, fill: '#475569' }}
                  tickLine={false}
                  axisLine={{ stroke: '#e2e8f0' }}
                  label={{
                    value: metric === 'fill_pct' ? '% Llenado' : 'Volumen (hm³)',
                    angle: -90,
                    position: 'insideLeft',
                    style: { fill: '#475569', fontSize: 12 }
                  }}
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
                {selected.map((s, i) => (
                  <Line
                    key={s.id}
                    type="monotone"
                    dataKey={`r${i}`}
                    name={s.name}
                    stroke={colors[i]}
                    strokeWidth={2.5}
                    dot={false}
                    activeDot={{ r: 4, fill: colors[i], stroke: '#fff', strokeWidth: 2 }}
                    connectNulls
                  />
                ))}
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}
    </div>
  )
}
