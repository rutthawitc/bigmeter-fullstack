# BigMeter API Spec (Frontend)

Purpose: Practical reference for building UI against the BigMeter API.

- Base URL: `/api/v1`
- Content-Type: `application/json; charset=utf-8`
- CORS: `*` (no credentials)
- Time: Timestamps are ISO 8601 (RFC 3339). Treat as UTC unless stated.
- Pagination: `limit` default 50 (max 500), `offset` default 0 (where supported)
- Search: `q` is case-insensitive substring across documented fields
- Sorting: `order_by` allowlist per endpoint; `sort=ASC|DESC` (default ASC)

## Endpoints

### Health
- GET `/healthz`
- 200 OK
- Response:
  {
    "status": "ok",
    "time": "2025-10-01T08:00:00+07:00"
  }

### Version
- GET `/version`
- 200 OK
- Response:
  {
    "service": "bigmeter-sync-api",
    "version": "0.1.0",
    "commit": "dev"
  }

### Branches
- GET `/branches`
- Query: `q` (optional substring match on code/name)
- Notes: Returns full list (no pagination); `name` may be omitted when not available.
- 200 OK
- Response:
  {
    "items": [ {"code": "BA01", "name": "..."}, {"code": "BA02"} ],
    "total": 2,
    "limit": 0,
    "offset": 0
  }

### Yearly Snapshot (Top-200 Custcodes)
- GET `/custcodes`
- Required: `branch=BAxx` and either `fiscal_year=YYYY` or `ym=YYYYMM`
  - Fiscal year: Oct–Dec → year+1; Jan–Sep → year. If `ym` provided, server derives `fiscal_year`.
- Optional:
  - `q`: searches across `cust_code, meter_no, use_type, org_name, use_name, cust_name, address, route_code, meter_size, meter_brand, meter_state, debt_ym`
  - `limit` (1..500; default 50), `offset` (>=0)
  - `order_by` allowlist: `cust_code, meter_no, use_type, created_at, org_name, use_name, cust_name, address, route_code, meter_size, meter_brand, meter_state, debt_ym`
  - `sort`: `ASC|DESC` (default ASC)
- 200 OK (example item; nullable fields omitted when null):
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

### Monthly Details
- GET `/details`
- Required: `ym=YYYYMM`, `branch=BAxx`
- Optional:
  - `cust_code`: filter to one or more custcodes. Accepts repeated query keys and/or comma-separated values (e.g., `cust_code=C1&cust_code=C2` or `cust_code=C1,C2`).
  - `q`: searches across `cust_code, meter_no, cust_name, address, route_code, org_name, use_type, use_name`
  - `limit` (1..500; default 50), `offset` (>=0)
  - `order_by` allowlist: `cust_code, present_water_usg, present_meter_count, average, created_at, org_name, use_type, use_name, cust_name, address, route_code, meter_no, meter_size, meter_brand, meter_state, debt_ym`
  - `sort`: `ASC|DESC`
- 200 OK (example; nullable fields omitted):
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
        "cust_code": "C99999",
        "meter_no": "M-0999",
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
- Notes:
  - "Zeroed" rows indicate a cohort cust_code had no Oracle data for the month; numeric fields are 0 and many text fields are null/omitted. The boolean `is_zeroed` is computed by the API.

### Monthly Details Summary
- GET `/details/summary`
- Required: `ym=YYYYMM`, `branch=BAxx`
- 200 OK:
  {
    "ym": "202410",
    "branch": "BA01",
    "total": 200,
    "zeroed": 15,
    "active": 185,
    "sum_present_water_usg": 12345.67
  }

### Series by Custcode
- GET `/custcodes/{cust_code}/details`
- Required (query): `branch=BAxx`, `from=YYYYMM`, `to=YYYYMM`
- 200 OK:
  {
    "cust_code": "C12345",
    "branch_code": "BA01",
    "from": "202410",
    "to": "202503",
    "series": [
      {"ym": "202410", "present_water_usg": 15.0, "present_meter_count": 300, "is_zeroed": false},
      {"ym": "202411", "present_water_usg": 0.0,  "present_meter_count": 0,   "is_zeroed": true}
    ]
  }

## Errors
- Format: `{ "error": "message" }`
- Examples:
  - 400 Bad Request: `{ "error": "ym and branch are required" }`
  - 404 Not Found: `{ "error": "not found" }` (not currently used by these endpoints)
  - 500 Internal Server Error: `{ "error": "internal error" }`

## Usage Notes
- Branch list: If not configured via env, the server loads branch codes from `docs/r6_branches.csv`.
- YM and Fiscal year: You can pass `ym=YYYYMM` and the API will derive `fiscal_year` where needed.
- Nullable fields: Many descriptive fields are nullable and will be omitted in JSON. Frontend should handle missing keys.
- Performance: Prefer server-side pagination and filtering for large lists.

## Examples (curl)

- Custcodes (derived fiscal year from ym):
  curl -s "http://localhost:8089/api/v1/custcodes?branch=BA01&ym=202410&limit=50&order_by=created_at&sort=DESC"

- Details (filter + search):
  curl -s "http://localhost:8089/api/v1/details?branch=BA01&ym=202410&cust_code=C1,C2&order_by=present_water_usg&sort=DESC&q=john"

- Summary:
  curl -s "http://localhost:8089/api/v1/details/summary?branch=BA01&ym=202410"

- Series by custcode:
  curl -s "http://localhost:8089/api/v1/custcodes/C12345/details?branch=BA01&from=202410&to=202503"

## Admin (Stub)

Status: Admin only; stubbed for frontend integration. No authentication is enforced yet and these endpoints do not execute real jobs.

- POST `/sync/init`
  - Body (JSON):
    { "branches": ["BA01", "BA02"], "debt_ym": "202410" }
  - 200 OK (stub response):
    {
      "fiscal_year": 2025,
      "branches": ["BA01"],
      "debt_ym": "202410",
      "stats": {"upserted": 0},
      "started_at": "2024-10-15T22:00:01Z",
      "finished_at": "2024-10-15T22:00:01Z",
      "note": "stub only; no jobs executed"
    }
  - Curl:
    curl -X POST -H "Content-Type: application/json" \
      -d '{"branches":["BA01"],"debt_ym":"202410"}' \
      http://localhost:8089/api/v1/sync/init

- POST `/sync/monthly`
  - Body (JSON):
    { "branches": ["BA01", "BA02"], "ym": "202410" }
  - 200 OK (stub response):
    {
      "ym": "202410",
      "branches": ["BA01"],
      "stats": {"upserted": 0, "zeroed": 0},
      "started_at": "2024-10-16T08:00:01Z",
      "finished_at": "2024-10-16T08:00:01Z",
      "note": "stub only; no jobs executed"
    }
  - Curl:
    curl -X POST -H "Content-Type: application/json" \
      -d '{"branches":["BA01"],"ym":"202410"}' \
      http://localhost:8089/api/v1/sync/monthly

- GET `/config`
  - 200 OK:
    { "timezone": "Asia/Bangkok", "cron_yearly": "0 0 22 15 10 *", "cron_monthly": "0 0 8 16 * *", "branches_count": 34 }
  - Curl:
    curl -s http://localhost:8089/api/v1/config
