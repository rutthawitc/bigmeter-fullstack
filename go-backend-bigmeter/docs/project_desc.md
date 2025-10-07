BigMeter Sync & API — Project Description

Purpose

- Capture and track a yearly cohort of “top 200” water‑usage debt customers per branch (BA code) from an Oracle system, and persist the cohort and their subsequent monthly details in PostgreSQL for reporting and APIs.
- Align scheduling with the Thai fiscal year: October → September.

High‑Level Flow

- Yearly init (Oct 16, 01:30 Asia/Bangkok)
  - For each branch (`ORG_OWNER_ID = ba_code` from docs/r6_branches.csv), run `sqls/200-meter-minimal.sql` against Oracle with `DEBT_YM=YYYY10` (October) and take the first 200 rows by current usage.
  - Upsert results into Postgres table `bm_custcode_init` with a fiscal year label (Oct–Dec → next label year).
  - This defines the cohort of customers we will follow for the rest of the fiscal year.

- Monthly sync (16th, 08:00 Asia/Bangkok)
  - Determine the previous month `YM=YYYYMM`.
  - For each branch, load the cohort custcodes from `bm_custcode_init` for the current fiscal year.
  - Query Oracle with `sqls/200-meter-details.sql` filtered to only those custcodes; run in batches using a parameterized `IN (:C0, :C1, …)` clause (batch size configurable; default 100).
  - Upsert details into Postgres table `bm_meter_details` keyed by `(year_month, branch_code, cust_code)`.
  - Any `FETCH FIRST 200 ROWS ONLY` in the monthly query is removed automatically so all matching cohort rows are captured.

Data Model (PostgreSQL)

- `bm_custcode_init`
  - Columns: `fiscal_year, branch_code, cust_code, use_type, meter_no, meter_state, debt_ym, created_at`.
  - Uniqueness: `(fiscal_year, branch_code, cust_code)`.
  - Population source: Oracle minimal query (top 200 per branch each year in October).

- `bm_meter_details`
  - Columns: `year_month, branch_code, org_name, cust_code, use_type, use_name, cust_name, address, route_code, meter_no, meter_size, meter_brand, meter_state, average, present_meter_count, present_water_usg, debt_ym, created_at`.
  - Uniqueness: `(year_month, branch_code, cust_code)`.
  - Population source: Oracle details query filtered by the cohort custcodes.

Scheduling & Timezone

- Timezone: `Asia/Bangkok`.
- Yearly init: cron `0 30 1 16 10 *` (16 Oct 01:30).
- Monthly sync: cron `0 0 8 16 * *` (16th 08:00 for previous month).
- Fiscal year label: if month ≥ 10 (Oct–Dec), label = `year + 1`; else label = `year`.

Configuration

- Env vars (see `configs/.env.example`):
  - `ORACLE_DSN`, `POSTGRES_DSN`, `TIMEZONE`, `CRON_YEARLY`, `CRON_MONTHLY`.
  - `BRANCHES` (comma‑separated BA codes). If unset, the app auto‑loads BA codes from `docs/r6_branches.csv` (first column `ba_code`).
  - `MODE`:
    - empty → run scheduler
    - `init-once` with optional `DEBT_YM=YYYYMM`
    - `month-once` with optional `YM=YYYYMM`

Components

- `cmd/sync`: executable that runs either the scheduler or one‑shot jobs.
- `internal/sync`: service implementing yearly init and monthly sync logic.
- `internal/database`: Oracle and Postgres connectors.
- `migrations/0001_init.sql`: creates Postgres tables.
- `docs/*.sql`: the Oracle SQL templates used by the jobs.

Key Behaviors & Constraints

- Cohort tracking: monthly data is captured only for custcodes that were recorded during yearly init for the active fiscal year. This keeps longitudinal tracking consistent throughout the year.
- Upsert semantics: reruns are idempotent — records are updated on conflict using the unique keys.
- Batch querying: reduces round‑trips to Oracle by grouping custcodes into `IN (...)` lists; default batch size is 100 and can be tuned.
- Error handling: each job uses transactions on Postgres; errors abort the transaction for that branch/month to avoid partial writes.

Assumptions

- `ORG_OWNER_ID` in Oracle equals `ba_code` (first column of r6_branches.csv).
- Yearly init uses October (`DEBT_YM=YYYY10`) as the reference month.
- The provided SQL templates are valid for the Oracle schema in use; the monthly template supports appending an extra `AND trn.CUST_CODE IN (...)` before its `ORDER BY` clause.

Future Work

- API binary under `cmd/api/` to expose reporting endpoints from Postgres.
- Structured logging and Prometheus metrics for job timings, row counts, and errors.
- Dockerfiles and compose for `api`, `sync`, and dependencies.
- Tests (unit with sqlmock, table‑driven) and integration tests against containers.
