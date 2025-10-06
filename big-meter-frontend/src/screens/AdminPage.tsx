import { useState } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "../lib/auth";
import { useIsAdmin } from "../lib/useIsAdmin";
import { getBranches } from "../api/branches";
import {
  triggerYearlyInit,
  triggerMonthlySync,
  type YearlyInitResponse,
  type MonthlySyncResponse,
} from "../api/sync";
import { getSyncLogs } from "../api/syncLogs";

export default function AdminPage() {
  const navigate = useNavigate();
  const { user, hydrated, logout } = useAuth();
  const isAdmin = useIsAdmin();
  const isAuthenticated = hydrated && Boolean(user);

  // Yearly Init Form State
  const [yearlyBranches, setYearlyBranches] = useState<string[]>([]);
  const [debtYm, setDebtYm] = useState("");
  const [yearlyLoading, setYearlyLoading] = useState(false);
  const [yearlyResult, setYearlyResult] = useState<YearlyInitResponse | null>(
    null,
  );
  const [yearlyError, setYearlyError] = useState<string | null>(null);

  // Monthly Sync Form State
  const [monthlyBranches, setMonthlyBranches] = useState<string[]>([]);
  const [monthlyYm, setMonthlyYm] = useState("");
  const [monthlyLoading, setMonthlyLoading] = useState(false);
  const [monthlyResult, setMonthlySyncResponse] =
    useState<MonthlySyncResponse | null>(null);
  const [monthlyError, setMonthlyError] = useState<string | null>(null);

  // Fetch branches
  const branchesQuery = useQuery({
    queryKey: ["branches"],
    queryFn: () => getBranches({ limit: 1000 }),
    enabled: isAuthenticated && isAdmin,
  });

  const branches = branchesQuery.data?.items ?? [];

  // Fetch sync logs
  const syncLogsQuery = useQuery({
    queryKey: ["syncLogs"],
    queryFn: () => getSyncLogs({ limit: 20 }),
    enabled: isAuthenticated && isAdmin,
    refetchInterval: 10000, // Refresh every 10 seconds
  });

  const syncLogs = syncLogsQuery.data?.items ?? [];

  // Redirect non-authenticated users
  if (hydrated && !isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  // Redirect non-admin users
  if (hydrated && isAuthenticated && !isAdmin) {
    return <Navigate to="/details" replace />;
  }

  const handleYearlySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (yearlyBranches.length === 0 || !debtYm) return;

    setYearlyLoading(true);
    setYearlyError(null);
    setYearlyResult(null);

    try {
      const result = await triggerYearlyInit({
        branches: yearlyBranches,
        debt_ym: debtYm,
      });
      setYearlyResult(result);
      // Refetch sync logs to show the new entry
      syncLogsQuery.refetch();
    } catch (error) {
      setYearlyError(
        error instanceof Error ? error.message : "เกิดข้อผิดพลาดที่ไม่ทราบสาเหตุ",
      );
    } finally {
      setYearlyLoading(false);
    }
  };

  const handleMonthlySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (monthlyBranches.length === 0 || !monthlyYm) return;

    setMonthlyLoading(true);
    setMonthlyError(null);
    setMonthlySyncResponse(null);

    try {
      const result = await triggerMonthlySync({
        branches: monthlyBranches,
        ym: monthlyYm,
      });
      setMonthlySyncResponse(result);
      // Refetch sync logs to show the new entry
      syncLogsQuery.refetch();
    } catch (error) {
      setMonthlyError(
        error instanceof Error ? error.message : "เกิดข้อผิดพลาดที่ไม่ทราบสาเหตุ",
      );
    } finally {
      setMonthlyLoading(false);
    }
  };

  const handleLogout = () => {
    logout();
    navigate("/");
  };

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-2xl font-bold text-slate-900">
                หน้าจัดการระบบ (Admin)
              </h1>
              {user && (
                <p className="text-sm text-slate-600 mt-1">
                  ผู้ใช้: {user.firstname} {user.lastname} ({user.username})
                </p>
              )}
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => navigate("/details")}
                className="px-4 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50"
              >
                กลับไปหน้ารายงาน
              </button>
              <button
                onClick={handleLogout}
                className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700"
              >
                ออกจากระบบ
              </button>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Yearly Init Section */}
          <section className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <h2 className="text-xl font-semibold text-slate-900 mb-4">
              🔄 Yearly Initialization Sync
            </h2>
            <p className="text-sm text-slate-600 mb-6">
              เริ่มต้นข้อมูลรายปี (Top-200 Customers)
            </p>

            <form onSubmit={handleYearlySubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  สาขา (Branches)
                </label>
                <select
                  multiple
                  value={yearlyBranches}
                  onChange={(e) =>
                    setYearlyBranches(
                      Array.from(e.target.selectedOptions, (opt) => opt.value),
                    )
                  }
                  className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-32"
                  disabled={yearlyLoading}
                >
                  {branches.map((branch) => (
                    <option key={branch.code} value={branch.code}>
                      {branch.code} - {branch.name || "N/A"}
                    </option>
                  ))}
                </select>
                <p className="text-xs text-slate-500 mt-1">
                  กด Cmd/Ctrl เพื่อเลือกหลายสาขา
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  Debt Year-Month (YYYYMM)
                </label>
                <input
                  type="text"
                  value={debtYm}
                  onChange={(e) => setDebtYm(e.target.value)}
                  placeholder="202410"
                  pattern="\d{6}"
                  className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  disabled={yearlyLoading}
                  required
                />
                <p className="text-xs text-slate-500 mt-1">
                  ตัวอย่าง: 202410 (ตุลาคม 2024)
                </p>
              </div>

              <button
                type="submit"
                disabled={
                  yearlyLoading || yearlyBranches.length === 0 || !debtYm
                }
                className="w-full px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:bg-slate-300 disabled:cursor-not-allowed"
              >
                {yearlyLoading ? "กำลังดำเนินการ..." : "เริ่มต้น Yearly Sync"}
              </button>
            </form>

            {yearlyError && (
              <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-md">
                <p className="text-sm text-red-800">
                  <strong>ข้อผิดพลาด:</strong> {yearlyError}
                </p>
              </div>
            )}

            {yearlyResult && (
              <div className="mt-4 p-4 bg-green-50 border border-green-200 rounded-md">
                <p className="text-sm font-semibold text-green-900 mb-2">
                  ✓ สำเร็จ
                </p>
                <div className="text-sm text-green-800 space-y-1">
                  <p>
                    <strong>Fiscal Year:</strong> {yearlyResult.fiscal_year}
                  </p>
                  <p>
                    <strong>Branches:</strong> {yearlyResult.branches.join(", ")}
                  </p>
                  <p>
                    <strong>Started:</strong>{" "}
                    {new Date(yearlyResult.started_at).toLocaleString("th-TH")}
                  </p>
                  <p>
                    <strong>Finished:</strong>{" "}
                    {new Date(yearlyResult.finished_at).toLocaleString("th-TH")}
                  </p>
                  {yearlyResult.note && (
                    <p className="text-xs text-green-700 mt-2">
                      <em>Note: {yearlyResult.note}</em>
                    </p>
                  )}
                </div>
              </div>
            )}
          </section>

          {/* Monthly Sync Section */}
          <section className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <h2 className="text-xl font-semibold text-slate-900 mb-4">
              📅 Monthly Sync
            </h2>
            <p className="text-sm text-slate-600 mb-6">
              ซิงค์ข้อมูลรายเดือน (Usage Details)
            </p>

            <form onSubmit={handleMonthlySubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  สาขา (Branches)
                </label>
                <select
                  multiple
                  value={monthlyBranches}
                  onChange={(e) =>
                    setMonthlyBranches(
                      Array.from(e.target.selectedOptions, (opt) => opt.value),
                    )
                  }
                  className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-32"
                  disabled={monthlyLoading}
                >
                  {branches.map((branch) => (
                    <option key={branch.code} value={branch.code}>
                      {branch.code} - {branch.name || "N/A"}
                    </option>
                  ))}
                </select>
                <p className="text-xs text-slate-500 mt-1">
                  กด Cmd/Ctrl เพื่อเลือกหลายสาขา
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  Year-Month (YYYYMM)
                </label>
                <input
                  type="text"
                  value={monthlyYm}
                  onChange={(e) => setMonthlyYm(e.target.value)}
                  placeholder="202410"
                  pattern="\d{6}"
                  className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  disabled={monthlyLoading}
                  required
                />
                <p className="text-xs text-slate-500 mt-1">
                  ตัวอย่าง: 202410 (ตุลาคม 2024)
                </p>
              </div>

              <button
                type="submit"
                disabled={
                  monthlyLoading || monthlyBranches.length === 0 || !monthlyYm
                }
                className="w-full px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:bg-slate-300 disabled:cursor-not-allowed"
              >
                {monthlyLoading ? "กำลังดำเนินการ..." : "เริ่มต้น Monthly Sync"}
              </button>
            </form>

            {monthlyError && (
              <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-md">
                <p className="text-sm text-red-800">
                  <strong>ข้อผิดพลาด:</strong> {monthlyError}
                </p>
              </div>
            )}

            {monthlyResult && (
              <div className="mt-4 p-4 bg-green-50 border border-green-200 rounded-md">
                <p className="text-sm font-semibold text-green-900 mb-2">
                  ✓ สำเร็จ
                </p>
                <div className="text-sm text-green-800 space-y-1">
                  <p>
                    <strong>Year-Month:</strong> {monthlyResult.ym}
                  </p>
                  <p>
                    <strong>Branches:</strong> {monthlyResult.branches.join(", ")}
                  </p>
                  <p>
                    <strong>Started:</strong>{" "}
                    {new Date(monthlyResult.started_at).toLocaleString("th-TH")}
                  </p>
                  <p>
                    <strong>Finished:</strong>{" "}
                    {new Date(monthlyResult.finished_at).toLocaleString("th-TH")}
                  </p>
                  {monthlyResult.note && (
                    <p className="text-xs text-green-700 mt-2">
                      <em>Note: {monthlyResult.note}</em>
                    </p>
                  )}
                </div>
              </div>
            )}
          </section>
        </div>

        {/* Sync Logs Section */}
        <div className="mt-8">
          <h2 className="text-xl font-semibold text-slate-900 mb-4">
            📋 Sync Operation Logs
          </h2>
          <div className="bg-white rounded-lg shadow-sm border border-slate-200 overflow-hidden">
            {syncLogsQuery.isLoading ? (
              <div className="p-8 text-center text-slate-500">
                กำลังโหลด...
              </div>
            ) : syncLogsQuery.isError ? (
              <div className="p-8 text-center text-red-600">
                ไม่สามารถโหลดข้อมูล log ได้
              </div>
            ) : syncLogs.length === 0 ? (
              <div className="p-8 text-center text-slate-500">
                ยังไม่มีประวัติการ sync
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead className="bg-slate-50 border-b border-slate-200">
                    <tr>
                      <th className="px-4 py-3 text-left font-medium text-slate-700">
                        เวลา
                      </th>
                      <th className="px-4 py-3 text-left font-medium text-slate-700">
                        ประเภท
                      </th>
                      <th className="px-4 py-3 text-left font-medium text-slate-700">
                        สาขา
                      </th>
                      <th className="px-4 py-3 text-left font-medium text-slate-700">
                        สถานะ
                      </th>
                      <th className="px-4 py-3 text-right font-medium text-slate-700">
                        Upserted
                      </th>
                      <th className="px-4 py-3 text-right font-medium text-slate-700">
                        Zeroed
                      </th>
                      <th className="px-4 py-3 text-right font-medium text-slate-700">
                        ระยะเวลา
                      </th>
                      <th className="px-4 py-3 text-left font-medium text-slate-700">
                        Triggered By
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-200">
                    {syncLogs.map((log) => (
                      <tr
                        key={log.id}
                        className="hover:bg-slate-50 transition-colors"
                      >
                        <td className="px-4 py-3 text-slate-600">
                          {new Date(log.started_at).toLocaleString("th-TH", {
                            dateStyle: "short",
                            timeStyle: "short",
                          })}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${
                              log.sync_type === "yearly_init"
                                ? "bg-purple-100 text-purple-700"
                                : "bg-blue-100 text-blue-700"
                            }`}
                          >
                            {log.sync_type === "yearly_init"
                              ? "Yearly Init"
                              : "Monthly Sync"}
                          </span>
                        </td>
                        <td className="px-4 py-3 font-mono text-slate-700">
                          {log.branch_code}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${
                              log.status === "success"
                                ? "bg-green-100 text-green-700"
                                : log.status === "error"
                                  ? "bg-red-100 text-red-700"
                                  : "bg-yellow-100 text-yellow-700"
                            }`}
                          >
                            {log.status === "success"
                              ? "✓ Success"
                              : log.status === "error"
                                ? "✗ Error"
                                : "⋯ In Progress"}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right font-mono text-slate-700">
                          {log.records_upserted?.toLocaleString() ?? "-"}
                        </td>
                        <td className="px-4 py-3 text-right font-mono text-slate-700">
                          {log.records_zeroed?.toLocaleString() ?? "-"}
                        </td>
                        <td className="px-4 py-3 text-right text-slate-600">
                          {log.duration_ms
                            ? `${(log.duration_ms / 1000).toFixed(1)}s`
                            : "-"}
                        </td>
                        <td className="px-4 py-3 text-slate-600">
                          {log.triggered_by}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>

        {/* Info Section */}
        <div className="mt-8 p-6 bg-blue-50 border border-blue-200 rounded-lg">
          <h3 className="text-sm font-semibold text-blue-900 mb-2">
            ℹ️ คำแนะนำ
          </h3>
          <ul className="text-sm text-blue-800 space-y-1 list-disc list-inside">
            <li>
              <strong>Yearly Init:</strong> ใช้เมื่อต้องการเริ่มต้นข้อมูลรายปี
              (Top-200 customers)
            </li>
            <li>
              <strong>Monthly Sync:</strong> ใช้เมื่อต้องการซิงค์ข้อมูลการใช้น้ำรายเดือน
            </li>
            <li>
              สามารถเลือกหลายสาขาพร้อมกันได้ โดยกด Cmd (Mac) หรือ Ctrl (Windows)
              ค้างไว้
            </li>
            <li>
              รูปแบบ Year-Month ต้องเป็น 6 หลัก เช่น 202410 (ตุลาคม 2024)
            </li>
          </ul>
        </div>
      </main>
    </div>
  );
}
