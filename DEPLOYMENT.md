# Big Meter - Production Deployment Guide

This guide explains how to deploy Big Meter to a production server using pre-built Docker images.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Build and Push Images](#build-and-push-images)
3. [Production Server Setup](#production-server-setup)
4. [Running the Application](#running-the-application)
5. [Maintenance](#maintenance)
6. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### On Build Server (CI/CD or Development Machine)

- Docker installed
- Access to source code repository
- Access to Docker registry (Docker Hub, AWS ECR, Google GCR, etc.)

### On Production Server

- Docker installed (version 20.10+)
- Docker Compose installed (version 2.0+)
- Network access to:
  - Docker registry (to pull images)
  - Oracle database (for sync service)
  - PWA Intranet (for authentication)

---

## Build and Push Images

### Step 1: Set Registry Configuration

```bash
# Export your Docker registry URL
export DOCKER_REGISTRY="your-registry.com"  # or docker.io/username

# Optional: Set image tag (default: latest)
export IMAGE_TAG="v1.0.0"  # or use git commit hash

# Set admin usernames (comma-separated, required for admin page access)
export VITE_ADMIN_USERNAMES="admin,user1,user2"
```

### Step 2: Login to Docker Registry

```bash
# Docker Hub
docker login

# AWS ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 123456789012.dkr.ecr.us-east-1.amazonaws.com

# Google Container Registry
gcloud auth configure-docker

# Azure Container Registry
az acr login --name myregistry
```

### Step 3: Build Images

```bash
cd go-backend-bigmeter

# Build API service (with Oracle support)
docker build \
  -f docker/Dockerfile.api-thick \
  -t ${DOCKER_REGISTRY}/bigmeter-api:${IMAGE_TAG:-latest} \
  .

# Build Sync service
docker build \
  -f docker/Dockerfile.sync-thick \
  -t ${DOCKER_REGISTRY}/bigmeter-sync:${IMAGE_TAG:-latest} \
  .

# Build Frontend (with admin usernames)
docker build \
  --build-arg VITE_ADMIN_USERNAMES=${VITE_ADMIN_USERNAMES:-admin} \
  -t ${DOCKER_REGISTRY}/bigmeter-frontend:${IMAGE_TAG:-latest} \
  ../big-meter-frontend
```

### Step 4: Push Images to Registry

```bash
docker push ${DOCKER_REGISTRY}/bigmeter-api:${IMAGE_TAG:-latest}
docker push ${DOCKER_REGISTRY}/bigmeter-sync:${IMAGE_TAG:-latest}
docker push ${DOCKER_REGISTRY}/bigmeter-frontend:${IMAGE_TAG:-latest}
```

### Step 5: Verify Images

```bash
# List pushed images
docker images | grep bigmeter

# Test pull (optional)
docker pull ${DOCKER_REGISTRY}/bigmeter-api:${IMAGE_TAG:-latest}
```

---

## Production Server Setup

### Step 1: Create Application Directory

```bash
# SSH to production server
ssh user@production-server

# Create deployment directory
mkdir -p /opt/bigmeter
cd /opt/bigmeter
```

### Step 2: Prepare Required Files

#### Download docker-compose.prod.yml

```bash
# Option 1: Download from repository
curl -O https://raw.githubusercontent.com/your-org/big-meter/main/docker-compose.prod.yml

# Option 2: Copy from local machine
scp docker-compose.prod.yml user@production-server:/opt/bigmeter/

# Option 3: Create manually (see docker-compose.prod.yml in repo)
nano docker-compose.prod.yml
```

#### Create Migrations Directory

```bash
mkdir -p migrations seed-data

# Option 1: Copy all migration files (for migration-based approach)
scp go-backend-bigmeter/migrations/*.sql user@production-server:/opt/bigmeter/migrations/

# Option 2: Copy only init_complete.sql (for fresh install)
scp go-backend-bigmeter/migrations/init_complete.sql user@production-server:/opt/bigmeter/migrations/

# Copy branch seed data
scp go-backend-bigmeter/docs/r6_branches.csv user@production-server:/opt/bigmeter/seed-data/
```

### Step 3: Create Environment File

```bash
cat > .env <<'EOF'
# Docker Registry
DOCKER_REGISTRY=your-registry.com/bigmeter
IMAGE_TAG=latest

# Database Configuration
POSTGRES_USER=postgres
POSTGRES_PASSWORD=CHANGE_THIS_SECURE_PASSWORD
POSTGRES_DB=bigmeter

# Application Configuration
TIMEZONE=Asia/Bangkok
PORT=8089

# Oracle Connection (for sync service)
ORACLE_DSN=username/password@oracle-host:1521/service_name

# Branch Configuration
BRANCHES=BA01,BA02,BA03,BA04,BA05

# Sync Configuration (Optional)
MODE=
SYNC_CONCURRENCY=2
SYNC_RETRIES=2
SYNC_RETRY_DELAY=10s
BATCH_SIZE=100
ENABLE_YEARLY_INIT=true
ENABLE_MONTHLY_SYNC=true

# Telegram Notifications (Optional)
TELEGRAM_ENABLED=true
TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
TELEGRAM_CHAT_ID=-1001234567890
EOF

# Secure the .env file
chmod 600 .env
```

### Step 4: Login to Docker Registry

```bash
# Docker Hub
docker login

# AWS ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 123456789012.dkr.ecr.us-east-1.amazonaws.com

# Or use credentials stored in environment
echo $DOCKER_PASSWORD | docker login -u $DOCKER_USERNAME --password-stdin
```

---

## Running the Application

### Step 1: Pull Images

```bash
cd /opt/bigmeter

# Pull all images
docker compose -f docker-compose.prod.yml pull

# Verify images
docker images | grep bigmeter
```

### Step 2: Run Database Setup (First Time Only)

**Option A: Using docker-compose migrate service (runs all migrations)**

```bash
# Run migrations and seed branches
docker compose -f docker-compose.prod.yml --profile setup up migrate seed_branches

# Wait for completion, then check logs
docker compose -f docker-compose.prod.yml --profile setup logs migrate
docker compose -f docker-compose.prod.yml --profile setup logs seed_branches
```

**Option B: Using init_complete.sql (fresh install, all-in-one)**

```bash
# Run complete initialization script
docker compose -f docker-compose.prod.yml exec -T postgres \
  psql -U postgres -d bigmeter < migrations/init_complete.sql

# Seed branches
docker compose -f docker-compose.prod.yml --profile setup up seed_branches

# Verify tables exist
docker compose -f docker-compose.prod.yml exec -T postgres \
  psql -U postgres -d bigmeter -c "\dt bm_*"
```

**Note:**
- Option A: Runs migrations 0001 → 0006 sequentially (recommended for upgrades)
- Option B: Creates all tables at once with final schema (recommended for fresh installs)
- Both result in the same database schema including fiscal_year column

### Step 3: Start Application Services

```bash
# Start all services in background
docker compose -f docker-compose.prod.yml up -d

# Check service status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs -f
```

### Step 4: Verify Deployment

```bash
# Check API health
curl http://localhost:8089/api/v1/healthz

# Check frontend
curl http://localhost:3000

# Check database
docker compose -f docker-compose.prod.yml exec postgres psql -U postgres -d bigmeter -c "\dt"

# Check sync service logs
docker compose -f docker-compose.prod.yml logs sync
```

---

## Maintenance

### Updating Existing Database Schema

**If you already have a running database and need to add fiscal_year column:**

```bash
cd /opt/bigmeter

# Download migration 0006
curl -o migrations/0006_add_fiscal_year_to_details.sql \
  https://raw.githubusercontent.com/rutthawitc/bigmeter-fullstack/main/go-backend-bigmeter/migrations/0006_add_fiscal_year_to_details.sql

# Run the migration
docker compose -f docker-compose.prod.yml exec -T postgres \
  psql -U postgres -d bigmeter < migrations/0006_add_fiscal_year_to_details.sql

# Verify fiscal_year column was added
docker compose -f docker-compose.prod.yml exec -T postgres \
  psql -U postgres -d bigmeter -c "\d bm_meter_details"
```

**Important:** After running this migration, you must rebuild and redeploy the API service with the updated code.

### Viewing Logs

```bash
# All services
docker-compose -f docker-compose.prod.yml logs -f

# Specific service
docker-compose -f docker-compose.prod.yml logs -f api
docker-compose -f docker-compose.prod.yml logs -f sync
docker-compose -f docker-compose.prod.yml logs -f frontend

# Last N lines
docker-compose -f docker-compose.prod.yml logs --tail=100 api
```

### Updating to New Version

```bash
cd /opt/bigmeter

# Update .env with new tag
sed -i 's/IMAGE_TAG=.*/IMAGE_TAG=v1.1.0/' .env

# Pull new images
docker-compose -f docker-compose.prod.yml pull

# Restart services with zero downtime (if using load balancer)
docker-compose -f docker-compose.prod.yml up -d --no-deps api
docker-compose -f docker-compose.prod.yml up -d --no-deps sync
docker-compose -f docker-compose.prod.yml up -d --no-deps frontend

# Or restart all at once
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d
```

### Backup Database

```bash
# Create backup directory
mkdir -p /opt/bigmeter/backups

# Backup database
docker-compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U postgres bigmeter > backups/bigmeter_$(date +%Y%m%d_%H%M%S).sql

# Compress backup
gzip backups/bigmeter_$(date +%Y%m%d_%H%M%S).sql

# Automated daily backup (add to crontab)
0 2 * * * cd /opt/bigmeter && docker-compose -f docker-compose.prod.yml exec -T postgres pg_dump -U postgres bigmeter | gzip > backups/bigmeter_$(date +\%Y\%m\%d).sql.gz
```

### Restore Database

```bash
# Stop application services (keep postgres running)
docker-compose -f docker-compose.prod.yml stop api sync frontend

# Restore from backup
gunzip < backups/bigmeter_20250104.sql.gz | \
  docker-compose -f docker-compose.prod.yml exec -T postgres \
  psql -U postgres bigmeter

# Restart services
docker-compose -f docker-compose.prod.yml start api sync frontend
```

### Manual Sync Operations

```bash
# Trigger yearly init manually
docker-compose -f docker-compose.prod.yml exec api curl -X POST \
  http://localhost:8089/api/v1/sync/init \
  -H "Content-Type: application/json" \
  -d '{"fiscal_year": 2025, "branch_code": "BA01", "debt_ym": "256710"}'

# Trigger monthly sync manually
docker-compose -f docker-compose.prod.yml exec api curl -X POST \
  http://localhost:8089/api/v1/sync/monthly \
  -H "Content-Type: application/json" \
  -d '{"year_month": "202501", "branch_code": "BA01", "batch_size": 100}'

# Run one-off sync (init-once mode)
docker-compose -f docker-compose.prod.yml run --rm \
  -e MODE=init-once \
  -e YM=202410 \
  sync

# Run one-off sync (month-once mode)
docker-compose -f docker-compose.prod.yml run --rm \
  -e MODE=month-once \
  -e YM=202501 \
  sync
```

### Stopping Services

```bash
# Stop all services
docker-compose -f docker-compose.prod.yml stop

# Stop specific service
docker-compose -f docker-compose.prod.yml stop sync

# Stop and remove containers (data persists in volumes)
docker-compose -f docker-compose.prod.yml down

# Stop and remove everything including volumes (⚠️ deletes data)
docker-compose -f docker-compose.prod.yml down -v
```

---

## Troubleshooting

### Services Won't Start

```bash
# Check service status
docker-compose -f docker-compose.prod.yml ps

# Check logs for errors
docker-compose -f docker-compose.prod.yml logs api
docker-compose -f docker-compose.prod.yml logs sync

# Verify environment variables
docker-compose -f docker-compose.prod.yml config

# Check resource usage
docker stats
```

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker-compose -f docker-compose.prod.yml exec postgres pg_isready

# Test connection from api container
docker-compose -f docker-compose.prod.yml exec api env | grep POSTGRES

# Check PostgreSQL logs
docker-compose -f docker-compose.prod.yml logs postgres
```

### Oracle Connection Issues

```bash
# Check Oracle DSN configuration
docker-compose -f docker-compose.prod.yml exec sync env | grep ORACLE

# Test Oracle connectivity
docker-compose -f docker-compose.prod.yml run --rm \
  -e MODE=ora-test \
  -e YM=202501 \
  -e BRANCHES=BA01 \
  sync
```

### Frontend Not Loading

```bash
# Check nginx status
docker-compose -f docker-compose.prod.yml exec frontend nginx -t

# Check frontend logs
docker-compose -f docker-compose.prod.yml logs frontend

# Verify API connectivity from frontend container
docker-compose -f docker-compose.prod.yml exec frontend wget -O- http://api:8089/api/v1/healthz
```

### Disk Space Issues

```bash
# Check disk usage
df -h

# Clean up unused Docker resources
docker system prune -a

# Remove old images
docker images | grep bigmeter
docker rmi your-registry/bigmeter-api:old-tag

# Check volume sizes
docker system df -v
```

### Performance Issues

```bash
# Monitor container resources
docker stats

# Check API response time
time curl http://localhost:8089/api/v1/healthz

# Check database connections
docker-compose -f docker-compose.prod.yml exec postgres \
  psql -U postgres -d bigmeter -c "SELECT count(*) FROM pg_stat_activity;"
```

---

## Production Checklist

### Pre-Deployment

- [ ] Build and test all Docker images
- [ ] Push images to registry
- [ ] Verify image tags and registry access
- [ ] Prepare migration files
- [ ] Create `.env` with production credentials
- [ ] Test Oracle connectivity from production network

### Deployment

- [ ] Create deployment directory structure
- [ ] Copy `docker-compose.prod.yml` and migrations
- [ ] Configure environment variables
- [ ] Pull Docker images
- [ ] Run database migrations
- [ ] Seed branch data
- [ ] Start all services
- [ ] Verify health endpoints

### Post-Deployment

- [ ] Configure firewall rules (ports 3000, 8089, 5432)
- [ ] Set up reverse proxy (nginx/traefik) with SSL/TLS
- [ ] Configure log rotation
- [ ] Set up monitoring and alerting
- [ ] Schedule automated backups
- [ ] Document emergency procedures
- [ ] Test disaster recovery process

### Security

- [ ] Use strong passwords for all services
- [ ] Restrict PostgreSQL port (remove public exposure)
- [ ] Enable SSL for PostgreSQL connections
- [ ] Use Docker secrets for sensitive data (optional)
- [ ] Configure firewall and security groups
- [ ] Enable automatic security updates
- [ ] Set up SSL/TLS certificates for frontend
- [ ] Review and harden nginx configuration

---

## Environment Variables Reference

| Variable                   | Required   | Default        | Description                                      |
| -------------------------- | ---------- | -------------- | ------------------------------------------------ |
| `DOCKER_REGISTRY`          | Yes        | -              | Docker registry URL                              |
| `IMAGE_TAG`                | No         | `latest`       | Image version tag                                |
| `VITE_ADMIN_USERNAMES`     | No         | `admin`        | Comma-separated admin usernames (build-time)     |
| `POSTGRES_USER`            | Yes        | `postgres`     | PostgreSQL username                              |
| `POSTGRES_PASSWORD`        | Yes        | -              | PostgreSQL password                              |
| `POSTGRES_DB`              | Yes        | `bigmeter`     | Database name                                    |
| `TIMEZONE`                 | No         | `Asia/Bangkok` | Server timezone                                  |
| `PORT`                     | No         | `8089`         | API server port                                  |
| `ORACLE_DSN`               | Yes (sync) | -              | Oracle connection string                         |
| `BRANCHES`                 | Yes        | -              | Comma-separated branch codes                     |
| `MODE`                     | No         | empty          | Sync mode: `init-once`, `month-once`, `ora-test` |
| `SYNC_CONCURRENCY`         | No         | `2`            | Number of concurrent sync jobs                   |
| `SYNC_RETRIES`             | No         | `2`            | Number of retry attempts                         |
| `SYNC_RETRY_DELAY`         | No         | `10s`          | Delay between retries                            |
| `BATCH_SIZE`               | No         | `100`          | Batch size for Oracle queries                    |
| `ENABLE_YEARLY_INIT`       | No         | `true`         | Enable yearly cohort init cron job               |
| `ENABLE_MONTHLY_SYNC`      | No         | `true`         | Enable monthly sync cron job                     |
| `TELEGRAM_ENABLED`         | No         | `false`        | Enable Telegram notifications                    |
| `TELEGRAM_BOT_TOKEN`       | No         | -              | Telegram bot API token                           |
| `TELEGRAM_CHAT_ID`         | No         | `0`            | Telegram chat/group ID (negative for groups)     |
| `TELEGRAM_YEARLY_PREFIX`   | No         | Default        | Prefix for yearly sync messages                  |
| `TELEGRAM_MONTHLY_PREFIX`  | No         | Default        | Prefix for monthly sync messages                 |
| `TELEGRAM_YEARLY_SUCCESS`  | No         | Default        | Success message template for yearly              |
| `TELEGRAM_YEARLY_FAILURE`  | No         | Default        | Failure message template for yearly              |
| `TELEGRAM_MONTHLY_SUCCESS` | No         | Default        | Success message template for monthly             |
| `TELEGRAM_MONTHLY_FAILURE` | No         | Default        | Failure message template for monthly             |

---

## Telegram Notifications Setup

### Step 1: Create Telegram Bot

1. Open Telegram and search for **@BotFather**
2. Send `/newbot` command
3. Follow the prompts to create your bot
4. Copy the **bot token** (e.g., `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`)

### Step 2: Get Chat ID

**For Private Chat:**

1. Start a chat with your bot
2. Send any message to the bot
3. Visit: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
4. Look for `"chat":{"id":123456789}` in the JSON response

**For Group Chat:**

1. Create a group and add your bot
2. Send a message in the group
3. Visit: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
4. Look for `"chat":{"id":-1001234567890}` (negative number for groups)

### Step 3: Configure Environment

```bash
# Enable Telegram notifications
TELEGRAM_ENABLED=true
TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
TELEGRAM_CHAT_ID=-1001234567890
```

### Step 4: Customize Messages (Optional)

You can customize notification messages using placeholders:

**Available Placeholders:**

- Yearly: `{fiscal_year}`, `{branches}`, `{count}`, `{duration}`, `{timestamp}`, `{failed_branches}`, `{error}`
- Monthly: `{year_month}`, `{branches}`, `{count}`, `{duration}`, `{timestamp}`, `{failed_branches}`, `{error}`

**Example Custom Messages:**

```bash
# Thai language notifications
TELEGRAM_YEARLY_PREFIX="🔄 <b>Big Meter - การซิงค์รายปี</b>"
TELEGRAM_YEARLY_SUCCESS="✅ การซิงค์รายปีเสร็จสมบูรณ์\nปีงบประมาณ: {fiscal_year}\nสาขา: {count} แห่ง\nระยะเวลา: {duration}"
TELEGRAM_YEARLY_FAILURE="❌ การซิงค์รายปีล้มเหลว\nปีงบประมาณ: {fiscal_year}\nสาขาที่ล้มเหลว: {failed_branches}\nข้อผิดพลาด: {error}"

TELEGRAM_MONTHLY_PREFIX="📊 <b>Big Meter - การซิงค์รายเดือน</b>"
TELEGRAM_MONTHLY_SUCCESS="✅ การซิงค์รายเดือนเสร็จสมบูรณ์\nเดือน: {year_month}\nสาขา: {count} แห่ง\nระยะเวลา: {duration}"
TELEGRAM_MONTHLY_FAILURE="❌ การซิงค์รายเดือนล้มเหลว\nเดือน: {year_month}\nสาขาที่ล้มเหลว: {failed_branches}\nข้อผิดพลาด: {error}"
```

### Step 5: Test Notifications

```bash
# Run a one-off sync to test notifications
docker-compose -f docker-compose.prod.yml run --rm \
  -e MODE=month-once \
  -e YM=202501 \
  sync
```

### Notification Examples

**Success Notification:**

```
🔄 Big Meter - Yearly Sync

✅ Yearly cohort init completed successfully
Fiscal Year: 2025
Branches: 3 (BA01, BA02, BA03)
Duration: 45.2s
Time: 2025-10-15 22:00:15
```

**Failure Notification:**

```
🔄 Big Meter - Yearly Sync

❌ Yearly cohort init failed
Fiscal Year: 2025
Failed Branches: BA02, BA03
Error: connection timeout
Time: 2025-10-15 22:05:30
```

### Troubleshooting Telegram

```bash
# Check if Telegram is enabled
docker compose -f docker-compose.prod.yml exec sync env | grep TELEGRAM

# View notification logs
docker compose -f docker-compose.prod.yml logs sync | grep telegram

# Common issues:
# - Bot token invalid: Check BotFather for correct token
# - Chat ID wrong: Use negative number for groups
# - Bot not in group: Add bot to group and make it admin
# - Network issues: Check firewall allows https://api.telegram.org
```

---

## Support

For issues and questions:

- Check logs: `docker-compose -f docker-compose.prod.yml logs`
- Review documentation in `go-backend-bigmeter/docs/`
- Consult `CLAUDE.md` for architecture details
- Open an issue in the repository

---

**Last Updated**: 2025-01-04
**Version**: 1.0.0
