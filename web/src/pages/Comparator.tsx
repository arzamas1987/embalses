import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAllReservoirs, useComparatorData } from '../hooks/useQueries'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import TimeRangeSelector from '../components/TimeRangeSelector'
import { getRangeDates } from '../utils/date'
import type { TimeRange } from '../utils/date'
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

const SearchIcon = () => (
  <svg className="w-4 h-4 text-[#94a3b8]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
  </svg>
)

const colors = ['#003366', '#16a34a', '#dc2626', '#9333ea', '#ea580c']

export default function Comparator() {
  const { t } = useTranslation()
  const { data } = useAllReservoirs()
  const reservoirs = (data?.data as ReservoirSummary[] | undefined) ?? []

  const [selected, setSelected] = useState<ReservoirSummary[]>([])
  const [metric, setMetric] = useState<'fill_pct' | 'volume_hm3'>('fill_pct')
  const [range, setRange] = useState<TimeRange>('1y')
  const [search, setSearch] = useState('')

  const { since, until } = getRangeDates(range)
  const selectedSlugs = selected.map((s) => s.slug).filter(Boolean) as string[]
  const { data: compareData, isLoading: compareLoading } = useComparatorData(selectedSlugs, since, until)
  const series = compareData?.data as Record<string, { observed_at: string; fill_pct: number; volume_hm3: number }[]> | undefined

  const addReservoir = (r: ReservoirSummary) => {
    if (selected.length >= 5) return
    if (selected.find((s) => s.id === r.id)) return
    setSelected([...selected, r])
    setSearch('')
  }

  const removeReservoir = (id: number) => {
    setSelected(selected.filter((s) => s.id !== id))
  }

  const filteredReservoirs = useMemo(() => {
    const term = search.toLowerCase().trim()
    if (!term) return []
    return reservoirs
      .filter((r) => !selected.find((s) => s.id === r.id))
      .filter(
        (r) =>
          r.name.toLowerCase().includes(term) ||
          (r.basin_name?.toLowerCase() || '').includes(term) ||
          (r.province_name?.toLowerCase() || '').includes(term)
      )
      .slice(0, 50)
  }, [reservoirs, search, selected])

  const chartData = useMemo(() => {
    if (!series || selectedSlugs.length === 0) return []

    const dateSet = new Set<string>()
    selectedSlugs.forEach((slug) => {
      series[slug]?.forEach((p) => dateSet.add(p.observed_at))
    })
    const dates = Array.from(dateSet).sort()

    return dates.map((date) => {
      const row: Record<string, number | string> = { date }
      selectedSlugs.forEach((slug, i) => {
        const point = series[slug]?.find((p) => p.observed_at === date)
        if (point) {
          row[`r${i}`] = metric === 'fill_pct' ? point.fill_pct : point.volume_hm3
        } else {
          row[`r${i}`] = NaN
        }
      })
      return row
    })
  }, [series, selectedSlugs, metric])

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
            <div className="absolute left-3 top-1/2 -translate-y-1/2">
              <SearchIcon />
            </div>
            <input
              type="text"
              placeholder="Buscar embalse por nombre, cuenca o provincia..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9 pr-4 py-2.5 w-full border border-[#e2e8f0] rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-[#003366]/20 focus:border-[#003366] transition-all bg-white"
            />
            {search && filteredReservoirs.length > 0 && (
              <div className="absolute z-10 mt-1 w-full bg-white border border-[#e2e8f0] rounded-lg shadow-elevated max-h-60 overflow-auto">
                {filteredReservoirs.map((r) => (
                  <button
                    key={r.id}
                    onClick={() => addReservoir(r)}
                    className="w-full text-left px-4 py-2 text-sm hover:bg-[#f8fafc] border-b border-[#f1f5f9] last:border-0"
                  >
                    <div className="font-medium text-[#0f172a]">{r.name}</div>
                    <div className="text-xs text-[#475569]">{r.basin_name} {r.province_name ? `· ${r.province_name}` : ''}</div>
                  </button>
                ))}
              </div>
            )}
            {search && filteredReservoirs.length === 0 && (
              <div className="absolute z-10 mt-1 w-full bg-white border border-[#e2e8f0] rounded-lg shadow-elevated p-3 text-sm text-[#475569]">
                No se encontraron embalses
              </div>
            )}
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

        <div className="space-y-6">
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

          <div className="gov-card p-5">
            <div className="section-title mb-3">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              Periodo
            </div>
            <TimeRangeSelector value={range} onChange={setRange} />
          </div>
        </div>
      </div>

      {selected.length > 0 && (
        <div className="gov-card-elevated p-5 sm:p-6">
          <div className="section-title">
            <CompareIcon />
            Comparación de {metric === 'fill_pct' ? 'porcentaje de llenado' : 'volumen'}
          </div>
          {compareLoading ? (
            <div className="flex items-center justify-center py-16 text-[#94a3b8]">
              <svg className="animate-spin h-6 w-6 mr-3" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              Cargando datos...
            </div>
          ) : chartData.length > 0 ? (
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
                    domain={metric === 'fill_pct' ? [0, 100] : ['auto', 'auto']}
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
                      stroke={colors[i % colors.length]}
                      strokeWidth={2.5}
                      dot={false}
                      activeDot={{ r: 4, fill: colors[i], stroke: '#fff', strokeWidth: 2 }}
                      connectNulls
                    />
                  ))}
                </LineChart>
              </ResponsiveContainer>
            </div>
          ) : (
            <div className="text-center py-16 text-[#94a3b8]">
              <p className="font-medium">No hay datos para el periodo seleccionado</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
