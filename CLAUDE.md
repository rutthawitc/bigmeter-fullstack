# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Big Meter** is a full-stack Dockerized water usage monitoring system for the Provincial Waterworks Authority (PWA) of Thailand. The system:
- Syncs top-200 high-usage customer data from Oracle to PostgreSQL on a scheduled basis
- Provides a React-based dashboard for analyzing water consumption patterns
- Identifies significant usage decreases and generates Excel reports
- Uses Thai Buddhist calendar and Thai language interface

## Repository Structure

```
big-meter/
├── go-backend-bigmeter/     # Go backend (API + sync service)
│   ├── cmd/api/             # REST API server entry point
│   ├── cmd/sync/            # Sync service entry point (cron scheduler)
│   ├── internal/            # Internal packages
│   │   ├── api/             # API handlers and routes
│   │   ├── config/          # Configuration loading
│   │   ├── database/        # Oracle & PostgreSQL clients
│   │   └── sync/            # Sync business logic
│   ├── migrations/          # PostgreSQL schema migrations
│   ├── sqls/                # Oracle SQL query templates
│   └── docker-compose.yml   # Docker orchestration (all services)
│
└── big-meter-frontend/      # React TypeScript frontend
    ├── src/api/             # API client layer
    ├── src/screens/         # Page components
    └── src/lib/             # Utilities and hooks
```

## Development Commands

### Working Directory
All Docker commands must be run from `go-backend-bigmeter/` directory.

### Database Setup
```bash
cd go-backend-bigmeter

# Run migrations and seed branches (required first-time setup)
docker-compose --profile setup up migrate seed_branches
```

### Running Services
```bash
cd go-backend-bigmeter

# Start all services
docker-compose up --build

# Start in background
docker-compose up -d --build

# View logs
docker-compose logs -f [service_name]  # api, sync, frontend, postgres

# Stop services
docker-compose down

# Stop and remove volumes (⚠️ deletes data)
docker-compose down -v
```

### Backend Development
```bash
cd go-backend-bigmeter

# Run API server locally (requires PostgreSQL)
go run cmd/api/main.go

# Run sync service locally (requires Oracle + PostgreSQL)
go run cmd/sync/main.go

# Run tests
go test ./...
```

### Frontend Development
```bash
cd big-meter-frontend

# Install dependencies
pnpm install

# Run dev server (port 5173)
pnpm dev

# Build production bundle
pnpm build

# Preview production build
pnpm preview

# Lint code
pnpm lint

# Format code
pnpm format
```

## Architecture

### Multi-Service Docker Stack

1. **Frontend (nginx + React)** - Port 3000
   - Serves static React bundle via nginx
   - Proxies `/api` requests to backend
   - Thai language UI with water usage analytics

2. **Backend API (Go + Gin)** - Port 8089
   - REST API for branches, customer codes, and details
   - Health check endpoints
   - Manual sync triggers

3. **Sync Service (Go + Cron)** - Internal
   - Yearly init: Oct 15, 22:00 Bangkok time → captures top-200 customers per branch
   - Monthly sync: 16th, 08:00 Bangkok time → updates usage details
   - Oracle → PostgreSQL data synchronization

4. **PostgreSQL Database** - Port 5432
   - `bm_branches` - Branch metadata
   - `bm_custcode_init` - Yearly customer cohort (top-200)
   - `bm_meter_details` - Monthly usage records
   - `bm_sync_logs` - Sync operation history and monitoring

### Data Flow

```
Oracle DB → [Sync Service] → PostgreSQL → [API Server] → Frontend
             (cron jobs)                   (REST API)    (React)
```

### Sync Logic

**Yearly Init (October)**:
- For each branch, query Oracle for top-200 customers by current usage debt
- Store cohort in `bm_custcode_init` with fiscal year label
- Fiscal year: Oct-Dec of year N → labeled as year N+1

**Monthly Sync (16th of each month)**:
- Load cohort custcodes from `bm_custcode_init` for current fiscal year
- Query Oracle for detailed usage data filtered to cohort (batched, default 100 per batch)
- Upsert into `bm_meter_details` keyed by `(year_month, branch_code, cust_code)`

## Key Implementation Details

### Environment Configuration

Edit `go-backend-bigmeter/.env`:
```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=bigmeter
TIMEZONE=Asia/Bangkok
PORT=8089
ORACLE_DSN=user/pass@host:1521/service_name
BRANCHES=BA01,BA02,BA03  # Comma-separated branch codes
```

### Sync Service Modes

The sync service (`cmd/sync/main.go`) supports multiple modes via `MODE` env var:

- **Empty (default)**: Run scheduler with cron jobs
- **`init-once`**: Run yearly init once (use `YM=YYYYMM` or defaults to current October)
- **`month-once`**: Run monthly sync once (requires `YM=YYYYMM`)
- **`ora-test`**: Test Oracle connectivity (requires `BRANCHES` and `YM=YYYYMM`)

### Date Format Conversions

- **Gregorian ↔ Thai Buddhist Calendar**: Thai year = Gregorian year + 543
- **Year-month format**: `YYYYMM` (e.g., `202412` for December 2024)
- Oracle queries use Thai Buddhist calendar; internal processing uses Gregorian
- `normalizeGregorianYM()` and `toThaiYM()` handle conversions

### Frontend Data Fetching

- Uses **TanStack Query (React Query)** for caching and parallel requests
- `useQueries` fetches multiple months of data simultaneously (up to 12 months)
- Lazy-loads Excel export library to reduce bundle size (~700KB savings)

### Sync Operation Logging

**Purpose**: Track all sync operations for monitoring, debugging, and audit trail

**Features**:
- Automatic logging of every sync operation (yearly init and monthly sync)
- Records: status (success/error/in_progress), duration, record counts, error messages
- Tracks trigger source: `api` (manual via admin UI), `scheduler` (cron job), `manual` (CLI)
- Admin UI displays real-time sync history with auto-refresh every 10 seconds
- Color-coded status badges and performance metrics

**Implementation**:
- Backend: `internal/sync/log.go` (LogRepository), `internal/sync/service.go` (logging integration)
- Frontend: `src/api/syncLogs.ts` (API client), `src/screens/AdminPage.tsx` (sync logs table)
- Database: `bm_sync_logs` table with indexes on branch_code, sync_type, status, created_at

**Admin Page Features**:
- View last 20 sync operations with full details
- Filter by branch, type, or status (via API query params)
- See actual upserted/zeroed record counts
- Monitor operation duration and success rate
- Automatically refreshes after triggering new sync operations

### API Endpoints

Base URL: `http://localhost:8089/api/v1`

- `GET /healthz` - Health check
- `GET /branches?q=&limit=&offset=` - List branches
- `GET /custcodes?branch=&ym=&fiscal_year=&q=&limit=&offset=` - Customer codes
- `GET /details?ym=&branch=&q=&limit=&offset=&order_by=&sort=` - Usage details
- `POST /sync/init` - Trigger yearly sync (body: `{"branches": ["BA01"], "debt_ym": "202410"}`)
  - Returns: `{"fiscal_year": 2025, "branches": [...], "stats": {"upserted": 199, "zeroed": 0}, "started_at": "...", "finished_at": "..."}`
- `POST /sync/monthly` - Trigger monthly sync (body: `{"branches": ["BA01"], "ym": "202501", "batch_size": 100}`)
  - Returns: `{"ym": "202501", "branches": [...], "stats": {"upserted": 199, "zeroed": 5}, "started_at": "...", "finished_at": "..."}`
- `GET /sync/logs?branch=&sync_type=&status=&limit=&offset=` - Retrieve sync operation logs
  - Returns: `{"items": [...], "total": 1, "limit": 50, "offset": 0}`

## Database Schema

### bm_branches
- `code` (PK) - Branch code (e.g., "BA01")
- `name` - Branch name
- `created_at` - Timestamp

### bm_custcode_init
- `fiscal_year` (PK) - Fiscal year label
- `branch_code` (PK) - Branch code
- `cust_code` (PK) - Customer code
- `org_name`, `use_type`, `meter_no`, `meter_state`, `debt_ym`, `created_at`

### bm_meter_details
- `year_month` (PK) - YYYYMM format
- `branch_code` (PK) - Branch code
- `cust_code` (PK) - Customer code
- `org_name`, `use_type`, `use_name`, `cust_name`, `address`, `meter_no`, `present_water_usg`, `average`, etc.

### bm_sync_logs
- `id` (PK) - Auto-incrementing log ID
- `sync_type` - 'yearly_init' or 'monthly_sync'
- `branch_code` - Branch code for this operation
- `year_month` - YYYYMM (for monthly_sync)
- `fiscal_year` - Fiscal year (for yearly_init)
- `debt_ym` - Thai Buddhist YYYYMM (for yearly_init)
- `status` - 'success', 'error', or 'in_progress'
- `started_at` - Operation start timestamp
- `finished_at` - Operation completion timestamp
- `duration_ms` - Duration in milliseconds
- `records_upserted` - Number of records inserted/updated
- `records_zeroed` - Number of zeroed records (monthly_sync)
- `error_message` - Error details if failed
- `triggered_by` - 'api', 'scheduler', or 'manual'
- `created_at` - Log creation timestamp

## Important Conventions

### Go Backend

- Uses **Go 1.25** with modules
- Driver: `github.com/sijms/go-ora/v2` for Oracle (thick client, requires Oracle Instant Client in Docker)
- Driver: `github.com/jackc/pgx/v5` for PostgreSQL
- Web framework: `github.com/gin-gonic/gin`
- Scheduler: `github.com/robfig/cron/v3` with 6-field cron (supports seconds)
- Config loading: `internal/config/config.go` reads from env vars with defaults
- Database connections: `internal/database/oracle.go` and `postgres.go`

### Frontend

- **React 19** with TypeScript, built via **Vite** and **SWC** compiler
- **Tailwind CSS v4** with Vite plugin (no PostCSS config needed)
- Authentication: PWA Intranet login via `VITE_LOGIN_API` env var
- State: React Query for server state, Context for auth, local state for UI
- Sparkline charts: `@visx/visx` for historical usage visualization
- Excel export: `xlsx` library (lazy-loaded to optimize bundle)

### Docker

- Multi-stage builds for frontend (builder + nginx)
- Backend uses thick Oracle client (`Dockerfile.sync-thick`)
- All services orchestrated via `docker-compose.yml` in `go-backend-bigmeter/`
- Profiles: `--profile setup` for one-time migration/seed tasks

## Testing

### Backend
```bash
cd go-backend-bigmeter
go test ./...
```

### Frontend
No test suite currently configured. Use ESLint for code quality:
```bash
cd big-meter-frontend
pnpm lint
```

## Common Issues

### Database Connection Failed
Check PostgreSQL is running and env vars match:
```bash
docker-compose ps postgres
docker-compose logs postgres
docker-compose exec postgres pg_isready
```

### Oracle Sync Errors
Verify Oracle DSN and connectivity:
```bash
# Test Oracle connection
docker-compose exec sync env | grep ORACLE
# Or run ora-test mode
MODE=ora-test YM=202501 BRANCHES=BA01 docker-compose up sync
```

### Frontend 404 or Blank Page
Check nginx config and API proxy:
```bash
docker-compose logs frontend
docker exec bigmeter_frontend nginx -t
```

## Security Notes

- Never commit `.env` with real credentials
- Use strong passwords for PostgreSQL
- Oracle DSN contains sensitive credentials
- PWA Intranet login endpoint is hardcoded in Docker build args
- Frontend uses localStorage for auth tokens (`big-meter.auth.user`)

## Additional Documentation

- **Go Backend Guidelines**: `go-backend-bigmeter/go-backend-guidelines.md` - Comprehensive coding standards
- **Project Description**: `go-backend-bigmeter/docs/project_desc.md` - Sync logic and data model
- **API Spec**: `go-backend-bigmeter/docs/API-Spec.md` - Detailed API documentation
- **Frontend CLAUDE.md**: `big-meter-frontend/CLAUDE.md` - Frontend architecture and patterns

## Version Info

- **Go**: 1.25.0
- **Node/pnpm**: pnpm 10.17.0
- **PostgreSQL**: 17.5-alpine
- **React**: 19.0.0
- **Vite**: 5.4.8
- **Docker Compose**: v2.0+ required
