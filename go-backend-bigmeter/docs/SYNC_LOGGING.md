# Sync Operation Logging System

## Overview

The sync logging system provides comprehensive tracking and monitoring of all sync operations (yearly initialization and monthly synchronization) in the Big Meter application.

## Implementation Date

2025-10-06

## Features

### Core Functionality

1. **Automatic Logging**: Every sync operation is automatically logged to the database
2. **Real-time Monitoring**: Admin UI displays operation history with 10-second auto-refresh
3. **Detailed Metrics**: Tracks duration, record counts, success/failure status
4. **Audit Trail**: Records who triggered each operation (api/scheduler/manual)
5. **Error Tracking**: Captures error messages for failed operations

### What Gets Logged

For each sync operation:
- **Sync Type**: yearly_init or monthly_sync
- **Branch Code**: Which branch was synced
- **Time Metadata**: Year-month, fiscal year, debt year-month
- **Status**: success, error, or in_progress
- **Timestamps**: Started at, finished at
- **Performance**: Duration in milliseconds
- **Results**: Records upserted and zeroed
- **Error Details**: Error message if failed
- **Source**: Triggered by (api, scheduler, manual)

## Database Schema

### Table: bm_sync_logs

```sql
CREATE TABLE bm_sync_logs (
    id SERIAL PRIMARY KEY,
    sync_type VARCHAR(20) NOT NULL,
    branch_code VARCHAR(10) NOT NULL,
    year_month VARCHAR(6),
    fiscal_year INTEGER,
    debt_ym VARCHAR(6),
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    duration_ms INTEGER,
    records_upserted INTEGER,
    records_zeroed INTEGER,
    error_message TEXT,
    triggered_by VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sync_logs_branch_type ON bm_sync_logs(branch_code, sync_type);
CREATE INDEX idx_sync_logs_created_at ON bm_sync_logs(created_at DESC);
CREATE INDEX idx_sync_logs_status ON bm_sync_logs(status);
```

**Migration**: `migrations/0005_sync_logs.sql`

## Backend Implementation

### 1. Repository Layer

**File**: `internal/sync/log.go`

```go
type LogRepository struct {
    pool *pgxpool.Pool
}

// Core methods:
RecordSyncStart(ctx, syncType, branchCode, triggeredBy, yearMonth, debtYM, fiscalYear) (int64, error)
UpdateSyncSuccess(ctx, logID, upserted, zeroed) error
UpdateSyncError(ctx, logID, errorMsg) error
ListSyncLogs(ctx, filter) ([]SyncLog, int, error)
```

### 2. Service Integration

**File**: `internal/sync/service.go`

**Changes**:
- `InitCustcodes()` signature: `(ctx, fiscal, branch, debtYM, triggeredBy) (int, int, error)`
- `MonthlyDetails()` signature: `(ctx, ym, branch, batchSize, triggeredBy) (int, int, error)`

Both methods now:
1. Call `LogRepo.RecordSyncStart()` at the beginning
2. Return actual record counts `(upserted, zeroed, error)`
3. Call `LogRepo.UpdateSyncSuccess()` or `UpdateSyncError()` on completion

### 3. API Endpoints

**File**: `internal/api/server.go`

**Updated Endpoints**:

```
POST /api/v1/sync/init
Body: {"branches": ["BA01"], "debt_ym": "202410"}
Returns: {
  "fiscal_year": 2025,
  "branches": ["BA01"],
  "stats": {"upserted": 199, "zeroed": 0},
  "started_at": "2025-10-06T15:17:35Z",
  "finished_at": "2025-10-06T15:17:38Z"
}
```

```
POST /api/v1/sync/monthly
Body: {"branches": ["BA01"], "ym": "202509", "batch_size": 100}
Returns: {
  "ym": "202509",
  "branches": ["BA01"],
  "stats": {"upserted": 199, "zeroed": 5},
  "started_at": "2025-10-06T15:17:35Z",
  "finished_at": "2025-10-06T15:17:38Z"
}
```

**New Endpoint**:

```
GET /api/v1/sync/logs?branch=BA01&sync_type=monthly_sync&status=success&limit=20&offset=0
Returns: {
  "items": [
    {
      "id": 1,
      "sync_type": "monthly_sync",
      "branch_code": "1075",
      "year_month": "202509",
      "fiscal_year": 2025,
      "status": "success",
      "started_at": "2025-10-06T15:17:35Z",
      "finished_at": "2025-10-06T15:17:38Z",
      "duration_ms": 3192,
      "records_upserted": 199,
      "records_zeroed": 5,
      "triggered_by": "api",
      "created_at": "2025-10-06T15:17:35Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

### 4. CLI Integration

**File**: `cmd/sync/main.go`

All sync operations now pass `triggered_by` parameter:
- **Cron jobs**: `"scheduler"`
- **Manual runs** (init-once, month-once): `"manual"`

## Frontend Implementation

### 1. API Client

**File**: `src/api/syncLogs.ts`

```typescript
interface SyncLog {
  id: number;
  sync_type: string;
  branch_code: string;
  year_month?: string | null;
  fiscal_year?: number | null;
  debt_ym?: string | null;
  status: string;
  started_at: string;
  finished_at?: string | null;
  duration_ms?: number | null;
  records_upserted?: number | null;
  records_zeroed?: number | null;
  error_message?: string | null;
  triggered_by: string;
  created_at: string;
}

getSyncLogs(params: GetSyncLogsParams): Promise<SyncLogsResponse>
```

### 2. Admin UI

**File**: `src/screens/AdminPage.tsx`

**Features**:
- Displays last 20 sync operations in a table
- Auto-refreshes every 10 seconds via React Query
- Color-coded status badges:
  - Green: Success
  - Red: Error
  - Yellow: In Progress
- Shows:
  - Time (Thai locale)
  - Type (Yearly Init / Monthly Sync)
  - Branch code
  - Status
  - Upserted count
  - Zeroed count
  - Duration (seconds)
  - Triggered by
- Automatically refreshes after manual sync operations

## Usage Examples

### 1. Trigger Manual Sync and View Logs

```bash
# In admin UI:
1. Select branch(es)
2. Enter year-month
3. Click "เริ่มต้น Monthly Sync"
4. View results in success message
5. See new log entry appear in "Sync Operation Logs" table
```

### 2. Query Logs via API

```bash
# Get all logs
curl http://localhost:8089/api/v1/sync/logs

# Get logs for specific branch
curl http://localhost:8089/api/v1/sync/logs?branch=1075

# Get only failed operations
curl http://localhost:8089/api/v1/sync/logs?status=error

# Get monthly syncs only
curl http://localhost:8089/api/v1/sync/logs?sync_type=monthly_sync&limit=50
```

### 3. Monitor Scheduler Operations

Cron jobs automatically log with `triggered_by: "scheduler"`:

```sql
SELECT
  sync_type,
  branch_code,
  status,
  duration_ms / 1000.0 as duration_sec,
  records_upserted,
  started_at
FROM bm_sync_logs
WHERE triggered_by = 'scheduler'
ORDER BY started_at DESC
LIMIT 10;
```

## Benefits

1. **Operational Visibility**: See all sync operations at a glance
2. **Performance Monitoring**: Track operation durations and identify slow syncs
3. **Error Diagnosis**: Immediate access to error messages for failed operations
4. **Audit Trail**: Know who triggered each sync and when
5. **Capacity Planning**: Analyze record counts and sync patterns over time
6. **Debugging**: Quickly identify which operations succeeded/failed

## Troubleshooting

### Logs Not Appearing in UI

1. Check API endpoint: `curl http://localhost:8089/api/v1/sync/logs`
2. Verify database table exists: `\d bm_sync_logs` in psql
3. Check browser console for errors
4. Verify React Query is fetching: DevTools > Network tab

### Stats Showing Zero

This was the original issue that led to this implementation. The problem was:
- Old implementation: Services returned only `error`, no stats
- Solution: Modified service signatures to return `(upserted, zeroed, error)`

### Missing Logs for Scheduler Operations

1. Check sync service logs: `docker compose logs sync`
2. Verify LogRepository is initialized: Check for nil checks in service.go
3. Ensure migrations ran: `docker compose logs migrate`

## Future Enhancements

Potential improvements:
1. **Filtering UI**: Add filter dropdowns in admin page
2. **Pagination**: Add pagination controls for logs table
3. **Charts**: Visualize sync success rate and performance trends
4. **Alerts**: Email/Slack notifications for failed operations
5. **Retention Policy**: Auto-archive old logs (e.g., > 90 days)
6. **Export**: Download logs as CSV/Excel
7. **Detailed View**: Click log entry to see full details including error stack traces

## Performance Considerations

- **Indexes**: Already optimized with indexes on branch_code, sync_type, status, created_at
- **Query Limit**: Default 50, max 500 to prevent large result sets
- **Auto-refresh**: 10-second interval balances freshness vs. server load
- **Concurrent Writes**: Each branch sync creates separate log entry (safe for parallel execution)

## Security

- Logs endpoint requires admin authentication (same as sync triggers)
- No sensitive data logged (only operation metadata)
- Error messages may contain technical details but no credentials

## Maintenance

### Regular Tasks

1. **Monitor table size**: `SELECT pg_size_pretty(pg_total_relation_size('bm_sync_logs'));`
2. **Archive old logs** (optional): Move records > 90 days to archive table
3. **Analyze performance**: Review slow syncs and optimize if needed

### Backup Considerations

- Logs table is included in standard PostgreSQL backup
- Can be safely dropped and recreated (historical data only)
- Consider separate retention policy from operational data

## Related Files

### Backend
- `migrations/0005_sync_logs.sql` - Database schema
- `internal/sync/log.go` - Repository layer
- `internal/sync/service.go` - Service integration
- `internal/api/server.go` - API endpoints
- `cmd/sync/main.go` - CLI integration

### Frontend
- `src/api/syncLogs.ts` - API client
- `src/screens/AdminPage.tsx` - UI implementation

### Documentation
- `CLAUDE.md` - Updated with sync logging section
- `docs/SYNC_LOGGING.md` - This file
