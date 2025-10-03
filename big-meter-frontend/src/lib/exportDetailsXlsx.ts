import { utils, writeFile } from "xlsx";

type ExportRow = {
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

type ExportOptions = {
  rows: ExportRow[];
  months: string[];
  monthLabels: string[];
  fileName: string;
};

export function exportDetailsToXlsx({
  rows,
  months,
  monthLabels,
  fileName,
}: ExportOptions) {
  if (!rows.length) return;

  const baseHeader = [
    "ลำดับ",
    "กปภ.สาขา",
    "เลขที่ผู้ใช้น้ำ",
    "ประเภท",
    "รายละเอียด",
    "ชื่อผู้ใช้น้ำ",
    "ที่อยู่",
    "เส้นทาง",
    "หมายเลขมาตร",
    "ขนาดมาตร",
    "ยี่ห้อ",
    "สถานะมาตร",
    "หน่วยน้ำเฉลี่ย",
    "เลขมาตรที่อ่านได้",
  ];

  const header = [...baseHeader, ...monthLabels];

  const body = rows.map((row, index) => {
    const branch = row.orgName ?? row.branchCode;
    const values = months.map((ym) => {
      const value = row.values[ym];
      return value == null ? null : Number(value);
    });

    return [
      index + 1,
      branch,
      row.custCode,
      row.useType ?? "",
      row.useName ?? "",
      row.custName ?? "",
      row.address ?? "",
      row.routeCode ?? "",
      row.meterNo ?? "",
      row.meterSize ?? "",
      row.meterBrand ?? "",
      row.meterState ?? "",
      row.average ?? null,
      row.presentMeterCount ?? null,
      ...values,
    ];
  });

  const worksheet = utils.aoa_to_sheet([header, ...body]);

  worksheet["!cols"] = header.map((cell) => ({
    wch: typeof cell === "string" ? Math.max(cell.length + 2, 12) : 12,
  }));

  const workbook = utils.book_new();
  utils.book_append_sheet(workbook, worksheet, "รายละเอียด");

  writeFile(workbook, fileName, { bookType: "xlsx" });
}
