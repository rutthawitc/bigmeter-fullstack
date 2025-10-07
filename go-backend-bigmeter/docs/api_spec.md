API Spec: BigMeter Sync (REST v1)

หมายเหตุ: เอกสารย่อสำหรับ Frontend อยู่ที่ `docs/API-Spec.md` (ตัวอย่างการเรียกใช้งานครบถ้วน)

พื้นฐาน

- Base URL: `/api/v1`
- Content-Type: `application/json; charset=utf-8`
- Auth (ทางเลือก): `Authorization: Bearer <JWT>` หรือ `X-API-Key: <key>`
- Pagination: `limit` (เริ่มที่ 50, สูงสุด 500), `offset` (เริ่มที่ 0)
- เวลา/เขตเวลา: ทุกเวลาจัดเก็บเป็น UTC; แสดงตามค่าโครงสร้างพื้นฐานของระบบ

สถานะปัจจุบัน (2025-09)

- ระบบยังไม่เปิดใช้การยืนยันตัวตน/สิทธิ์ (ไม่มี Auth) — API เปิดอ่านข้อมูลสาธารณะในโหมดพัฒนา
- Endpoints ที่ใช้งานได้: `GET /healthz`, `GET /version`, `GET /branches`, `GET /custcodes`, `GET /details`, `GET /details/summary`, `GET /custcodes/{cust_code}/details`, `POST /sync/init`, `POST /sync/monthly`, `GET /config`
- หมายเหตุ: `/sync/*` และ `/config` เป็น "stub" สำหรับการพัฒนา Frontend (ตอบกลับโครงสร้างข้อมูล แต่ไม่ได้รันงานจริง)

1) Health & Version

- GET `/healthz`
  - 200 OK
  - ตัวอย่างตอบกลับ:
    {
      "status": "ok",
      "time": "2025-10-01T08:00:00Z"
    }

- GET `/version`
  - 200 OK
  - ตัวอย่างตอบกลับ:
    {
      "service": "bigmeter-sync-api",
      "version": "1.0.0",
      "commit": "<gitsha>"
    }

2) Branches

- GET `/branches`
  - อธิบาย: รายการสาขาที่ระบบรู้จัก (จากคอนฟิก/CSV)
  - คิวรี: `q` (ค้นหาบางส่วน, ทางเลือก), `limit`, `offset`
  - 200 OK
  - ตัวอย่างตอบกลับ:
    {
      "items": [ {"code": "BA01"}, {"code": "BA02"} ],
      "total": 2,
      "limit": 50,
      "offset": 0
    }

3) Yearly Snapshot (Top-200)

- GET `/custcodes`
  - อธิบาย: ดึงรายชื่อ cust_code 200 ราย (ต่อสาขา) สำหรับปีงบประมาณที่กำหนด
  - คิวรี (อย่างน้อยอย่างใดอย่างหนึ่ง):
    - `branch` (จำเป็น)
    - `fiscal_year` (แนะนำ) หรือ `ym` (ระบบแปลงเป็นปีงบฯ ให้อัตโนมัติ)
    - `q` (ค้นหาบางส่วน across: `cust_code, meter_no, use_type, org_name, use_name, cust_name, address, route_code, meter_size, meter_brand, meter_state, debt_ym`)
    - `limit`, `offset`, `order_by` (allowlist: `cust_code, meter_no, use_type, created_at, org_name, use_name, cust_name, address, route_code, meter_size, meter_brand, meter_state, debt_ym`), `sort` (`asc|desc`)
  - 200 OK
  - ตัวอย่างตอบกลับ:
    {
      "items": [
        {
          "fiscal_year": 2025,
          "branch_code": "BA01",
          "org_name": "BA01",
          "cust_code": "C12345",
          "use_type": "R",
          "use_name": "Residential",
          "cust_name": "John Doe",
          "address": "...",
          "route_code": "RT01",
          "meter_no": "M-0001",
          "meter_size": "1/2",
          "meter_brand": "XYZ",
          "meter_state": "N",
          "debt_ym": "202410",
          "created_at": "2024-10-15T22:05:02Z"
        }
      ],
      "total": 200,
      "limit": 50,
      "offset": 0
    }

4) Monthly Details (รายเดือน)

- GET `/details`
  - อธิบาย: ดึงข้อมูลรายเดือนจาก `bm_meter_details`
  - คิวรี: `ym` (จำเป็น), `branch` (จำเป็น), `cust_code` (รายการ/ซ้ำได้), `q` (ค้นหาข้อความ across: `cust_code, meter_no, cust_name, address, route_code, org_name, use_type, use_name`), `limit`, `offset`, `order_by` (allowlist: `cust_code, present_water_usg, present_meter_count, average, created_at, org_name, use_type, use_name, cust_name, address, route_code, meter_no, meter_size, meter_brand, meter_state, debt_ym`), `sort`
  - 200 OK
  - หมายเหตุ: แถวที่เกิดจาก “ไม่มีข้อมูลจริง” จะมีค่าเลขเป็น 0 และฟิลด์ข้อความจำนวนมากอาจไม่มีค่า (ละเว้นจาก JSON) — ควรตีความว่าเป็น “zeroed row”
  - ตัวอย่างตอบกลับ:
    {
      "items": [
        {
          "year_month": "202410",
          "branch_code": "BA01",
          "org_name": "BA01",
          "cust_code": "C12345",
          "use_type": "R",
          "use_name": "Residential",
          "cust_name": "John Doe",
          "address": "...",
          "route_code": "RT01",
          "meter_no": "M-0001",
          "meter_size": "1/2",
          "meter_brand": "XYZ",
          "meter_state": "N",
          "average": 12.5,
          "present_meter_count": 300,
          "present_water_usg": 15.0,
          "debt_ym": "202410",
          "created_at": "2024-10-16T08:05:02Z",
          "is_zeroed": false
        },
        {
          "year_month": "202410",
          "branch_code": "BA01",
          "org_name": null,
          "cust_code": "C99999",
          "use_type": "R",
          "use_name": null,
          "cust_name": null,
          "address": null,
          "route_code": null,
          "meter_no": "M-0999",
          "meter_size": null,
          "meter_brand": null,
          "meter_state": "N",
          "average": 0,
          "present_meter_count": 0,
          "present_water_usg": 0,
          "debt_ym": "202410",
          "created_at": "2024-10-16T08:05:02Z",
          "is_zeroed": true
        }
      ],
      "total": 200,
      "limit": 50,
      "offset": 0
    }

- GET `/custcodes/{cust_code}/details`
  - อธิบาย: ดูประวัติรายเดือนสำหรับลูกค้ารายหนึ่งในช่วงเดือน
  - คิวรี: `branch` (จำเป็น), `from` (YYYYMM, จำเป็น), `to` (YYYYMM, จำเป็น)
  - 200 OK
  - ตัวอย่างตอบกลับ:
    {
      "cust_code": "C12345",
      "branch_code": "BA01",
      "from": "202410",
      "to": "202503",
      "series": [
        {"ym": "202410", "present_water_usg": 15.0, "is_zeroed": false},
        {"ym": "202411", "present_water_usg": 0.0,  "is_zeroed": true}
      ]
    }

- GET `/details/summary`
  - อธิบาย: สรุปภาพรวมรายเดือนในสาขา
  - คิวรี: `ym` (จำเป็น), `branch` (จำเป็น)
  - คำนิยาม `is_zeroed`: แถวที่ `present_water_usg == 0` และ `present_meter_count == 0` และ `org_name` ว่างหรือไม่มีค่า
  - 200 OK
  - ตัวอย่างตอบกลับ:
    {
      "ym": "202410",
      "branch": "BA01",
      "total": 200,
      "zeroed": 15,
      "active": 185,
      "sum_present_water_usg": 12345.67
    }

5) Sync Triggers (Admin)

- สถานะ: Stub (ตอบกลับตัวอย่าง/ไม่รันงานจริง) — ใช้เพื่อเชื่อมต่อ Frontend ระหว่างพัฒนา

- POST `/sync/init`
  - อธิบาย: สั่งรัน Yearly Init แบบครั้งเดียว (upsert ชุด 200 ต่อสาขา)
  - Request body:
    {
      "branches": ["BA01", "BA02"],
      "debt_ym": "202410"  // ทางเลือก (ดีฟอลต์ ต.ค. ของปีปัจจุบัน)
    }
  - 202 Accepted | 200 OK (ถ้าเลือก synchronous)
  - ตัวอย่างตอบกลับ (sync):
    {
      "fiscal_year": 2025,
      "branches": ["BA01"],
      "debt_ym": "202410",
      "stats": {"upserted": 200},
      "started_at": "2024-10-15T22:00:01Z",
      "finished_at": "2024-10-15T22:05:20Z"
    }
  - ตัวอย่าง curl (วางแผน):
    curl -X POST -H "Content-Type: application/json" \
      -d '{"branches":["BA01"],"debt_ym":"202410"}' \
      http://localhost:8089/api/v1/sync/init

- POST `/sync/monthly`
  - อธิบาย: สั่งรัน Monthly Details แบบครั้งเดียว
  - Request body:
    {
      "branches": ["BA01", "BA02"],
      "ym": "202410"
    }
  - 202 Accepted | 200 OK (ถ้าเลือก synchronous)
  - ตัวอย่างตอบกลับ (sync):
    {
      "ym": "202410",
      "branches": ["BA01"],
      "stats": {"upserted": 200, "zeroed": 15},
      "started_at": "2024-10-16T08:00:01Z",
      "finished_at": "2024-10-16T08:08:40Z"
    }
  - ตัวอย่าง curl (วางแผน):
    curl -X POST -H "Content-Type: application/json" \
      -d '{"branches":["BA01"],"ym":"202410"}' \
      http://localhost:8089/api/v1/sync/monthly

6) Config Introspection

- สถานะ: Stub (ตอบกลับค่าคอนฟิกหลักจากโปรเซส)

- GET `/config`
  - อธิบาย: ดูค่า config สำคัญของบริการ (อ่านอย่างเดียว)
  - 200 OK
  - ตัวอย่างตอบกลับ:
    {
      "timezone": "Asia/Bangkok",
      "cron_yearly": "0 30 1 16 10 *",
      "cron_monthly": "0 0 8 16 * *",
      "branches_count": 34
    }
  - ตัวอย่าง curl (วางแผน):
    curl -s http://localhost:8089/api/v1/config

การจัดการ Error (ตัวอย่าง)

- 400 Bad Request: พารามิเตอร์ไม่ถูกต้อง
  {
    "error": "invalid ym format, expect YYYYMM"
  }
- 401 Unauthorized / 403 Forbidden: ไม่มีสิทธิ์เข้าถึง
  {
    "error": "unauthorized"
  }
- 404 Not Found: ไม่พบทรัพยากร
  {
    "error": "cust_code not found"
  }
- 500 Internal Server Error: เกิดข้อผิดพลาดภายในระบบ
  {
    "error": "internal error"
  }

หมายเหตุการออกแบบ

- `is_zeroed` เป็นฟิลด์คำนวณ (ไม่ได้เก็บจริงใน DB) เพื่อช่วย UI/รายงาน
- Endpoints `/sync/*` ควรปกป้องด้วยสิทธิ์ระดับ Admin และอาจทำงานแบบ async พร้อม job-id ได้ในอนาคต
- สามารถเพิ่ม Rate Limit และ ETag/Conditional Requests เพื่อประสิทธิภาพ
