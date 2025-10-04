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
export DOCKER_REGISTRY="your-registry.com/bigmeter"  # or docker.io/username

# Optional: Set image tag (default: latest)
export IMAGE_TAG="v1.0.0"  # or use git commit hash
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

# Build API service
docker build \
  -f docker/Dockerfile.api \
  -t ${DOCKER_REGISTRY}/bigmeter-api:${IMAGE_TAG:-latest} \
  .

# Build Sync service
docker build \
  -f docker/Dockerfile.sync-thick \
  -t ${DOCKER_REGISTRY}/bigmeter-sync:${IMAGE_TAG:-latest} \
  .

# Build Frontend
docker build \
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

# Copy migration files
scp go-backend-bigmeter/migrations/*.sql user@production-server:/opt/bigmeter/migrations/

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

# Metrics (Optional)
METRICS_ADDR=:9090
METRICS_PORT=9090
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
docker-compose -f docker-compose.prod.yml pull

# Verify images
docker images | grep bigmeter
```

### Step 2: Run Database Setup (First Time Only)

```bash
# Run migrations and seed branches
docker-compose -f docker-compose.prod.yml --profile setup up migrate seed_branches

# Wait for completion, then check logs
docker-compose -f docker-compose.prod.yml --profile setup logs migrate
docker-compose -f docker-compose.prod.yml --profile setup logs seed_branches
```

### Step 3: Start Application Services

```bash
# Start all services in background
docker-compose -f docker-compose.prod.yml up -d

# Check service status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

### Step 4: Verify Deployment

```bash
# Check API health
curl http://localhost:8089/api/v1/healthz

# Check frontend
curl http://localhost:3000

# Check database
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d bigmeter -c "\dt"

# Check sync service logs
docker-compose -f docker-compose.prod.yml logs sync
```

---

## Maintenance

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

# View sync service metrics (if enabled)
curl http://localhost:9090/metrics

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

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DOCKER_REGISTRY` | Yes | - | Docker registry URL |
| `IMAGE_TAG` | No | `latest` | Image version tag |
| `POSTGRES_USER` | Yes | `postgres` | PostgreSQL username |
| `POSTGRES_PASSWORD` | Yes | - | PostgreSQL password |
| `POSTGRES_DB` | Yes | `bigmeter` | Database name |
| `TIMEZONE` | No | `Asia/Bangkok` | Server timezone |
| `PORT` | No | `8089` | API server port |
| `ORACLE_DSN` | Yes (sync) | - | Oracle connection string |
| `BRANCHES` | Yes | - | Comma-separated branch codes |
| `MODE` | No | empty | Sync mode: `init-once`, `month-once`, `ora-test` |
| `SYNC_CONCURRENCY` | No | `2` | Number of concurrent sync jobs |
| `SYNC_RETRIES` | No | `2` | Number of retry attempts |
| `SYNC_RETRY_DELAY` | No | `10s` | Delay between retries |
| `BATCH_SIZE` | No | `100` | Batch size for Oracle queries |
| `METRICS_ADDR` | No | - | Prometheus metrics address |
| `METRICS_PORT` | No | `9090` | Prometheus metrics port |
| `ENABLE_YEARLY_INIT` | No | `true` | Enable yearly cohort init cron job |
| `ENABLE_MONTHLY_SYNC` | No | `true` | Enable monthly sync cron job |

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
