import { buildUrl, fetchJson } from './http'

export interface CustCodeItem {
  fiscal_year: number
  branch_code: string
  org_name: string | null
  cust_code: string
  use_type: string | null
  use_name: string | null
  cust_name: string | null
  address: string | null
  route_code: string | null
  meter_no: string | null
  meter_size: string | null
  meter_brand: string | null
  meter_state: string | null
  debt_ym: string | null
  created_at: string | null
}

export interface CustCodesResponse { items: CustCodeItem[]; total?: number; limit?: number; offset?: number }

export function getCustCodes(params: { branch: string; ym?: string; fiscal_year?: number; q?: string; limit?: number; offset?: number }) {
  return fetchJson<CustCodesResponse>(buildUrl('/api/v1/custcodes', params as Record<string, string | number | boolean | undefined>))
}

