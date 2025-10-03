import { useEffect, useMemo, useState } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { useQueries, useQuery } from "@tanstack/react-query";
import { getBranches } from "../api/branches";
import { getCustCodes } from "../api/custcodes";
import { getDetails } from "../api/details";
import type { DetailItem } from "../api/details";
import { useAuth } from "../lib/auth";
import { useMediaQuery } from "../lib/useMediaQuery";
import { useToast } from "../lib/useToast";
import { FilterSection } from "../components/DetailPage/FilterSection";
import { ResultsHeader } from "../components/DetailPage/ResultsHeader";
import { DataTable } from "../components/DetailPage/DataTable";
import {
  EmptyState,
  WarningState,
  ErrorState,
  LoadingState,
  Pager,
} from "../components/DetailPage/UIComponents";
import {
  DEFAULT_THRESHOLD,
  MAX_HISTORY_MONTHS,
  MONTH_OPTIONS,
  STORAGE_KEYS,
  defaultLatestYm,
  loadNumber,
  persistNumber,
  coerceHistoryMonths,
  buildMonths,
  combineRows,
  filterRows,
  fmtThMonth,
  sanitizeFileNamePart,
} from "../components/DetailPage/utils";

type AppliedFilters = {
  branch: string;
  ym: string;
};

type KnownError = Error & { status?: number };

export default function DetailPage() {
  const navigate = useNavigate();
  const { user, hydrated, logout: signOut } = useAuth();
  const { showToast } = useToast();
  const [branch, setBranch] = useState("");
  const [latestYm, setLatestYm] = useState(() => defaultLatestYm());
  const [threshold, setThreshold] = useState(() =>
    loadNumber(STORAGE_KEYS.threshold, DEFAULT_THRESHOLD),
  );
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [applied, setApplied] = useState<AppliedFilters | null>(null);
  const [historyMonths, setHistoryMonths] = useState(() =>
    coerceHistoryMonths(
      loadNumber(STORAGE_KEYS.months, MONTH_OPTIONS[1]),
    ),
  );
  const isMobile = useMediaQuery("(max-width: 767px)");
  const effectiveMonths = isMobile ? 3 : historyMonths;
  const isAuthenticated = hydrated && Boolean(user);
  const [isExporting, setIsExporting] = useState(false);

  useEffect(() => {
    persistNumber(STORAGE_KEYS.threshold, threshold);
  }, [threshold]);

  useEffect(() => {
    persistNumber(STORAGE_KEYS.months, historyMonths);
  }, [historyMonths]);

  useEffect(() => {
    setPage(1);
  }, [threshold, search, pageSize, effectiveMonths]);

  const branchesQuery = useQuery({
    queryKey: ["branches"],
    queryFn: () => getBranches({ limit: 1000 }),
    enabled: isAuthenticated,
  });
  const branches = useMemo(() => {
    const items = branchesQuery.data?.items ?? [];
    if (import.meta.env.DEV) {
      console.log('Branches query status:', branchesQuery.status);
      console.log('Branches data:', branchesQuery.data);
      console.log('Branches items:', items);
    }
    return items;
  }, [branchesQuery.data?.items, branchesQuery.status]);
  const defaultBranch = useMemo(() => {
    if (!user?.ba) return null;
    const matched = branches.find((branchItem) => branchItem.code === user.ba);
    return matched ? matched.code : null;
  }, [branches, user?.ba]);

  useEffect(() => {
    if (branch || !defaultBranch) return;
    setBranch(defaultBranch);
  }, [defaultBranch, branch]);

  const custcodesQuery = useQuery({
    queryKey: ["custcodes", applied?.branch, applied?.ym],
    queryFn: () =>
      getCustCodes({ branch: applied!.branch, ym: applied!.ym, limit: 200 }),
    enabled: Boolean(applied) && isAuthenticated,
  });

  const monthsAll = useMemo(
    () => (applied ? buildMonths(applied.ym, MAX_HISTORY_MONTHS) : []),
    [applied],
  );
  const exportMonthLabels = useMemo(() => monthsAll.map((ym) => fmtThMonth(ym)), [monthsAll]);

  const detailQueries = useQueries({
    queries: monthsAll.map((ym) => ({
      queryKey: ["details", applied?.branch, ym],
      queryFn: () =>
        getDetails({
          branch: applied!.branch,
          ym,
          limit: 200,
          order_by: "present_water_usg",
          sort: "DESC",
        }),
      enabled: Boolean(applied) && isAuthenticated,
      select: (data: { items: DetailItem[] }) => data.items,
    })),
  });

  const detailItems = useMemo(
    () => detailQueries.map((q) => q.data ?? []),
    [detailQueries],
  );
  const detailErrorResult = detailQueries.find((q) => q.error != null);
  const detailsError =
    detailErrorResult && detailErrorResult.error instanceof Error
      ? (detailErrorResult.error as KnownError)
      : undefined;

  const rows = useMemo(
    () =>
      applied
        ? combineRows(monthsAll, detailItems, custcodesQuery.data?.items ?? [])
        : [],
    [applied, monthsAll, detailItems, custcodesQuery.data?.items],
  );

  const prevMonth = monthsAll[1];
  const filteredRows = useMemo(
    () => filterRows(rows, applied?.ym, prevMonth, threshold, search),
    [rows, applied?.ym, prevMonth, threshold, search],
  );

  const totalPages = Math.max(1, Math.ceil(filteredRows.length / pageSize));
  const pageRows = filteredRows.slice(
    (page - 1) * pageSize,
    (page - 1) * pageSize + pageSize,
  );
  const monthsToDisplay = useMemo(
    () => monthsAll.slice(0, effectiveMonths),
    [monthsAll, effectiveMonths],
  );

  const isLoading =
    Boolean(applied) &&
    isAuthenticated &&
    (custcodesQuery.isLoading || detailQueries.some((q) => q.isLoading));
  const isFetching =
    Boolean(applied) &&
    isAuthenticated &&
    (custcodesQuery.isFetching || detailQueries.some((q) => q.isFetching));
  const custcodesError =
    custcodesQuery.error instanceof Error
      ? (custcodesQuery.error as KnownError)
      : undefined;
  const error = detailsError;
  const warning = !error && custcodesError ? custcodesError : undefined;

  function handleApply() {
    if (!branch) return;
    setApplied({ branch, ym: latestYm });
    setPage(1);
  }

  function handleReset() {
    setBranch("");
    const nextYm = defaultLatestYm();
    setLatestYm(nextYm);
    setThreshold(DEFAULT_THRESHOLD);
    setHistoryMonths(MONTH_OPTIONS[1]);
    setSearch("");
    setPage(1);
    setPageSize(10);
    setApplied(null);
  }

  function handleSignOut() {
    signOut();
    navigate("/", { replace: true });
  }

  async function handleExport() {
    if (isExporting) return;
    if (!applied) return;
    if (!filteredRows.length) return;
    if (!monthsAll.length) return;

    setIsExporting(true);

    try {
      const branchPart = applied.branch
        ? sanitizeFileNamePart(applied.branch)
        : "all";
      const fileName = `big-meter-${branchPart}-${applied.ym}.xlsx`;

      const { exportDetailsToXlsx } = await import("../lib/exportDetailsXlsx");
      await exportDetailsToXlsx({
        rows: filteredRows,
        months: monthsAll,
        monthLabels: exportMonthLabels,
        fileName,
      });
    } catch (error) {
      // Only log in development
      if (import.meta.env.DEV) {
        console.error("Export failed", error);
      }
      showToast("ไม่สามารถส่งออกไฟล์ได้ กรุณาลองอีกครั้ง", "error");
    } finally {
      setIsExporting(false);
    }
  }

  if (!hydrated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-100 text-slate-600">
        กำลังโหลด…
      </div>
    );
  }

  if (!user) {
    return <Navigate to="/" replace />;
  }

  return (
    <div className="min-h-screen bg-gray-50 px-3 py-4 md:px-6 md:py-8">
      <div className="mx-auto flex w-full max-w-[1400px] flex-col gap-8">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-slate-800 md:text-3xl">
              ระบบแสดงผลผู้ใช้น้ำรายใหญ่
            </h1>
            <p className="mt-1 text-sm text-slate-500 md:text-base">
              แดชบอร์ดสรุปข้อมูลการใช้น้ำ
            </p>
          </div>
          <div className="flex items-center gap-3">
            <div className="text-right text-sm leading-tight">
              <div className="font-semibold text-slate-700">{user.firstname}</div>
              <div className="text-xs text-slate-500">{user.username}</div>
            </div>
            <button
              type="button"
              onClick={handleSignOut}
              className="rounded-md border border-slate-300 px-3 py-1 text-sm font-medium text-slate-600 transition hover:bg-slate-100"
            >
              ออกจากระบบ
            </button>
          </div>
        </header>

        <FilterSection
          branch={branch}
          setBranch={setBranch}
          latestYm={latestYm}
          setLatestYm={setLatestYm}
          threshold={threshold}
          setThreshold={setThreshold}
          branches={branches}
          onApply={handleApply}
          onReset={handleReset}
        />

        <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <ResultsHeader
            filteredCount={filteredRows.length}
            search={search}
            setSearch={setSearch}
            pageSize={pageSize}
            setPageSize={setPageSize}
            historyMonths={historyMonths}
            setHistoryMonths={setHistoryMonths}
            onExport={handleExport}
            isExporting={isExporting}
            applied={!!applied}
            isMobile={isMobile}
          />

          {!applied && (
            <EmptyState message="เลือกเดือนและสาขา แล้วกดแสดงรายงานเพื่อดูข้อมูล" />
          )}

          {applied && warning && <WarningState warning={warning} />}
          {applied && error && <ErrorState error={error} />}

          {applied && !error && (
            <div className="mt-6">
              {isLoading ? (
                <LoadingState />
              ) : (
                <>
                  <DataTable
                    rows={pageRows}
                    months={monthsToDisplay}
                    latestYm={applied.ym}
                    baseIndex={(page - 1) * pageSize}
                    isMobile={isMobile}
                  />
                  {filteredRows.length === 0 && (
                    <div className="mt-4 rounded-md border border-dashed border-slate-200 p-8 text-center text-sm text-slate-500">
                      ไม่พบข้อมูลที่เข้าเงื่อนไขตามตัวกรองที่เลือก
                    </div>
                  )}
                </>
              )}

              <Pager page={page} totalPages={totalPages} onChange={setPage} />

              {isFetching && !isLoading && (
                <div className="mt-2 text-xs text-slate-500">
                  กำลังอัปเดตข้อมูลล่าสุด…
                </div>
              )}
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
