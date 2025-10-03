import { buildUrl, fetchJson } from './http'

export interface BranchItem { code: string; name?: string }
export interface BranchesResponse { items: BranchItem[]; total: number; limit: number; offset: number }

export const getBranches = (params: { q?: string; limit?: number; offset?: number } = {}) =>
  fetchJson<BranchesResponse>(buildUrl('/api/v1/branches', params))

