import { useQuery } from '@tanstack/react-query'
import {
  getRankings,
  getDataQuality,
  getReservoirs,
  getReservoir,
  getReservoirReadings,
  getBasins,
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

export function useBasins() {
  return useQuery({
    queryKey: ['basins'],
    queryFn: getBasins,
  })
}

export function useSources() {
  return useQuery({
    queryKey: ['sources'],
    queryFn: getSources,
  })
}
