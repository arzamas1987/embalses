import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useReservoirs } from '../hooks/useQueries'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import type { ReservoirSummary } from '../types'

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

  // Mock comparison data — in production this would fetch real aligned time-series
  const chartData = [
    { date: '2024-01', r1: 45, r2: 62, r3: 38 },
    { date: '2024-02', r1: 48, r2: 60, r3: 40 },
    { date: '2024-03', r1: 52, r2: 58, r3: 42 },
    { date: '2024-04', r1: 55, r2: 55, r3: 45 },
    { date: '2024-05', r1: 58, r2: 52, r3: 48 },
    { date: '2024-06', r1: 60, r2: 50, r3: 50 },
  ]

  const colors = ['#2563eb', '#16a34a', '#dc2626', '#9333ea', '#ea580c']

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">{t('comparator.title')}</h1>

      <div className="grid md:grid-cols-2 gap-6 mb-6">
        <div className="bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-2">{t('comparator.selectReservoirs')}</h2>
          <p className="text-sm text-gray-500 mb-3">{t('comparator.maxReservoirs')}</p>
          <select
            className="w-full border rounded px-3 py-2 mb-2"
            onChange={(e) => {
              const r = reservoirs?.find((x) => x.id === Number(e.target.value))
              if (r) addReservoir(r)
            }}
            value=""
          >
            <option value="">{t('comparator.add')}...</option>
            {reservoirs?.map((r) => (
              <option key={r.id} value={r.id}>
                {r.name} ({r.basin_name})
              </option>
            ))}
          </select>

          <div className="space-y-2">
            {selected.map((s) => (
              <div key={s.id} className="flex justify-between items-center bg-gray-50 rounded px-3 py-2">
                <span>{s.name}</span>
                <button onClick={() => removeReservoir(s.id)} className="text-red-600 text-sm hover:underline">
                  {t('comparator.remove')}
                </button>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-2">{t('comparator.metric')}</h2>
          <div className="flex gap-4">
            <button
              onClick={() => setMetric('fill_pct')}
              className={`px-4 py-2 rounded border ${metric === 'fill_pct' ? 'bg-blue-600 text-white border-blue-600' : 'bg-white border-gray-300'}`}
            >
              % Llenado
            </button>
            <button
              onClick={() => setMetric('volume_hm3')}
              className={`px-4 py-2 rounded border ${metric === 'volume_hm3' ? 'bg-blue-600 text-white border-blue-600' : 'bg-white border-gray-300'}`}
            >
              Volumen (hm³)
            </button>
          </div>
        </div>
      </div>

      {selected.length > 0 && (
        <div className="bg-white rounded-lg shadow p-4" style={{ height: 400 }}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Legend />
              {selected.map((s, i) => (
                <Line
                  key={s.id}
                  type="monotone"
                  dataKey={`r${i + 1}`}
                  name={s.name}
                  stroke={colors[i]}
                  strokeWidth={2}
                  dot={false}
                />
              ))}
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  )
}
