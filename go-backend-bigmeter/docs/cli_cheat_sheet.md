BigMeter CLI Cheat Sheet

Summary

- YM is Gregorian (YYYYMM) across all modes. The sync service converts to Thai DEBT_YM internally when calling Oracle.
- Leave `BRANCHES` empty to use all branches from `docs/r6_branches.csv`. Set `BRANCHES=1063` to limit.
- Never commit real credentials. Use `.env.example` as a template only.

Docker Compose

- Start base stack (Postgres + API):
  - `docker compose up -d --build`
- Migrations (one‑off):
  - `docker compose --profile setup run --rm migrate`
- Seed branches (one‑off):
  - `docker compose --profile setup run --rm seed_branches`
- Logs:
  - API: `docker compose logs -f api`
  - Sync: `docker compose logs -f sync`

One‑Shot Sync Jobs

- Yearly init (all branches):
  - `docker compose run --rm -e BRANCHES= -e MODE=init-once -e YM=202410 sync`
- Yearly init (single branch):
  - `docker compose run --rm -e BRANCHES=1063 -e MODE=init-once -e YM=202410 sync`
- Monthly details (all branches):
  - `docker compose run --rm -e BRANCHES= -e MODE=month-once -e YM=202410 sync`
- Monthly details (single branch):
  - `docker compose run --rm -e BRANCHES=1063 -e MODE=month-once -e YM=202410 sync`
- Oracle connectivity test:
  - `docker compose run --rm -e BRANCHES=1063 -e MODE=ora-test -e YM=202410 sync`

Tuning (optional)

- Batch size: `-e BATCH_SIZE=200`
- Concurrency: `-e SYNC_CONCURRENCY=4`
- Retries: `-e SYNC_RETRIES=2 -e SYNC_RETRY_DELAY=10s`

Scheduler Mode

- Starts automatically when `ORACLE_DSN` is set and `MODE` is empty:
  - `docker compose up -d sync`
- Cron specs (with seconds field):
  - `CRON_YEARLY="0 0 22 15 10 *"` (Oct 15, 22:00)
  - `CRON_MONTHLY="0 0 8 16 * *"` (16th, 08:00)
- Time zone: `TIMEZONE=Asia/Bangkok`

API (local without Docker)

- `POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/bigmeter?sslmode=disable TIMEZONE=Asia/Bangkok go run cmd/api/main.go`

Ad‑hoc Metrics (optional)

- Expose Prometheus metrics during a one‑off run:
  - `docker compose run --rm -p 9090:9090 -e METRICS_ADDR=':9090' -e MODE=month-once -e YM=202410 sync`
  - Scrape: `http://localhost:9090/metrics`

API (curl examples)

- Set base URL:
  - `BASE=http://localhost:8089/api/v1`
- Health:
  - `curl -s "$BASE/healthz" | jq`
- Branches (search):
  - `curl -s "$BASE/branches?q=BA" | jq`
- Custcodes (list top-200 cohort for a fiscal derived from YM):
  - `curl -s "$BASE/custcodes?branch=1063&ym=202410&limit=50" | jq`
  - With sorting: `curl -s "$BASE/custcodes?branch=1063&ym=202410&order_by=org_name&sort=ASC" | jq`
  - Search across many fields: `curl -s "$BASE/custcodes?branch=1063&ym=202410&q=RT01" | jq`
- Monthly details (list):
  - `curl -s "$BASE/details?branch=1063&ym=202410&limit=100" | jq`
  - Sort by usage: `curl -s "$BASE/details?branch=1063&ym=202410&order_by=present_water_usg&sort=DESC" | jq`
  - Search by route or name: `curl -s "$BASE/details?branch=1063&ym=202410&q=RT01" | jq`
  - Filter by custcodes (repeatable): `curl -s "$BASE/details?branch=1063&ym=202410&cust_code=C12345&cust_code=C67890" | jq`
  - Or comma-separated: `curl -s "$BASE/details?branch=1063&ym=202410&cust_code=C12345,C67890" | jq`
- Monthly details summary:
  - `curl -s "$BASE/details/summary?branch=1063&ym=202410" | jq`
- Customer series (range):
  - `curl -s "$BASE/custcodes/C12345/details?branch=1063&from=202410&to=202508" | jq`

Notes

- Pagination: `limit` (default 50, max 500), `offset` (default 0).
- Sorting (custcodes): `order_by` in `cust_code|meter_no|use_type|created_at`, `sort` in `ASC|DESC`.
- Sorting (details): `order_by` in `cust_code|present_water_usg|created_at`, `sort` in `ASC|DESC`.
