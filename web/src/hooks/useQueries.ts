import { useQuery } from '@tanstack/react-query'
import {
  getRankings,
  getDataQuality,
  getReservoirs,
  getReservoir,
  getReservoirReadings,
  getComparatorData,
  getBasins,
  getBasinSummary,
  getBasinDetail,
  getSources,
} from '../api/client'

export function useRankings(metric = 'fullest', limit = 5) {
  return useQuery({
    queryKey: ['rankings', metric, limit],
    queryFn: () => getRankings(metric, limit),
  })
}

export function useDataQuality() {
  return useQuery({
    queryKey: ['dataQuality'],
    queryFn: getDataQuality,
  })
}

export function useReservoirs(page = 1, perPage = 20) {
  return useQuery({
    queryKey: ['reservoirs', page, perPage],
    queryFn: () => getReservoirs(page, perPage),
  })
}

export function useAllReservoirs() {
  return useQuery({
    queryKey: ['reservoirs', 'all'],
    queryFn: () => getReservoirs(1, 10000),
  })
}

export function useReservoir(slug: string) {
  return useQuery({
    queryKey: ['reservoir', slug],
    queryFn: () => getReservoir(slug),
    enabled: !!slug,
  })
}

export function useReservoirReadings(slug: string, since?: string, until?: string) {
  return useQuery({
    queryKey: ['readings', slug, since, until],
    queryFn: () => getReservoirReadings(slug, since, until),
    enabled: !!slug,
  })
}

export function useComparatorData(slugs: string[], since?: string, until?: string) {
  return useQuery({
    queryKey: ['compare', slugs.join(','), since, until],
    queryFn: () => getComparatorData(slugs, since, until),
    enabled: slugs.length > 0,
  })
}

export function useBasins() {
  return useQuery({
    queryKey: ['basins'],
    queryFn: getBasins,
  })
}

export function useBasinSummary() {
  return useQuery({
    queryKey: ['basinSummary'],
    queryFn: getBasinSummary,
  })
}

export function useBasinDetail(slug: string) {
  return useQuery({
    queryKey: ['basinDetail', slug],
    queryFn: () => getBasinDetail(slug),
    enabled: !!slug,
  })
}

export function useSources() {
  return useQuery({
    queryKey: ['sources'],
    queryFn: getSources,
  })
}
