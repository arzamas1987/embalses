import { QueryClient } from '@tanstack/react-query'
import type { APIResponse, Source, ReservoirSummary, ReservoirDetail, Reading, Basin, BasinSummary, BasinDetail, RankingItem, DataQualityReport, QueryIntent } from '../types'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

const API_KEY = import.meta.env.VITE_API_KEY || 'test-key-123'
const API_BASE = '/api'

async function fetchAPI<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      'X-API-Key': API_KEY,
      'Content-Type': 'application/json',
    },
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error?.message || `HTTP ${res.status}`)
  }
  return res.json()
}

export async function getSources() {
  return fetchAPI<APIResponse<Source[]>>('/v1/sources')
}

export async function getReservoirs(page = 1, perPage = 20) {
  return fetchAPI<APIResponse<ReservoirSummary[]>>(`/v1/reservoirs?page=${page}&per_page=${perPage}`)
}

export async function getReservoir(slug: string) {
  return fetchAPI<APIResponse<ReservoirDetail>>(`/v1/reservoirs/${encodeURIComponent(slug)}`)
}

export async function getReservoirReadings(slug: string, since?: string, until?: string) {
  let qs = `?per_page=500`
  if (since) qs += `&since=${since}`
  if (until) qs += `&until=${until}`
  return fetchAPI<APIResponse<Reading[]>>(`/v1/reservoirs/${encodeURIComponent(slug)}/readings${qs}`)
}

export async function getBasins() {
  return fetchAPI<APIResponse<Basin[]>>('/v1/basins')
}

export async function getBasinSummary() {
  return fetchAPI<APIResponse<BasinSummary[]>>('/v1/basins/summary')
}

export async function getBasinDetail(slug: string) {
  return fetchAPI<APIResponse<BasinDetail>>(`/v1/basins/${encodeURIComponent(slug)}`)
}

export async function getRankings(metric = 'fullest', limit = 10) {
  return fetchAPI<APIResponse<RankingItem[]>>(`/v1/rankings/reservoirs?metric=${metric}&limit=${limit}`)
}

export async function getDataQuality() {
  return fetchAPI<APIResponse<DataQualityReport>>('/v1/data-quality')
}

export async function importReadingsCSV(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  const res = await fetch(`${API_BASE}/admin/readings/import`, {
    method: 'POST',
    headers: {
      'X-API-Key': API_KEY,
    },
    body: formData,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: { message: res.statusText } }))
    throw new Error(err.error?.message || `HTTP ${res.status}`)
  }
  return res.json() as Promise<APIResponse<{ imported: number }>>
}

export async function getComparatorData(slugs: string[], since?: string, until?: string) {
  const params = new URLSearchParams()
  slugs.forEach((s) => params.append('reservoir', s))
  if (since) params.set('since', since)
  if (until) params.set('until', until)
  return fetchAPI<APIResponse<Record<string, Reading[]>>>(`/v1/compare?${params.toString()}`)
}

export async function postQuery(intent: QueryIntent) {
  const res = await fetch(`${API_BASE}/v1/query`, {
    method: 'POST',
    headers: {
      'X-API-Key': API_KEY,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(intent),
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error?.message || `HTTP ${res.status}`)
  }
  return res.json()
}
