/**
 * DataTable component for DetailPage
 * Displays water usage data in a responsive table with sparkline charts
 */

import { useId, useMemo, useState } from "react";
import { Scale, Shape } from "@visx/visx";
import type { Row } from "./utils";
import {
  fmtNum,
  fmtPct,
  fmtThMonthParts,
  computePct,
  resolveBadgeClass,
} from "./utils";

const { scaleLinear } = Scale;
const { AreaClosed, LinePath } = Shape;

type TrendPoint = {
  ym: string;
  value: number;
};

export interface DataTableProps {
  rows: Row[];
  months: string[];
  latestYm: string;
  baseIndex: number;
  isMobile: boolean;
}

export function DataTable({
  rows,
  months,
  latestYm: _latestYm,
  baseIndex,
  isMobile,
}: DataTableProps) {
  if (!rows.length) return null;
  const prevYm = months[1];
  return (
    <div className="mt-6 overflow-x-auto">
      <table className="min-w-full text-xs md:text-sm">
        <thead className="bg-slate-100 text-slate-700">
          <tr>
            <th className="p-3 text-left">ลำดับ</th>
            <th className="hidden p-3 text-left md:table-cell">กปภ.สาขา</th>
            <th className="p-3 text-left">เลขที่ผู้ใช้น้ำ</th>
            <th className="hidden p-3 text-left md:table-cell">ประเภท</th>
            <th className="hidden p-3 text-left md:table-cell">รายละเอียด</th>
            <th className="p-3 text-left">ชื่อผู้ใช้น้ำ</th>
            <th className="p-3 text-left">ที่อยู่</th>
            <th className="hidden p-3 text-left md:table-cell">เส้นทาง</th>
            <th className="hidden p-3 text-left md:table-cell">หมายเลขมาตร</th>
            <th className="p-3 text-left">ขนาดมาตร</th>
            <th className="hidden p-3 text-left md:table-cell">ยี่ห้อ</th>
            <th className="hidden p-3 text-left md:table-cell">สถานะมาตร</th>
            <th className="p-3 text-right">หน่วยน้ำเฉลี่ย</th>
            <th className="p-3 text-right">เลขมาตรที่อ่านได้</th>
            {months.map((ym) => (
              <th key={ym} className="p-3 text-right">
                <MonthHeader ym={ym} />
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-200">
          {rows.map((row, index) => {
            const historyData = months.map((monthYm) => ({
              ym: monthYm,
              value: row.values[monthYm] ?? 0,
            }));
            return (
              <tr key={row.key} className="hover:bg-slate-50">
                <td className="p-3">{baseIndex + index + 1}</td>
                <td className="hidden p-3 md:table-cell">
                  {row.orgName ?? row.branchCode}
                </td>
                <td className="p-3 font-mono">{row.custCode}</td>
                <td className="hidden p-3 md:table-cell">{row.useType ?? "-"}</td>
                <td className="hidden p-3 md:table-cell">{row.useName ?? "-"}</td>
                <td className="p-3">{row.custName ?? "-"}</td>
                <td className="p-3">{row.address ?? "-"}</td>
                <td className="hidden p-3 md:table-cell">{row.routeCode ?? "-"}</td>
                <td className="hidden p-3 font-mono md:table-cell">
                  {row.meterNo ?? "-"}
                </td>
                <td className="p-3 text-center">{row.meterSize ?? "-"}</td>
                <td className="hidden p-3 md:table-cell">{row.meterBrand ?? "-"}</td>
                <td className="hidden p-3 md:table-cell">{row.meterState ?? "-"}</td>
                <td className="p-3 text-right">
                  {row.average != null ? fmtNum(row.average) : "-"}
                </td>
                <td className="p-3 text-right">
                  {row.presentMeterCount != null
                    ? fmtNum(row.presentMeterCount)
                    : "-"}
                </td>
                {months.map((ym, monthIndex) => {
                  const value = historyData[monthIndex].value;
                  if (monthIndex === 0) {
                    const pct = prevYm
                      ? computePct(row.values[prevYm], value)
                      : null;
                    return (
                      <td key={ym} className="p-3 text-right">
                        <LatestUsageCell
                          value={value}
                          pct={pct}
                          history={historyData}
                          isMobile={isMobile}
                        />
                      </td>
                    );
                  }
                  return (
                    <td key={ym} className="p-3 text-right">
                      {fmtNum(value)}
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function MonthHeader({ ym }: { ym: string }) {
  const parts = fmtThMonthParts(ym);
  return (
    <span className="flex flex-col items-end leading-tight text-[10px] md:text-xs lg:text-sm">
      <span>{parts.label}</span>
      <span>{parts.year}</span>
    </span>
  );
}

function LatestUsageCell({
  value,
  pct,
  history,
  isMobile,
}: {
  value: number;
  pct: number | null;
  history: TrendPoint[];
  isMobile: boolean;
}) {
  const [isHover, setIsHover] = useState(false);
  const pctText =
    pct != null && isFinite(pct) && pct !== 0 ? ` (${fmtPct(pct)})` : "";
  const badgeClass = resolveBadgeClass(pct);
  const showSparkline = isHover && history.length > 1;
  const latestLabel = fmtThMonthParts(history[0].ym);
  const { width, height } = isMobile
    ? { width: 160, height: 60 }
    : { width: 260, height: 90 };

  return (
    <div
      className="relative flex justify-end focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-400/60"
      onMouseEnter={() => setIsHover(true)}
      onMouseLeave={() => setIsHover(false)}
      onFocus={() => setIsHover(true)}
      onBlur={() => setIsHover(false)}
      onKeyDown={(event) => {
        if (event.key === "Escape") setIsHover(false);
      }}
      tabIndex={0}
      role="button"
      aria-haspopup="dialog"
      aria-expanded={showSparkline}
      aria-label={`ดูแนวโน้มล่าสุดสำหรับเดือน ${latestLabel.label} ${latestLabel.year}`}
    >
      <span className={`inline-block rounded px-2 py-1 text-xs md:text-sm ${badgeClass}`}>
        {fmtNum(value)}
        {pctText}
      </span>
      {showSparkline && (
        <div
          className="absolute right-0 top-full z-20 mt-2 overflow-hidden rounded-lg border border-slate-200 bg-white p-3 text-[10px] text-slate-500 shadow-lg"
          style={{ width: `${width}px` }}
        >
          <TrendSparkline data={history} width={width} height={height} />
        </div>
      )}
    </div>
  );
}

function TrendSparkline({
  data,
  width,
  height,
}: {
  data: TrendPoint[];
  width: number;
  height: number;
}) {
  const ordered = useMemo(() => [...data].reverse(), [data]);
  const gradientId = useId();

  const paddingY = 12;
  const paddingX = useMemo(() => Math.max(width * 0.06, 10), [width]);

  const { domainMin, domainMax, xDomainMax } = useMemo(() => {
    const values = ordered.map((point) => point.value);
    const minVal = Math.min(...values);
    const maxVal = Math.max(...values);
    return {
      domainMin: minVal === maxVal ? minVal - 1 : minVal,
      domainMax: minVal === maxVal ? maxVal + 1 : maxVal,
      xDomainMax: Math.max(ordered.length - 1, 1),
    };
  }, [ordered]);

  const xScale = useMemo(
    () =>
      scaleLinear({
        domain: [0, xDomainMax],
        range: [paddingX, width - paddingX],
      }),
    [xDomainMax, paddingX, width]
  );
  const yScale = useMemo(
    () =>
      scaleLinear({
        domain: [domainMin, domainMax],
        range: [height - paddingY, paddingY],
      }),
    [domainMin, domainMax, height, paddingY]
  );

  if (!ordered.length) return null;

  const oldest = ordered[0];
  const latest = ordered[ordered.length - 1];
  const oldestLabel = fmtThMonthParts(oldest.ym);
  const latestLabel = fmtThMonthParts(latest.ym);

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between text-[10px] uppercase tracking-wide text-slate-400">
        <span>
          {oldestLabel.label} {oldestLabel.year}
        </span>
        <span>
          {latestLabel.label} {latestLabel.year}
        </span>
      </div>
      <svg width={width} height={height} role="presentation">
        <defs>
          <linearGradient id={gradientId} x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="#60a5fa" stopOpacity={0.4} />
            <stop offset="100%" stopColor="#2563eb" stopOpacity={0.05} />
          </linearGradient>
        </defs>
        <AreaClosed
          data={ordered}
          x={(point, index) => xScale(index) ?? 0}
          y={(point) => yScale(point.value) ?? 0}
          yScale={yScale}
          fill={`url(#${gradientId})`}
          stroke="none"
        />
        <LinePath
          data={ordered}
          x={(point, index) => xScale(index) ?? 0}
          y={(point) => yScale(point.value) ?? 0}
          stroke="#2563eb"
          strokeWidth={2}
        />
        <circle
          cx={xScale(ordered.length - 1)}
          cy={yScale(latest.value)}
          r={3}
          fill="#2563eb"
          stroke="white"
          strokeWidth={1.5}
        />
      </svg>
      <div className="flex items-center justify-between text-[10px] text-slate-600">
        <span>{fmtNum(oldest.value)}</span>
        <span className="font-medium text-slate-700">{fmtNum(latest.value)}</span>
      </div>
    </div>
  );
}
