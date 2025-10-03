คำอธิบายการทำงาน: BigMeter Sync (Oracle → PostgreSQL)

ภาพรวม

- บริการนี้ทำหน้าที่ “ซิงก์” ข้อมูลจาก Oracle มายัง PostgreSQL ตามกำหนดเวลา เพื่อเก็บข้อมูลผู้ใช้น้ำ/มาตรที่ต้องติดตาม (top-200 ต่อสาขา) และรายละเอียดการใช้น้ำรายเดือนของกลุ่มนั้น
- การซิงก์แบ่งเป็น 2 เฟสหลัก:
  1) เฟสเริ่มต้นประจำปี (Yearly Init): ดึงรายชื่อลูกค้า/มาตร 200 รายต่อสาขา เก็บไว้ที่ตาราง `bm_custcode_init`
  2) เฟสประจำเดือน (Monthly Details): ดึงรายละเอียดรายเดือนเฉพาะลูกค้าที่อยู่ใน `bm_custcode_init` ไปอัปเซิร์ตใน `bm_meter_details`

กำหนดเวลา (Scheduler)

- โซนเวลาอ้างอิงจากค่า `TIMEZONE` (ดีฟอลต์ `Asia/Bangkok`)
- งานประจำปี: ทุกวันที่ 15 ต.ค. เวลา 22:00 (สเปกดีฟอลต์ `0 0 22 15 10 *`) เรียกใช้การเก็บ 200 รายแรกของปีงบประมาณใหม่
- งานประจำเดือน: ทุกวันที่ 16 ของเดือน เวลา 08:00 (สเปกดีฟอลต์ `0 0 8 16 * *`) ดึงข้อมูลของเดือนก่อนหน้าและอัปเดตลง `bm_meter_details`

โหมดการรัน (ผ่านตัวแปรแวดล้อม `MODE`)

- ไม่กำหนด `MODE`: รันเป็น Scheduler ตาม Cron Spec
- `MODE=init-once`: รันงาน Yearly Init ครั้งเดียว (ใช้ปีงบประมาณตามเวลา ณ ขณะรัน และกำหนด `DEBT_YM` ดีฟอลต์เป็น ต.ค. ของปีปัจจุบัน)
- `MODE=month-once`: รันงาน Monthly ครั้งเดียว (กำหนดเดือนเป้าหมายผ่าน `YM` รูปแบบ `YYYYMM` หากไม่กำหนดจะใช้เดือนก่อนหน้าจากวันที่ปัจจุบัน)

ตารางข้อมูลใน PostgreSQL

- `bm_custcode_init`: เก็บรายชื่อ top-200 ต่อสาขา พร้อมข้อมูลประกอบ (เช่น ประเภทการใช้, หมายเลขมิเตอร์, สถานะมิเตอร์, `debt_ym`) โดยมีคีย์ยูนีก `(fiscal_year, branch_code, cust_code)` เพื่อป้องกันซ้ำและรองรับ upsert
- `bm_meter_details`: เก็บรายละเอียดรายเดือนของลูกค้าที่คัดเลือกมา โดยมีคีย์ยูนีก `(year_month, branch_code, cust_code)` รองรับ upsert เมื่อข้อมูลเปลี่ยน

แหล่ง SQL และรูปแบบการดึงข้อมูล

- เฟส Yearly Init ใช้ไฟล์ `sqls/200-meter-minimal.sql` เพื่อดึง 200 ราย (ต่อสาขา) จาก Oracle ด้วยพารามิเตอร์ `:ORG_OWNER_ID` (สาขา) และ `:DEBT_YM`
- เฟส Monthly ใช้ไฟล์ `sqls/200-meter-details.sql` เพื่อดึงรายละเอียดรายเดือนจาก Oracle โดยระบบจะ:
  - อ่านรายชื่อ `cust_code` ของสาขานั้นจาก `bm_custcode_init` ตามปีงบประมาณที่สัมพันธ์กับ `YM`
  - ตัดคำสั่ง `FETCH FIRST 200 ROWS ONLY` ออก (ถ้ามี) เพื่อให้ดึงได้ครบตามรายการที่เลือก
  - ไฟล์ SQL มี placeholder `/*__CUSTCODE_FILTER__*/` ซึ่งโปรแกรมจะแทนที่ด้วยเงื่อนไข `AND trn.CUST_CODE IN (:C0, :C1, ...)` ตามแบตช์ของ `cust_code`
  - สร้างเงื่อนไข IN ตามชุด `cust_code` เป็นแบตช์ (เช่น 100 รายต่อครั้ง) แล้วอัปเซิร์ตลง `bm_meter_details`
  - กรณี `cust_code` ใดไม่มีข้อมูลคืนมาจาก Oracle (เช่น ไม่ได้เป็นผู้ใช้น้ำแล้ว) ระบบจะอัปเซิร์ตแถวที่มีค่าตัวเลขเป็น 0 ทั้งหมด โดยช่องข้อความที่มีใน snapshot (จาก `bm_custcode_init`) จะถูกนำมาเติมให้ เช่น `use_type`, `meter_no`, `meter_state` ส่วนฟิลด์ข้อความอื่นที่ไม่มีใน snapshot จะเว้นว่างไว้ เพื่อระบุว่า “ไม่มีข้อมูลในงวดนั้น”

การตั้งค่า (.env)

- `TIMEZONE` โซนเวลา เช่น `Asia/Bangkok`
- `ORACLE_DSN` DSN สำหรับเชื่อมต่อ Oracle (เช่นของไลบรารี `godror`)
- `POSTGRES_DSN` DSN สำหรับเชื่อมต่อ PostgreSQL (pgx pool)
- `BRANCHES` รายชื่อสาขาคั่นด้วยจุลภาค เช่น `BA01,BA02,...` หากไม่กำหนด ระบบจะอ่านจากไฟล์ `docs/r6_branches.csv` (คอลัมน์แรก `ba_code`)
- `CRON_YEARLY` สเปก cron สำหรับงานประจำปี (ดีฟอลต์ 22:00 15 ต.ค.)
- `CRON_MONTHLY` สเปก cron สำหรับงานประจำเดือน (ดีฟอลต์ 08:00 ทุกวันที่ 16)

ไฟล์/โมดูลสำคัญ

- `cmd/sync/main.go`: จุดเริ่มรันโปรแกรม โหลดคอนฟิก สร้างการเชื่อมต่อ DB และเรียกใช้งาน Service หรือ Scheduler ตามโหมด
- `internal/config/config.go`: โหลดตัวแปรแวดล้อม/ดีฟอลต์ และอ่านรายชื่อสาขาจากไฟล์ CSV เมื่อไม่ตั้งค่า `BRANCHES`
- `internal/database/oracle.go`, `internal/database/postgres.go`: การเชื่อมต่อ Oracle/PG
- `internal/sync/sync.go`: แกนหลักของงานซิงก์
  - `InitCustcodes(...)`: ดึงรายชื่อ 200 รายแรกต่อสาขา แล้ว upsert ลง `bm_custcode_init`
  - `SyncMonthlyDetails(...)`: ดึงรายละเอียดรายเดือน (ตาม `YM` และสาขา) สำหรับรายชื่อจาก `bm_custcode_init` แล้ว upsert ลง `bm_meter_details`
  - `Schedule(...)`: ตั้ง Cron Job ตามสเปกและโซนเวลา
  - ฟังก์ชัน `FiscalYear`, `FiscalYearFromYM` สำหรับคำนวณปีงบประมาณ (ต.ค.–ก.ย.)
- `migrations/0001_init.sql` และ `internal/models/tables.sql`: สคีมาของตารางใน PostgreSQL
- `configs/.env.example`: ตัวอย่างตัวแปรแวดล้อม (อย่าใส่ความลับจริงลงใน repo)
- `docs/*.sql`: ไฟล์ SQL ฝั่ง Oracle สำหรับใช้งานในแต่ละเฟส

ลำดับการทำงานโดยย่อ

1) เริ่มต้นโปรแกรมและโหลดคอนฟิก (.env) กำหนดโซนเวลา/สาขา/DSN เชื่อมต่อ Oracle และ PostgreSQL
2) หากทำงานแบบ Scheduler: ตั้งเวลาตาม `CRON_YEARLY` และ `CRON_MONTHLY` และรอทริกเกอร์
3) เมื่อถึงรอบ Yearly Init: รัน `sqls/200-meter-minimal.sql` ต่อสาขา → upsert ลง `bm_custcode_init`
4) เมื่อถึงรอบ Monthly: โหลดรายการ `cust_code` ของสาขา → รัน `sqls/200-meter-details.sql` แบบแบตช์ → upsert ลง `bm_meter_details`

ตัวอย่างการรัน (โหมดหลัก)

- รัน Scheduler: `go run cmd/sync/main.go`
- รัน Yearly Init ครั้งเดียว: `MODE=init-once go run cmd/sync/main.go`
- รัน Monthly ครั้งเดียว (กำหนดเดือน): `MODE=month-once YM=202410 go run cmd/sync/main.go`

หมายเหตุ

- โค้ดตัวอย่างนี้ออกแบบให้เข้าใจ flow และปรับแต่งได้ง่าย ในงานจริงอาจต้องเพิ่มการทำ batching/binding ของ IN-list ให้มีประสิทธิภาพยิ่งขึ้น และเพิ่มการจัดการ error/retry ตามเหมาะสม
- การตั้งค่า `BRANCHES`, DSN และสเปก cron ควรตรวจสอบให้ตรงสภาพแวดล้อมจริงก่อนใช้งาน
