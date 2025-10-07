import { buildUrl, fetchJson } from "./http";

export type YearlyInitRequest = {
  branches: string[];
  debt_ym: string;
};

export type YearlyInitResponse = {
  message?: string; // Present in async 202 response
  fiscal_year: number;
  branches: string[];
  debt_ym: string;
  stats?: {
    // Optional - only present in sync completion
    upserted: number;
  };
  started_at: string;
  finished_at?: string; // Optional - only present in sync completion
  note?: string;
};

export type MonthlySyncRequest = {
  branches: string[];
  ym: string;
};

export type MonthlySyncResponse = {
  message?: string; // Present in async 202 response
  ym: string;
  branches: string[];
  stats?: {
    // Optional - only present in sync completion
    upserted: number;
    zeroed: number;
  };
  started_at: string;
  finished_at?: string; // Optional - only present in sync completion
  note?: string;
};

/**
 * Trigger yearly initialization sync
 * POST /api/v1/sync/init
 */
export async function triggerYearlyInit(
  req: YearlyInitRequest,
): Promise<YearlyInitResponse> {
  const url = buildUrl("/api/v1/sync/init");
  return fetchJson<YearlyInitResponse>(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(req),
  });
}

/**
 * Trigger monthly sync
 * POST /api/v1/sync/monthly
 */
export async function triggerMonthlySync(
  req: MonthlySyncRequest,
): Promise<MonthlySyncResponse> {
  const url = buildUrl("/api/v1/sync/monthly");
  return fetchJson<MonthlySyncResponse>(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(req),
  });
}
