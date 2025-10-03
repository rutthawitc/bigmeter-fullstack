BigMeter Sync & API — Manual

This document explains how to run, test, and operate the BigMeter API and Oracle→PostgreSQL sync service. It’s written for clarity and quick onboarding.

1) What You Get

- API service (Gin) exposing read‑only endpoints over Postgres.
- Sync service that loads a yearly cohort (top‑200 per branch) and monthly details from Oracle into Postgres.
- Docker Compose for Postgres, migrations, branch seeding, API, and an optional sync container.

2) Requirements

- Docker 24+ and Docker Compose V2
- Go 1.22+ (only if you want to run binaries locally)
- Oracle access (only if you want to run the sync job)

3) Repo Layout (short)

- `cmd/api`   — API server main
- `cmd/sync`  — Sync/scheduler main
- `internal/*` — app packages (config, database, api, sync)
- `migrations/` — Postgres schema SQL
- `sqls/` — Oracle SQL templates
- `docs/` — branch CSV and docs
- `docker/` — Dockerfiles for api/sync

4) Env Files

- Do not commit secrets. Examples are provided only.
- Root `.env` is used by Docker Compose (see `.env.example`).
- App settings for local runs: copy `configs/.env.example` to `configs/.env` if needed.

Common variables (Compose reads these from `.env`):

- `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `TIMEZONE` (default `Asia/Bangkok`)
- `PORT` (API port, default `8089`)
- Optional for sync:
  - `ORACLE_DSN=oracle://USER:PASS@host:1521/SERVICE?timezone=Asia%2FBangkok`
  - `MODE` (empty=scheduler, or `init-once` / `month-once`)
  - `YM` (for `month-once`), `DEBT_YM` (for `init-once`)

See `docs/dev.md` for DSN examples and notes.

5) Quick Start with Docker

The base stack runs Postgres + API. The Sync service is included and starts automatically when `ORACLE_DSN` is set. The `setup` profile is used for one‑time schema + branch seeding.

- Start base stack (Postgres + API; Sync auto‑starts if `ORACLE_DSN` is set):
  - `docker compose up -d --build`
  - Health: `curl http://localhost:8089/api/v1/healthz`
  - Branches: `curl http://localhost:8089/api/v1/branches`

- One‑time setup (schema + branch seeding):
  - `docker compose --profile setup up -d migrate seed_branches`
  - Verify: `docker compose exec -T postgres psql -U $POSTGRES_USER -d $POSTGRES_DB -c "\\dt"`

- Sync service:
  - Provide `ORACLE_DSN` in `.env` and restart compose.
  - Logs: `docker logs -f bigmeter_sync`

- One‑shot sync via compose (Gregorian YM for all modes):
  - Yearly init: `docker compose run --rm -e MODE=init-once -e YM=202410 sync`
  - Monthly: `docker compose run --rm -e MODE=month-once -e YM=202410 sync`

6) Makefile Shortcuts

- `make docker-up` — build and start base stack (Postgres + API)
- `make docker-down` — stop containers
- `make docker-logs` — tail API logs
- `make migrate` — run migrations (`--profile setup`)
- `make seed` — seed branches (`--profile setup`)
- `make seed-sample` — insert demo rows for quick API testing
- `make api-local` — run API locally (needs `POSTGRES_DSN`)
- `make sync-init` — run yearly init once locally (needs Oracle)
- `make sync-month` — run monthly once locally (needs Oracle)
- `make sync-scheduler` — run scheduler locally (needs Oracle)

7) Running Locally (without Docker)

- API only:
  - `POSTGRES_DSN='postgres://postgres:postgres@localhost:5432/bigmeter?sslmode=disable' TIMEZONE=Asia/Bangkok go run cmd/api/main.go`
- Sync only (requires Oracle), all use Gregorian `YM`:
  - Yearly: `MODE=init-once YM=202410 POSTGRES_DSN=... ORACLE_DSN=... go run -tags oracle cmd/sync/main.go`
  - Monthly: `MODE=month-once YM=202410 POSTGRES_DSN=... ORACLE_DSN=... go run -tags oracle cmd/sync/main.go`
  - Scheduler: `TIMEZONE=Asia/Bangkok POSTGRES_DSN=... ORACLE_DSN=... go run -tags oracle cmd/sync/main.go`

8) Useful API Calls

- `GET /api/v1/healthz`
- `GET /api/v1/branches`
- `GET /api/v1/custcodes?branch=BA01&fiscal_year=2025`
- `GET /api/v1/details?branch=BA01&ym=202410`
- `GET /api/v1/details/summary?branch=BA01&ym=202410`
- `GET /api/v1/custcodes/C12345/details?branch=BA01&from=202410&to=202410`

Tip: Use `limit`, `offset`, `order_by`, and `sort` on list endpoints.

9) Common Tasks

- Change branches: set `BRANCHES` in env or update `docs/r6_branches.csv` and run `make seed`.
- Recreate DB: `docker compose down -v` then run setup profile again.
- Seed demo data: `make seed-sample` then query the API endpoints above.

10) Troubleshooting

- API up but empty results: run setup and seed or `make seed-sample`.
- Sync container Oracle errors: verify `ORACLE_DSN`, network, and include `?timezone=Asia%2FBangkok`.
- Port conflicts: change published ports in `docker-compose.yml`.

11) Security Notes

- Never commit real credentials. Use `.env.example` as a template only.
- `.env` and `configs/.env` are ignored by git.

12) Clean Up

- Stop: `docker compose down`
- Reset data: `docker compose down -v`
