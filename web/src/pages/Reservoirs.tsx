import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { useReservoirs } from '../hooks/useQueries'
import type { ReservoirSummary } from '../types'

type SortKey = 'name' | 'basin_name' | 'province_name' | 'latest_fill_pct' | 'capacity_hm3'
type SortDir = 'asc' | 'desc'

const SearchIcon = () => (
  <svg className="w-4 h-4 text-[#94a3b8]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
  </svg>
)

const SortAscIcon = () => (
  <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M5 15l7-7 7 7" />
  </svg>
)

const SortDescIcon = () => (
  <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
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
    <div className="gov-progress-bar w-24">
      <div
        className="gov-progress-fill"
        style={{ width: `${pct}%`, backgroundColor: color }}
      />
    </div>
  )
}

export default function Reservoirs() {
  const { t } = useTranslation()
  const { data, isLoading } = useReservoirs(1, 100)
  const reservoirs = data?.data as ReservoirSummary[] | undefined

  const [search, setSearch] = useState('')
  const [sortKey, setSortKey] = useState<SortKey>('name')
  const [sortDir, setSortDir] = useState<SortDir>('asc')

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(sortDir === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDir('asc')
    }
  }

  const sorted = useMemo(() => {
    if (!reservoirs) return []
    const filtered = search
      ? reservoirs.filter(
          (r) =>
            r.name.toLowerCase().includes(search.toLowerCase()) ||
            (r.basin_name?.toLowerCase() || '').includes(search.toLowerCase()) ||
            (r.province_name?.toLowerCase() || '').includes(search.toLowerCase())
        )
      : [...reservoirs]

    filtered.sort((a, b) => {
      const dir = sortDir === 'asc' ? 1 : -1
      const aVal = a[sortKey]
      const bVal = b[sortKey]
      if (aVal == null && bVal == null) return 0
      if (aVal == null) return 1
      if (bVal == null) return -1
      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return aVal.localeCompare(bVal) * dir
      }
      return ((aVal as number) - (bVal as number)) * dir
    })
    return filtered
  }, [reservoirs, search, sortKey, sortDir])

  const SortHeader = ({ label, key }: { label: string; key: SortKey }) => (
    <th onClick={() => handleSort(key)} className="select-none">
      <div className="flex items-center gap-1.5">
        {label}
        {sortKey === key && (sortDir === 'asc' ? <SortAscIcon /> : <SortDescIcon />)}
      </div>
    </th>
  )

  return (
    <div className="animate-fade-in">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-[#0f172a]">{t('nav.reservoirs')}</h1>
          <p className="text-[#475569] text-sm mt-1">
            {reservoirs?.length ?? 0} embalses registrados en el sistema
          </p>
        </div>
        <div className="relative">
          <div className="absolute left-3 top-1/2 -translate-y-1/2">
            <SearchIcon />
          </div>
          <input
            type="text"
            placeholder="Buscar embalse..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9 pr-4 py-2.5 w-full sm:w-72 border border-[#e2e8f0] rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-[#003366]/20 focus:border-[#003366] transition-all bg-white"
          />
        </div>
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
        <div className="gov-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="gov-table">
              <thead>
                <tr>
                  <SortHeader key="name" label={t('reservoir.basin')} />
                  <SortHeader key="basin_name" label={t('reservoir.basin')} />
                  <SortHeader key="province_name" label={t('reservoir.province')} />
                  <SortHeader key="latest_fill_pct" label={t('reservoir.fillPercent')} />
                  <SortHeader key="capacity_hm3" label={t('reservoir.capacity')} />
                </tr>
              </thead>
              <tbody>
                {sorted.map((r) => {
                  const pct = r.latest_fill_pct
                  return (
                    <tr key={r.id}>
                      <td>
                        <Link
                          to={`/embalses/${encodeURIComponent(r.external_id)}`}
                          className="font-medium text-[#003366] hover:text-[#004a74] hover:underline"
                        >
                          {r.name}
                        </Link>
                      </td>
                      <td>{r.basin_name ?? '-'}</td>
                      <td>{r.province_name ?? '-'}</td>
                      <td>
                        <div className="flex items-center gap-3">
                          <FillBar pct={pct ?? 0} />
                          <span className={`gov-badge ${getFillBadgeClass(pct ?? 0)}`}>
                            {pct != null ? `${Math.round(pct)}%` : '-'}
                          </span>
                        </div>
                      </td>
                      <td className="text-right font-medium">
                        {r.capacity_hm3 != null ? `${r.capacity_hm3.toLocaleString()} hm³` : '-'}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
          {!sorted.length && (
            <div className="p-8 text-center text-[#94a3b8]">
              <svg className="w-12 h-12 mx-auto mb-3 text-[#cbd5e1]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p className="font-medium">No se encontraron embalses</p>
              <p className="text-sm mt-1">Prueba con otro término de búsqueda</p>
            </div>
          )}
          <div className="px-4 py-3 bg-[#f8fafc] border-t border-[#e2e8f0] text-xs text-[#475569]">
            Mostrando {sorted.length} de {reservoirs?.length ?? 0} embalses
            {search && ` — filtrado por "${search}"`}
          </div>
        </div>
      )}
    </div>
  )
}
