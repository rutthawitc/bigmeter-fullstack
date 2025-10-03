# Big Meter - Water Usage Monitoring System

A full-stack Dockerized application for monitoring and analyzing large-scale water usage data for the Provincial Waterworks Authority (PWA) of Thailand.

## 📋 Overview

**Big Meter** is a comprehensive water usage monitoring system that:
- Syncs data from Oracle database to PostgreSQL
- Tracks top-200 high-usage customers per branch
- Provides interactive dashboard for water consumption analysis
- Identifies significant usage decrease patterns
- Exports detailed reports to Excel

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Compose                        │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────┐      ┌──────────────┐                │
│  │   Frontend   │      │   Backend    │                │
│  │   (Nginx)    │─────▶│     API      │                │
│  │   Port 3000  │      │  (Go/Gin)    │                │
│  └──────────────┘      └──────┬───────┘                │
│         │                      │                         │
│         │              ┌───────┴────────┐               │
│         │              │   PostgreSQL   │               │
│         │              │   Port 5432    │               │
│         │              └───────▲────────┘               │
│         │                      │                         │
│         │              ┌───────┴────────┐               │
│         └─────────────▶│  Sync Service  │               │
│                        │  (Cron Jobs)   │               │
│                        └───────┬────────┘               │
│                                │                         │
│                        ┌───────▼────────┐               │
│                        │  Oracle DB     │               │
│                        │  (External)    │               │
│                        └────────────────┘               │
└─────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- **Docker** (version 20.10+)
- **Docker Compose** (version 2.0+)
- (Optional) **Oracle Database** connection for data sync

### 1. Clone the Repository

```bash
git clone <repository-url>
cd big-meter
```

### 2. Configure Environment

```bash
cd go-backend-bigmeter
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Database
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=bigmeter

# Timezone
TIMEZONE=Asia/Bangkok

# API Port
PORT=8089

# Oracle Connection (for sync service)
ORACLE_DSN=user/pass@host:1521/service_name

# Branch List (comma-separated)
BRANCHES=BA01,BA02,BA03
```

### 3. Run Database Setup

```bash
# Run migrations and seed branches
docker-compose --profile setup up migrate seed_branches

# Wait for completion, then Ctrl+C
```

### 4. Start All Services

```bash
# Build and start all services
docker-compose up --build

# Or run in background
docker-compose up -d --build
```

### 5. Access the Application

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8089 (internal, proxied through frontend)
- **PostgreSQL**: localhost:5432

## 🛠️ Project Structure

```
big-meter/
├── big-meter-frontend/          # React/TypeScript frontend
│   ├── src/                     # Source code
│   │   ├── api/                 # API client
│   │   ├── screens/             # Page components
│   │   ├── lib/                 # Utilities
│   │   └── styles/              # CSS
│   ├── Dockerfile               # Frontend Docker build
│   ├── nginx.conf               # Nginx configuration
│   ├── package.json             # Node dependencies
│   └── vite.config.ts           # Vite build config
│
├── go-backend-bigmeter/         # Go backend
│   ├── cmd/                     # Application entrypoints
│   │   ├── api/                 # REST API server
│   │   └── sync/                # Sync service
│   ├── internal/                # Internal packages
│   │   ├── api/                 # API handlers
│   │   ├── config/              # Configuration
│   │   ├── database/            # Database clients
│   │   └── sync/                # Sync logic
│   ├── docker/                  # Dockerfiles
│   │   ├── Dockerfile.api       # API service
│   │   └── Dockerfile.sync-thick # Sync service
│   ├── migrations/              # SQL migrations
│   ├── sqls/                    # SQL queries
│   ├── docs/                    # Documentation
│   ├── docker-compose.yml       # Docker Compose config
│   ├── go.mod                   # Go dependencies
│   └── .env.example             # Environment template
│
└── README.md                    # This file
```

## 📦 Services

### Frontend (Nginx + React)
- **Container**: `bigmeter_frontend`
- **Port**: 3000 (host) → 80 (container)
- **Technology**: React 19, TypeScript, Vite, Tailwind CSS
- **Features**:
  - Dashboard for water usage analysis
  - Interactive sparkline charts
  - Excel export functionality
  - Thai language interface
  - Responsive design

### Backend API (Go + Gin)
- **Container**: `bigmeter_api`
- **Port**: 8089 (internal)
- **Technology**: Go 1.25, Gin web framework
- **Endpoints**:
  - `GET /api/v1/healthz` - Health check
  - `GET /api/v1/branches` - List branches
  - `GET /api/v1/custcodes` - Customer codes
  - `GET /api/v1/details` - Usage details
  - `POST /api/v1/sync/init` - Trigger yearly sync
  - `POST /api/v1/sync/monthly` - Trigger monthly sync

### Sync Service
- **Container**: `bigmeter_sync`
- **Technology**: Go, Cron scheduler
- **Features**:
  - Yearly initialization (Oct 15, 22:00)
  - Monthly sync (16th of month, 08:00)
  - Oracle → PostgreSQL data sync
  - Top-200 customer tracking

### PostgreSQL Database
- **Container**: `bigmeter_postgres`
- **Port**: 5432
- **Version**: PostgreSQL 17.5
- **Tables**:
  - `bm_branches` - Branch metadata
  - `bm_custcode_init` - Customer cohort
  - `bm_meter_details` - Monthly usage data

## 🔧 Docker Commands

### Start Services
```bash
cd go-backend-bigmeter

# Start all services
docker-compose up

# Start in background
docker-compose up -d

# Start specific services
docker-compose up postgres api frontend
```

### Stop Services
```bash
# Stop all services
docker-compose down

# Stop and remove volumes (⚠️ deletes data)
docker-compose down -v
```

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f frontend
docker-compose logs -f api
docker-compose logs -f sync
```

### Rebuild Services
```bash
# Rebuild all
docker-compose build

# Rebuild specific service
docker-compose build frontend
docker-compose build api

# Rebuild and restart
docker-compose up -d --build
```

### Database Operations
```bash
# Run migrations
docker-compose --profile setup up migrate

# Seed branches
docker-compose --profile setup up seed_branches

# Access PostgreSQL
docker exec -it bigmeter_postgres psql -U postgres -d bigmeter

# Backup database
docker exec bigmeter_postgres pg_dump -U postgres bigmeter > backup.sql

# Restore database
docker exec -i bigmeter_postgres psql -U postgres bigmeter < backup.sql
```

## 🔒 Environment Variables

### Required
- `POSTGRES_USER` - PostgreSQL username
- `POSTGRES_PASSWORD` - PostgreSQL password
- `POSTGRES_DB` - Database name

### Optional
- `TIMEZONE` - Server timezone (default: `Asia/Bangkok`)
- `PORT` - API port (default: `8089`)
- `BRANCHES` - Comma-separated branch codes
- `ORACLE_DSN` - Oracle connection string
- `MODE` - Sync mode: empty (scheduler), `init-once`, `month-once`, `ora-test`
- `YM` - Year-month for manual sync (YYYYMM format)

## 🧪 Development

### Frontend Development
```bash
cd big-meter-frontend

# Install dependencies
pnpm install

# Run dev server (without Docker)
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview
```

### Backend Development
```bash
cd go-backend-bigmeter

# Run API locally
go run cmd/api/main.go

# Run sync service locally
go run cmd/sync/main.go

# Run tests
go test ./...
```

### Hot Reload with Docker
```bash
# Use docker-compose watch (Docker Compose v2.22+)
docker-compose watch

# Or use bind mounts in docker-compose.override.yml
```

## 📊 Database Schema

### bm_branches
```sql
- code (PK)        - Branch code
- name             - Branch name
- created_at       - Creation timestamp
```

### bm_custcode_init
```sql
- fiscal_year (PK) - Fiscal year
- branch_code (PK) - Branch code
- cust_code (PK)   - Customer code
- org_name         - Organization name
- use_type         - Usage type
- meter_no         - Meter number
- ... (more fields)
```

### bm_meter_details
```sql
- year_month (PK)       - YYYYMM format
- branch_code (PK)      - Branch code
- cust_code (PK)        - Customer code
- present_water_usg     - Water usage (m³)
- present_meter_count   - Meter reading
- average               - Average usage
- ... (more fields)
```

## 🚨 Troubleshooting

### Frontend not loading
```bash
# Check nginx logs
docker-compose logs frontend

# Verify nginx config
docker exec bigmeter_frontend nginx -t

# Restart frontend
docker-compose restart frontend
```

### API connection errors
```bash
# Check API logs
docker-compose logs api

# Verify database connection
docker-compose exec api env | grep POSTGRES

# Test database connectivity
docker-compose exec postgres pg_isready
```

### Database connection failed
```bash
# Check PostgreSQL status
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Sync service not running
```bash
# Check sync logs
docker-compose logs sync

# Verify Oracle connection
docker-compose exec sync env | grep ORACLE

# Test Oracle connectivity (requires ORACLE_DSN)
docker-compose exec sync go run cmd/sync/main.go -mode ora-test
```

## 📝 API Documentation

See `go-backend-bigmeter/docs/API-Spec.md` for detailed API documentation.

## 🔐 Security Notes

- Never commit `.env` files with real credentials
- Use strong passwords for PostgreSQL
- Update `nginx.conf` with proper CSP headers for production
- Enable HTTPS in production
- Restrict PostgreSQL port exposure in production
- Use Docker secrets for sensitive data in production

## 📄 License

[Add your license here]

## 👥 Contributors

- PWA Development Team

## 📞 Support

For issues and questions:
- Check `go-backend-bigmeter/docs/` for detailed documentation
- Review `big-meter-frontend/CLAUDE.md` for frontend architecture
- Open an issue in the repository

---

**Last Updated**: 2025-10-03
**Version**: 1.0.0
