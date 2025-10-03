/**
 * ResultsHeader component for DetailPage
 * Displays result count, legend, search input, export button, history months toggle, and page size selector
 */

import { Legend } from "./UIComponents";
import { MONTH_OPTIONS } from "./utils";

export interface ResultsHeaderProps {
  filteredCount: number;
  search: string;
  setSearch: (value: string) => void;
  pageSize: number;
  setPageSize: (value: number) => void;
  historyMonths: number;
  setHistoryMonths: (value: number) => void;
  onExport: () => void;
  isExporting: boolean;
  applied: boolean;
  isMobile: boolean;
}

export function ResultsHeader({
  filteredCount,
  search,
  setSearch,
  pageSize,
  setPageSize,
  historyMonths,
  setHistoryMonths,
  onExport,
  isExporting,
  applied,
  isMobile,
}: ResultsHeaderProps) {
  return (
    <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
      <div>
        <h3 className="text-lg font-semibold text-slate-800">
          ผลลัพธ์: <span className="text-blue-600">{filteredCount}</span>{" "}
          รายชื่อที่เข้าเงื่อนไข
        </h3>
        <div className="mt-2 flex flex-wrap items-center gap-4 text-sm text-slate-500">
          <span>คำอธิบายสี:</span>
          <Legend color="bg-yellow-400" label="5-15%" />
          <Legend color="bg-orange-500" label="15-30%" />
          <Legend color="bg-red-500" label="&gt; 30%" />
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2 md:justify-end">
        <div className="relative w-full md:w-auto">
          <input
            type="text"
            placeholder="ค้นหาในตาราง..."
            className="w-full rounded-md border border-slate-300 py-2 pl-3 pr-10 text-sm md:w-64"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
          <span className="pointer-events-none absolute inset-y-0 right-2 flex items-center text-slate-400">
            🔍
          </span>
        </div>

        <button
          type="button"
          className="flex items-center gap-2 rounded-md border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
          onClick={onExport}
          disabled={!applied || filteredCount === 0 || isExporting}
        >
          {isExporting ? "กำลังส่งออก…" : "Export"}
        </button>

        {!isMobile && (
          <div className="hidden items-center gap-3 text-sm md:flex">
            <span className="text-slate-600">ช่วงข้อมูล:</span>
            <div className="inline-flex overflow-hidden rounded-md border border-slate-300">
              {MONTH_OPTIONS.map((option, index) => {
                const isActive = historyMonths === option;
                return (
                  <button
                    key={option}
                    type="button"
                    onClick={() => setHistoryMonths(option)}
                    className={`px-3 py-1 text-sm font-medium ${
                      isActive
                        ? "bg-blue-50 text-blue-700"
                        : "bg-white text-slate-600 hover:bg-slate-50"
                    } ${index !== MONTH_OPTIONS.length - 1 ? "border-r border-slate-300" : ""}`}
                  >
                    {option} เดือน
                  </button>
                );
              })}
            </div>
          </div>
        )}

        <div className="flex items-center gap-2 text-sm">
          <label className="text-slate-600">แสดง:</label>
          <select
            className="rounded-md border border-slate-300 p-1"
            value={pageSize}
            onChange={(event) => setPageSize(Number(event.target.value))}
          >
            <option value={10}>10</option>
            <option value={25}>25</option>
            <option value={50}>50</option>
          </select>
        </div>
      </div>
    </div>
  );
}
