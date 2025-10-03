/**
 * Utility functions for DetailPage
 */

import type { DetailItem } from "../../api/details";
import type { CustCodeItem } from "../../api/custcodes";
import type { BranchItem } from "../../api/branches";

export type Row = {
  key: string;
  branchCode: string;
  orgName: string | null;
  custCode: string;
  useType: string | null;
  useName: string | null;
  custName: string | null;
  address: string | null;
  routeCode: string | null;
  meterNo: string | null;
  meterSize: string | null;
  meterBrand: string | null;
  meterState: string | null;
  average: number | null;
  presentMeterCount: number | null;
  values: Record<string, number>;
};

export const DEFAULT_THRESHOLD = 10;
export const MAX_HISTORY_MONTHS = 12;
export const MONTH_OPTIONS = [3, 6, 12] as const;
export const STORAGE_KEYS = {
  threshold: "detail.threshold",
  months: "detail.months",
};

export const TH_MONTHS = [
  "มกราคม",
  "กุมภาพันธ์",
  "มีนาคม",
  "เมษายน",
  "พฤษภาคม",
  "มิถุนายน",
  "กรกฎาคม",
  "สิงหาคม",
  "กันยายน",
  "ตุลาคม",
  "พฤศจิกายน",
  "ธันวาคม",
];

export const TH_MONTH_ABBR = [
  "ม.ค.",
  "ก.พ.",
  "มี.ค.",
  "เม.ย.",
  "พ.ค.",
  "มิ.ย.",
  "ก.ค.",
  "ส.ค.",
  "ก.ย.",
  "ต.ค.",
  "พ.ย.",
  "ธ.ค.",
] as const;

export function fmtNum(value: number) {
  return new Intl.NumberFormat("th-TH", { maximumFractionDigits: 2 }).format(
    value
  );
}

export function fmtPct(value: number) {
  return `${new Intl.NumberFormat("th-TH", { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(value)}%`;
}

export function fmtThMonth(ym: string) {
  const parts = fmtThMonthParts(ym);
  return `${parts.label} ${parts.year}`;
}

export function fmtThMonthParts(ym: string) {
  const year = Number(ym.slice(0, 4)) + 543;
  const month = Number(ym.slice(4, 6));
  return { label: TH_MONTH_ABBR[month - 1], year: String(year).slice(-2) };
}

export function ymParts(ym: string) {
  return { y: Number(ym.slice(0, 4)), m: Number(ym.slice(4, 6)) };
}

export function partsToYm(parts: { y: number; m: number }) {
  return `${parts.y}${String(parts.m).padStart(2, "0")}`;
}

export function yearOptions() {
  const current = new Date().getFullYear();
  return [current + 1, current, current - 1, current - 2, current - 3];
}

export function defaultLatestYm() {
  const now = new Date();
  const base =
    now.getDate() < 16
      ? new Date(now.getFullYear(), now.getMonth() - 1, 1)
      : new Date(now.getFullYear(), now.getMonth(), 1);
  return `${base.getFullYear()}${String(base.getMonth() + 1).padStart(2, "0")}`;
}

export function normalizeThreshold(value: string) {
  const num = Number(value);
  if (!Number.isFinite(num)) return DEFAULT_THRESHOLD;
  return Math.min(100, Math.max(0, Math.round(num)));
}

export function coalesce<T>(
  current: T | null | undefined,
  next: T | null | undefined
) {
  return next != null ? next : (current ?? null);
}

export function loadNumber(key: string, fallback: number) {
  if (typeof window === "undefined") return fallback;
  const raw = window.localStorage.getItem(key);
  const parsed = raw == null ? NaN : Number(raw);
  return Number.isFinite(parsed) ? parsed : fallback;
}

export function persistNumber(key: string, value: number) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(key, String(value));
}

export function coerceHistoryMonths(value: number) {
  if (MONTH_OPTIONS.includes(value as (typeof MONTH_OPTIONS)[number])) {
    return value;
  }
  return MONTH_OPTIONS[1];
}

export function formatBranchLabel(item: BranchItem) {
  return item.name ? `${item.code} - ${item.name}` : item.code;
}

export function buildMonths(latest: string, count: number) {
  const months: string[] = [];
  let year = Number(latest.slice(0, 4));
  let month = Number(latest.slice(4, 6));
  for (let i = 0; i < count; i++) {
    months.push(`${year}${String(month).padStart(2, "0")}`);
    month -= 1;
    if (month === 0) {
      month = 12;
      year -= 1;
    }
  }
  return months;
}

export function computePct(prev?: number, curr?: number) {
  if (prev == null || prev === 0) return null;
  const prevVal = prev;
  const currVal = curr ?? 0;
  return ((currVal - prevVal) / prevVal) * 100;
}

export function combineRows(
  months: string[],
  detailLists: DetailItem[][],
  metaItems: CustCodeItem[]
): Row[] {
  const metaMap = new Map(metaItems.map((item) => [item.cust_code, item]));
  const rows = new Map<string, Row>();
  const ensureRow = (custCode: string) => {
    const existing = rows.get(custCode);
    if (existing) return existing;
    const seed = metaMap.get(custCode);
    const base: Row = {
      key: custCode,
      branchCode: seed?.branch_code ?? "",
      orgName: seed?.org_name ?? null,
      custCode,
      useType: seed?.use_type ?? null,
      useName: seed?.use_name ?? null,
      custName: seed?.cust_name ?? null,
      address: seed?.address ?? null,
      routeCode: seed?.route_code ?? null,
      meterNo: seed?.meter_no ?? null,
      meterSize: seed?.meter_size ?? null,
      meterBrand: seed?.meter_brand ?? null,
      meterState: seed?.meter_state ?? null,
      average: null,
      presentMeterCount: null,
      values: {},
    };
    rows.set(custCode, base);
    return base;
  };

  detailLists.forEach((items, index) => {
    const ym = months[index];
    items.forEach((detail) => {
      const row = ensureRow(detail.cust_code);
      row.branchCode = detail.branch_code ?? row.branchCode;
      row.orgName = coalesce(row.orgName, detail.org_name);
      row.useType = coalesce(row.useType, detail.use_type);
      row.useName = coalesce(row.useName, detail.use_name);
      row.custName = coalesce(row.custName, detail.cust_name);
      row.address = coalesce(row.address, detail.address);
      row.routeCode = coalesce(row.routeCode, detail.route_code);
      row.meterNo = coalesce(row.meterNo, detail.meter_no);
      row.meterSize = coalesce(row.meterSize, detail.meter_size);
      row.meterBrand = coalesce(row.meterBrand, detail.meter_brand);
      row.meterState = coalesce(row.meterState, detail.meter_state);
      row.values[ym] = detail.present_water_usg ?? 0;
      if (index === 0) {
        row.average = detail.average ?? row.average;
        row.presentMeterCount =
          detail.present_meter_count ?? row.presentMeterCount;
      }
    });
  });

  metaItems.forEach((item) => {
    const row = ensureRow(item.cust_code);
    row.branchCode = item.branch_code ?? row.branchCode;
    row.orgName = coalesce(row.orgName, item.org_name);
    row.useType = coalesce(row.useType, item.use_type);
    row.useName = coalesce(row.useName, item.use_name);
    row.custName = coalesce(row.custName, item.cust_name);
    row.address = coalesce(row.address, item.address);
    row.routeCode = coalesce(row.routeCode, item.route_code);
    row.meterNo = coalesce(row.meterNo, item.meter_no);
    row.meterSize = coalesce(row.meterSize, item.meter_size);
    row.meterBrand = coalesce(row.meterBrand, item.meter_brand);
    row.meterState = coalesce(row.meterState, item.meter_state);
  });

  return Array.from(rows.values());
}

export function filterRows(
  rows: Row[],
  latestYm: string | undefined,
  prevYm: string | undefined,
  threshold: number,
  search: string
) {
  const query = search.trim().toLowerCase();
  return rows.filter((row) => {
    const current = latestYm ? (row.values[latestYm] ?? 0) : 0;
    const previous = prevYm ? (row.values[prevYm] ?? 0) : 0;
    const pct = previous > 0 ? ((current - previous) / previous) * 100 : null;
    const passesThreshold =
      threshold <= 0 || (pct != null && pct <= -threshold);
    if (!passesThreshold) return false;
    if (!query) return true;
    return [
      row.orgName,
      row.custCode,
      row.useType,
      row.useName,
      row.custName,
      row.address,
      row.routeCode,
      row.meterNo,
      row.meterSize,
      row.meterBrand,
      row.meterState,
    ]
      .map((value) => (value == null ? "" : String(value).toLowerCase()))
      .some((value) => value.includes(query));
  });
}

export function resolveBadgeClass(pct: number | null) {
  if (pct == null || !isFinite(pct) || pct >= 0)
    return "bg-slate-100 text-slate-800";
  const drop = Math.abs(pct);
  if (drop > 30) return "bg-red-500/20 text-red-700";
  if (drop >= 15) return "bg-orange-500/20 text-orange-700";
  if (drop >= 5) return "bg-yellow-400/30 text-yellow-800";
  return "bg-slate-100 text-slate-800";
}

export function sanitizeFileNamePart(part: string) {
  const normalized = part.trim();
  if (!normalized) return "all";
  return normalized.replace(/[^0-9A-Za-zก-๛_-]+/g, "-");
}

type KnownError = Error & { status?: number };

export function interpretErrorMessage(error: KnownError) {
  const status = error.status;
  const raw = (error.message ?? "").trim();
  if (status && status >= 500) {
    const main = /number of field descriptions/i.test(raw)
      ? "ระบบยังไม่สามารถเตรียมข้อมูล Top-200 ของสาขานี้ได้ โปรดลองอีกครั้งภายหลัง"
      : "ระบบไม่สามารถเชื่อมต่อข้อมูลจากเซิร์ฟเวอร์ได้ในขณะนี้";
    return { main, detail: raw || undefined };
  }
  if (status === 404) {
    return { main: "ไม่พบข้อมูลที่ร้องขอ", detail: raw || undefined };
  }
  if (status === 400) {
    return {
      main: "พารามิเตอร์ไม่ถูกต้อง กรุณาตรวจสอบค่าอีกครั้ง",
      detail: raw || undefined,
    };
  }
  if (raw) return { main: raw, detail: undefined };
  return { main: "เกิดข้อผิดพลาดไม่ทราบสาเหตุ", detail: undefined };
}
