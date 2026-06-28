export interface APIResponse<T> {
  data?: T;
  meta?: Meta;
  error?: APIError;
  lineage?: Lineage;
}

export interface Meta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface APIError {
  code: string;
  message: string;
}

export interface Lineage {
  source: string;
  licence: string;
  attribution: string;
  fetched_at?: string;
}

export interface ReservoirSummary {
  id: number;
  name: string;
  external_id: string;
  basin_name?: string;
  province_name?: string;
  capacity_hm3?: number;
  latest_fill_pct?: number;
  latitude?: number;
  longitude?: number;
}

export interface ReservoirDetail {
  id: number;
  name: string;
  external_id: string;
  basin_name?: string;
  province_name?: string;
  capacity_hm3?: number;
  latest_volume_hm3?: number;
  latest_fill_pct?: number;
  dam_name?: string;
}

export interface Reading {
  observed_at: string;
  volume_hm3: number;
  capacity_hm3: number;
  fill_pct: number;
  weekly_variation_hm3?: number;
  is_provisional: boolean;
}

export interface Source {
  name: string;
  organism: string;
  licence: string;
  attribution: string;
  url?: string;
}

export interface Basin {
  id: number;
  name: string;
  code?: string;
}

export interface RankingItem {
  rank: number;
  reservoir_id: number;
  name: string;
  value: number;
  metric: string;
}

export interface DataQualityReport {
  total_reservoirs: number;
  reservoirs_with_readings: number;
  latest_reading_date?: string;
  oldest_reading_date?: string;
  provisional_count: number;
  official_count: number;
}

export interface QueryIntent {
  entity: string;
  metrics: string[];
  filters?: {
    slugs?: string[];
    basin?: string;
    province?: string;
    since?: string;
    until?: string;
  };
  aggregation: string;
  sort?: { field: string; order?: string };
  limit?: number;
  chart_hint?: string;
}
