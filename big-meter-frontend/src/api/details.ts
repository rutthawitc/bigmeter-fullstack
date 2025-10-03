import { buildUrl, fetchJson } from './http'

export interface DetailItem {
  year_month: string
  branch_code: string
  org_name?: string | null
  cust_code: string
  use_type?: string | null
  use_name?: string | null
  cust_name?: string | null
  address?: string | null
  route_code?: string | null
  meter_no: string | null
  meter_size?: string | null
  meter_brand?: string | null
  meter_state?: string | null
  average?: number
  present_water_usg: number
  present_meter_count: number
  is_zeroed?: boolean
}

export interface DetailsResponse { items: DetailItem[]; total: number; limit: number; offset: number }

export function getDetails(params: {
  ym: string
  branch: string
  q?: string
  limit?: number
  offset?: number
  order_by?: string
  sort?: 'ASC' | 'DESC' | 'asc' | 'desc'
}) {
  return fetchJson<DetailsResponse>(buildUrl('/api/v1/details', params as Record<string, string | number | boolean | undefined>))
}
