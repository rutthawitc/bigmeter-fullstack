Dev Notes — BigMeter Sync & API

What this repo contains

- `cmd/sync`: scheduler/CLI for Oracle→Postgres sync (yearly init + monthly).
- `internal/sync`: job logic (init cohort, monthly details, cron scheduler).
- `internal/config`: env/config loader (timezone, DSNs, cron specs, branches).
- `internal/database`: Oracle (godror thick) and Postgres connectors.
- `migrations/`: Postgres schema.
- `docs/`: Oracle SQL templates, branch CSV, and project docs.

Quick setup

- Go 1.22+
- Copy `configs/.env.example` → `.env` and fill:
  - `ORACLE_DSN` for godror (EZCONNECT):
    - Service Name: `USER/PASS@host:1521/SERVICE`
    - SID: `USER/PASS@host:1521/ORCL`
    - Username/Password → credentials in DSN (do not commit)
    - Note: Oracle stores `DEBT_YM` in Buddhist year (e.g., 256710). Pass Thai YYYYMM when binding.
  - `POSTGRES_DSN` (e.g., `postgres://postgres:postgres@localhost:5432/bigmeter?sslmode=disable`)
  - Optional: `BRANCHES` (comma‑separated BA codes). If empty, uses `docs/r6_branches.csv` first column (`ba_code`).
  - Optional: `TIMEZONE`, `CRON_YEARLY`, `CRON_MONTHLY` (defaults match requirements).

Apply migrations

- With psql: `psql "$POSTGRES_DSN" -f migrations/0001_init.sql`
- Then apply subsequent files in order (e.g., `0002_*.sql`, `0003_*.sql`, `0004_extend_custcode_init.sql`).
- Or use Docker compose profiles: `docker compose --profile setup run --rm migrate`.

Run modes (Gregorian YM)

- Diagnostic (Oracle): `MODE=ora-test BRANCHES=1063 YM=202410 go run cmd/sync/main.go`
- Yearly init one‑shot: `MODE=init-once YM=202410 go run cmd/sync/main.go`

Behavior recap

- Yearly init (Oct 15 22:00, Asia/Bangkok): runs `sqls/200-meter-minimal.sql` per branch, keeps top‑200 and upserts into `bm_custcode_init` with fiscal year label. As of 2025‑09, the query limits/deduplicates first, then joins heavy dimensions for better performance, and includes richer fields persisted by migration `0004`.
- Monthly (16th 08:00): loads cohort custcodes from `bm_custcode_init`, runs `sqls/200-meter-details.sql` filtered to those codes in batches, and upserts into `bm_meter_details`. Any `FETCH FIRST 200 ROWS ONLY` is removed automatically in monthly. The details SQL is trimmed to core numeric/identity fields; descriptive fields not present will be stored as NULL and omitted from API JSON.
- Details SQL contains a placeholder `/*__CUSTCODE_FILTER__*/` which the service replaces at runtime with an `AND trn.CUST_CODE IN (:C0, :C1, ...)` clause for the current batch.
- No‑rows case (monthly): if a cust_code in the cohort returns no rows from Oracle for the given YM, the service upserts a "zeroed" row into `bm_meter_details` with numeric fields set to 0 and selected text fields filled from the snapshot (`bm_custcode_init`): `use_type`, `meter_no`, `meter_state`. Other text fields remain empty.
- ORG_OWNER_ID mapping = `ba_code` (first column in `docs/r6_branches.csv`).
- Fiscal year: Oct–Dec → year+1; Jan–Sep → year.

Logging (monthly)

- Per batch: logs `ym`, `branch`, zeroed count, batch index and row range processed.
- After commit: logs total zeroed rows for the run.

Useful envs (from `.env`)

- `TIMEZONE` (default `Asia/Bangkok`)
- `CRON_YEARLY` (default `0 30 1 16 10 *`)
- `CRON_MONTHLY` (default `0 0 8 16 * *`)
- `MODE` (`init-once`, `month-once`, or empty for scheduler)
- `YM` (Gregorian YYYYMM) for both init‑once and month‑once

Oracle DSN notes (thick)

- Prefer EZCONNECT: `USER/PASS@host:1521/SERVICE`.
- For SCAN issues, prefer node VIP or DESCRIPTION DSN with `SERVER=DEDICATED`.
- Keep credentials local only; never commit real passwords. `.env.example` provides placeholders.

Common tasks

- Change branch list: set `BRANCHES` or edit `docs/r6_branches.csv` (first col) then run.
- Change cron timings: set `CRON_YEARLY`/`CRON_MONTHLY` in `.env`.
- Run for different month: `MODE=month-once YM=YYYYMM ...`.
- Re‑run yearly init for debugging: `MODE=init-once DEBT_YM=YYYY10 ...` (idempotent upsert).

Coding conventions

- Go formatting: `gofmt -s -w .`
- Vet: `go vet ./...`
- File naming: snake_case for files; packages short/lowercase.
- Errors: wrap with context using `fmt.Errorf("context: %w", err)`.

Testing (next)

- Use `testify` + `sqlmock` for unit tests; table‑driven style.
- Add integration tests under `tests/` (can guard with `//go:build integration`).

Planned next session / TODOs

- Metrics: add structured metrics (Prometheus), job duration and row‑count.
- Dockerfiles + docker‑compose for `api`, `sync`, Postgres, Oracle (or instant client), Redis if needed.
- Implement `cmd/api` with read‑only endpoints over `bm_custcode_init` and `bm_meter_details` per `docs/api_spec.md` (branch scope required on `/details`).
- Configurable Oracle batch size for monthly IN‑clause (default 100) and proper array binds if needed.
- Add retry/backoff and per‑branch isolation for monthly runs.

API & Diagrams

- API spec (draft): see `docs/api_spec.md`. Note: `GET /api/v1/details` requires `branch`; does not return all branches by default.
- Mermaid diagrams: `docs/system_overview.mmd`, `docs/data_flow.mmd`, `docs/monthly_sequence.mmd`, `docs/db_er.mmd`.
