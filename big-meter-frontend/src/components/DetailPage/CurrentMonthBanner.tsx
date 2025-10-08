/**
 * CurrentMonthBanner component
 * Displays a warning banner before the 16th of current month
 */

interface CurrentMonthBannerProps {
  onDismiss: () => void;
}

export function CurrentMonthBanner({ onDismiss }: CurrentMonthBannerProps) {
  // Get current month name in Thai
  const getCurrentMonthThaiName = () => {
    const today = new Date();
    const thaiMonths = [
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
    const thaiYear = today.getFullYear() + 543;
    const monthName = thaiMonths[today.getMonth()];
    return `${monthName} ${thaiYear}`;
  };

  return (
    <div className="rounded-lg border border-yellow-300 bg-yellow-50 px-4 py-3 shadow-sm">
      <div className="flex items-center justify-between gap-3">
        <div className="flex flex-1 items-center justify-center gap-2 text-sm text-yellow-800">
          <span className="text-base">⏰</span>
          <p>
            ข้อมูลเดือนปัจจุบัน จะนำเข้าและแสดงผลได้ในวันที่ 16{" "}
            {getCurrentMonthThaiName()}
          </p>
        </div>
        <button
          type="button"
          onClick={onDismiss}
          className="rounded p-1 text-yellow-600 transition hover:bg-yellow-100 hover:text-yellow-800"
          aria-label="ปิดข้อความแจ้งเตือน"
        >
          <svg
            className="h-4 w-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>
    </div>
  );
}
