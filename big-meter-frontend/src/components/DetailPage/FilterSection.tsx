/**
 * FilterSection component for DetailPage
 * Handles month/year selection, branch selection, threshold input, and apply/reset actions
 */

import type { BranchItem } from "../../api/branches";
import {
  TH_MONTHS,
  ymParts,
  partsToYm,
  formatBranchLabel,
  normalizeThreshold,
  yearOptions,
} from "./utils";

export interface FilterSectionProps {
  branch: string;
  setBranch: (value: string) => void;
  latestYm: string;
  setLatestYm: (value: string) => void;
  threshold: number;
  setThreshold: (value: number) => void;
  branches: BranchItem[];
  onApply: () => void;
  onReset: () => void;
}

export function FilterSection({
  branch,
  setBranch,
  latestYm,
  setLatestYm,
  threshold,
  setThreshold,
  branches,
  onApply,
  onReset,
}: FilterSectionProps) {
  const yearOptionsList = yearOptions();

  return (
    <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-xl font-semibold text-slate-700">ตัวกรองข้อมูล</h2>
      <div className="mt-4 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">
            เดือน/ปี
          </label>
          <div className="flex gap-2">
            <select
              className="w-full rounded-md border border-slate-300 p-2"
              value={Number(latestYm.slice(4, 6))}
              onChange={(event) =>
                setLatestYm(
                  partsToYm({
                    ...ymParts(latestYm),
                    m: Number(event.target.value),
                  })
                )
              }
            >
              {TH_MONTHS.map((name, index) => (
                <option key={name} value={index + 1}>
                  {name}
                </option>
              ))}
            </select>
            <select
              className="w-full rounded-md border border-slate-300 p-2"
              value={Number(latestYm.slice(0, 4))}
              onChange={(event) =>
                setLatestYm(
                  partsToYm({
                    ...ymParts(latestYm),
                    y: Number(event.target.value),
                  })
                )
              }
            >
              {yearOptionsList.map((year) => (
                <option key={year} value={year}>
                  {year + 543}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">
            สาขา
          </label>
          <select
            className="w-full rounded-md border border-slate-300 p-2"
            value={branch}
            onChange={(event) => setBranch(event.target.value)}
          >
            <option value="">เลือกสาขา</option>
            {branches.map((item) => (
              <option key={item.code} value={item.code}>
                {formatBranchLabel(item)}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">
            เปอร์เซ็นต์ผลต่าง (การใช้น้ำลดลง)
          </label>
          <div className="flex items-center gap-2">
            <input
              type="number"
              min={0}
              max={100}
              className="w-32 rounded-md border border-slate-300 p-2"
              value={threshold}
              onChange={(event) =>
                setThreshold(normalizeThreshold(event.target.value))
              }
            />
            <span className="text-slate-600">%</span>
          </div>
        </div>

        <div className="flex flex-col justify-end gap-2">
          <div className="flex gap-2">
            <button
              type="button"
              className="w-full rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
              onClick={onReset}
            >
              ล้างค่า
            </button>
            <button
              type="button"
              className="w-full rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-blue-300"
              onClick={onApply}
              disabled={!branch}
            >
              แสดงรายงาน
            </button>
          </div>
          <p className="text-xs text-slate-500">
            ข้อมูลจะโหลดเมื่อกดปุ่ม "แสดงรายงาน"
          </p>
        </div>
      </div>
    </section>
  );
}
