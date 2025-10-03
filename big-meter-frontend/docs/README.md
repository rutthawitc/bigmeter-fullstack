BigMeter Sync (Oracle → PostgreSQL)

Overview

- Yearly init on Oct 15th 22:00: query Oracle with `sqls/200-meter-minimal.sql` per branch and store custcodes to `bm_custcode_init`.
- Monthly on the 16th at 08:00: query Oracle with `sqls/200-meter-details.sql` and upsert into `bm_meter_details`. The query is filtered by the custcodes captured at init and runs in batches; any `FETCH FIRST 200 ROWS ONLY` is removed automatically.
- Schedules run in `Asia/Bangkok` by default.

Quick Start

1. Copy `.env.example` to `.env` (root) and fill DSNs + branch list. For Docker Compose, you can also copy `configs/.env.example` to `configs/.env`.
2. Apply migrations in Postgres using your preferred tool with `migrations/0001_init.sql`.
3. Run sync scheduler:
   - `go run cmd/sync/main.go` (scheduler)
   - `MODE=init-once YM=202410 go run cmd/sync/main.go` (one-time yearly init; YM is Gregorian)
   - `MODE=month-once YM=202410 go run cmd/sync/main.go` (one-time monthly)
   - CLI reference: see `docs/cli_cheat_sheet.md` for more examples

API Server (Gin)

- Start API (Gin): `go run cmd/api/main.go` (uses `POSTGRES_DSN` from `.env`)
- Base URL: `/api/v1`
- Full API spec: `docs/API-Spec.md`
- Useful endpoints:
  - `GET /api/v1/healthz`
  - `GET /api/v1/branches`
  - `GET /api/v1/custcodes?branch=BA01&ym=202410`
  - `GET /api/v1/details?branch=BA01&ym=202410`

Search/Sort

- Custcodes: query `q` searches across `cust_code, meter_no, use_type, org_name, use_name, cust_name, address, route_code, meter_size, meter_brand, meter_state, debt_ym`.
- Details: query `q` searches across `cust_code, meter_no, cust_name, address, route_code, org_name, use_type, use_name`.
- `order_by` allowlist (custcodes): `cust_code, meter_no, use_type, created_at, org_name, use_name, cust_name, address, route_code, meter_size, meter_brand, meter_state, debt_ym`.
- `order_by` allowlist (details): `cust_code, present_water_usg, present_meter_count, average, created_at, org_name, use_type, use_name, cust_name, address, route_code, meter_no, meter_size, meter_brand, meter_state, debt_ym`.
- `sort`: `ASC|DESC`.

Notes

- Adjust SQL filters to limit details by custcodes per branch if you want to restrict to the 200-list captured on init. The skeleton executes as provided and can be extended to add `IN (:cust1, :cust2, ...)` binding or temporary tables.
- If `BRANCHES` is unset, the app auto-loads branch list from `docs/r6_branches.csv` (first column `ba_code`).
- All times and cron schedules honor `TIMEZONE`.
- Do not commit secrets or real DSNs. Use `.env.example` (root) and `configs/.env.example` as templates; `.env` and `configs/.env` are git-ignored.

Schema changes (2025-09)

- Migration `0004_extend_custcode_init.sql` adds richer snapshot fields to `bm_custcode_init`.
- `/custcodes` now returns these fields (nullable and omitted when null): `org_name, use_name, cust_name, address, route_code, meter_size, meter_brand`.
- `/details` omits nullable fields when not present.

Docker (Postgres + API + Sync)

- Requirements: Docker 24+, Docker Compose V2
- Prepare env for Compose (no secrets; examples below):
  1. Copy `configs/.env.example` to `configs/.env` and adjust values as needed.
  2. Optionally create a root `.env` file from `.env.example` (or export envs) for Compose variables:
     - `POSTGRES_USER=postgres`
     - `POSTGRES_PASSWORD=postgres`
     - `POSTGRES_DB=bigmeter`
     - `TIMEZONE=Asia/Bangkok`
- Start services (Postgres 17.6 + API; Sync runs automatically if `ORACLE_DSN` is set). Optional profile: `setup` (migrations + branch seeding):
  - Base stack: `docker compose up -d --build`
  - If you set `ORACLE_DSN` in `.env`, the Sync service will start and run the scheduler by default.
  - Run setup (first time only): `docker compose --profile setup up -d migrate seed_branches`
  - API available at `http://localhost:8089/api/v1/healthz`
- What runs:
  - `postgres` (17.6) with persistent volume `pgdata`
  - `migrate` (profile `setup`) applies all SQL in `migrations/`
  - `seed_branches` (profile `setup`) imports branch list from `docs/r6_branches.csv` into `bm_branches`
  - `api` built from `docker/Dockerfile.api`, uses `POSTGRES_DSN` pointing at the `postgres` service
  - `sync` built from `docker/Dockerfile.sync-thick` using godror (Oracle Instant Client, thick). Runs one-shots via `MODE`.

Notes:

- The sync container uses Oracle Instant Client (thick) — place an Instant Client ZIP in `orc_client/` (see `docker/Dockerfile.sync-thick`).
- Provide a Service Name/SID DSN as shown in `docs/dev.md` (EZCONNECT recommended).
- The sync service is part of the base stack; leave `ORACLE_DSN` empty to skip it, or provide the DSN to enable it.
- You can still run sync locally with `make sync-init`.

Troubleshooting (Oracle)

- If SCAN endpoints close connections (ORA-12537) or heavy queries time out, use a node VIP or a DESCRIPTION DSN with `SERVER=DEDICATED`. See `docs/issues-2025-09-19.md` for findings and mitigations.

Branch list in DB

- Table: `bm_branches(code TEXT PRIMARY KEY, name TEXT)` created by `migrations/0002_branches.sql`.
- Seed: `seed_branches` service loads `docs/r6_branches.csv`. To re-seed manually:
  - `docker compose run --rm seed_branches`
  - Or from psql: `\copy bm_branches(code,name) FROM 'docs/r6_branches.csv' CSV HEADER` after creating a staging table/select as in compose.
