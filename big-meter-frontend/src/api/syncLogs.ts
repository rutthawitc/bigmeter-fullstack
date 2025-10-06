import { buildUrl, fetchJson } from "./http";

export interface SyncLog {
  id: number;
  sync_type: string;
  branch_code: string;
  year_month?: string | null;
  fiscal_year?: number | null;
  debt_ym?: string | null;
  status: string;
  started_at: string;
  finished_at?: string | null;
  duration_ms?: number | null;
  records_upserted?: number | null;
  records_zeroed?: number | null;
  error_message?: string | null;
  triggered_by: string;
  created_at: string;
}

export interface GetSyncLogsParams {
  branch?: string;
  sync_type?: string;
  status?: string;
  limit?: number;
  offset?: number;
}

export interface SyncLogsResponse {
  items: SyncLog[];
  total: number;
  limit: number;
  offset: number;
}

export async function getSyncLogs(
  params: GetSyncLogsParams = {},
): Promise<SyncLogsResponse> {
  const queryParams: Record<string, string> = {};

  if (params.branch) queryParams.branch = params.branch;
  if (params.sync_type) queryParams.sync_type = params.sync_type;
  if (params.status) queryParams.status = params.status;
  if (params.limit) queryParams.limit = String(params.limit);
  if (params.offset) queryParams.offset = String(params.offset);

  const url = buildUrl("/api/v1/sync/logs", queryParams);
  return fetchJson<SyncLogsResponse>(url);
}
