/**
 * Simple UI components for DetailPage
 */

export function Legend({ color, label }: { color: string; label: string }) {
  return (
    <span className="flex items-center gap-2">
      <span className={`h-3 w-3 rounded-full ${color}`} />
      <span>{label}</span>
    </span>
  );
}

export function EmptyState({ message }: { message: string }) {
  return (
    <div className="mt-6 rounded-md border border-dashed border-slate-200 p-8 text-center text-sm text-slate-500">
      {message}
    </div>
  );
}

export function LoadingState() {
  return (
    <div className="mt-6 flex flex-col items-center gap-3 rounded-md border border-slate-200 p-8 text-sm text-slate-600">
      <span className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-blue-200 border-t-blue-600" />
      <span>กำลังโหลดข้อมูล…</span>
    </div>
  );
}

type KnownError = Error & { status?: number };
type ErrorMessage = { main: string; detail?: string };

export function WarningState({ warning }: { warning: KnownError }) {
  const message = interpretError(warning);
  return (
    <div className="mt-6 rounded-md border border-yellow-200 bg-yellow-50 p-6 text-sm text-yellow-800">
      <p>{message.main}</p>
      {message.detail && (
        <p className="mt-2 text-xs opacity-80">
          รายละเอียดระบบ: {message.detail}
        </p>
      )}
    </div>
  );
}

export function ErrorState({ error }: { error: KnownError }) {
  const message = interpretError(error);
  return (
    <div className="mt-6 rounded-md border border-red-200 bg-red-50 p-6 text-sm text-red-700">
      <p>{message.main}</p>
      {message.detail && (
        <p className="mt-2 text-xs opacity-80">
          รายละเอียดระบบ: {message.detail}
        </p>
      )}
    </div>
  );
}

function interpretError(error: KnownError): ErrorMessage {
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

export function Pager({
  page,
  totalPages,
  onChange,
}: {
  page: number;
  totalPages: number;
  onChange: (value: number) => void;
}) {
  if (!totalPages) return null;
  const canPrev = page > 1;
  const canNext = page < totalPages;
  const pages: Array<number | "…"> = [];
  for (let i = 1; i <= totalPages; i++) {
    if (i === 1 || i === totalPages || Math.abs(i - page) <= 1) {
      pages.push(i);
    } else if (pages[pages.length - 1] !== "…") {
      pages.push("…");
    }
  }
  return (
    <div className="mt-6 flex flex-col items-center justify-between gap-4 md:flex-row">
      <div className="text-sm text-slate-500">
        หน้า {page} จาก {totalPages}
      </div>
      <div className="flex items-center gap-1">
        <button
          type="button"
          className="rounded-md px-3 py-2 text-sm text-slate-600 hover:bg-slate-100 disabled:text-slate-400"
          disabled={!canPrev}
          onClick={() => canPrev && onChange(page - 1)}
        >
          &laquo; ก่อนหน้า
        </button>
        {pages.map((value, index) =>
          value === "…" ? (
            <span key={`ellipsis-${index}`} className="px-2 text-slate-400">
              …
            </span>
          ) : (
            <button
              key={value}
              type="button"
              className={`h-8 w-8 rounded-md text-sm ${value === page ? "bg-blue-600 text-white" : "hover:bg-slate-100"}`}
              onClick={() => onChange(value)}
            >
              {value}
            </button>
          )
        )}
        <button
          type="button"
          className="rounded-md px-3 py-2 text-sm text-slate-600 hover:bg-slate-100 disabled:text-slate-400"
          disabled={!canNext}
          onClick={() => canNext && onChange(page + 1)}
        >
          ถัดไป &raquo;
        </button>
      </div>
    </div>
  );
}
