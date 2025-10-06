package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncLog represents a sync operation log entry
type SyncLog struct {
	ID             int64      `json:"id"`
	SyncType       string     `json:"sync_type"`
	BranchCode     string     `json:"branch_code"`
	YearMonth      *string    `json:"year_month,omitempty"`
	FiscalYear     *int       `json:"fiscal_year,omitempty"`
	DebtYM         *string    `json:"debt_ym,omitempty"`
	Status         string     `json:"status"`
	StartedAt      time.Time  `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	DurationMs     *int       `json:"duration_ms,omitempty"`
	RecordsUpserted *int      `json:"records_upserted,omitempty"`
	RecordsZeroed   *int      `json:"records_zeroed,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	TriggeredBy    string     `json:"triggered_by"`
	CreatedAt      time.Time  `json:"created_at"`
}

// LogRepository handles sync log persistence
type LogRepository struct {
	pool *pgxpool.Pool
}

// NewLogRepository creates a new log repository
func NewLogRepository(pool *pgxpool.Pool) *LogRepository {
	return &LogRepository{pool: pool}
}

// RecordSyncStart creates a new sync log entry with in_progress status
func (r *LogRepository) RecordSyncStart(ctx context.Context, syncType, branchCode, triggeredBy string, yearMonth, debtYM *string, fiscalYear *int) (int64, error) {
	query := `INSERT INTO bm_sync_logs (sync_type, branch_code, year_month, fiscal_year, debt_ym, status, started_at, triggered_by)
	          VALUES ($1, $2, $3, $4, $5, 'in_progress', $6, $7)
	          RETURNING id`

	var logID int64
	err := r.pool.QueryRow(ctx, query, syncType, branchCode, yearMonth, fiscalYear, debtYM, time.Now(), triggeredBy).Scan(&logID)
	if err != nil {
		return 0, fmt.Errorf("insert sync log start: %w", err)
	}
	return logID, nil
}

// UpdateSyncSuccess updates the log entry with success status and stats
func (r *LogRepository) UpdateSyncSuccess(ctx context.Context, logID int64, upserted, zeroed int) error {
	now := time.Now()
	query := `UPDATE bm_sync_logs
	          SET status = 'success',
	              finished_at = $2,
	              duration_ms = EXTRACT(EPOCH FROM ($2 - started_at)) * 1000,
	              records_upserted = $3,
	              records_zeroed = $4
	          WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, logID, now, upserted, zeroed)
	if err != nil {
		return fmt.Errorf("update sync log success: %w", err)
	}
	return nil
}

// UpdateSyncError updates the log entry with error status and message
func (r *LogRepository) UpdateSyncError(ctx context.Context, logID int64, errorMsg string) error {
	now := time.Now()
	query := `UPDATE bm_sync_logs
	          SET status = 'error',
	              finished_at = $2,
	              duration_ms = EXTRACT(EPOCH FROM ($2 - started_at)) * 1000,
	              error_message = $3
	          WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, logID, now, errorMsg)
	if err != nil {
		return fmt.Errorf("update sync log error: %w", err)
	}
	return nil
}

// ListSyncLogsFilter defines filters for listing sync logs
type ListSyncLogsFilter struct {
	BranchCode *string
	SyncType   *string
	Status     *string
	Limit      int
	Offset     int
}

// ListSyncLogs retrieves sync logs with optional filtering and pagination
func (r *LogRepository) ListSyncLogs(ctx context.Context, filter ListSyncLogsFilter) ([]SyncLog, int, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if filter.BranchCode != nil && *filter.BranchCode != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("branch_code = $%d", argIdx))
		args = append(args, *filter.BranchCode)
		argIdx++
	}
	if filter.SyncType != nil && *filter.SyncType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("sync_type = $%d", argIdx))
		args = append(args, *filter.SyncType)
		argIdx++
	}
	if filter.Status != nil && *filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			whereClause += " AND " + whereClauses[i]
		}
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM bm_sync_logs " + whereClause
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sync logs: %w", err)
	}

	// Query logs
	query := fmt.Sprintf(`SELECT id, sync_type, branch_code, year_month, fiscal_year, debt_ym, status,
	                             started_at, finished_at, duration_ms, records_upserted, records_zeroed,
	                             error_message, triggered_by, created_at
	                      FROM bm_sync_logs %s
	                      ORDER BY created_at DESC
	                      LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query sync logs: %w", err)
	}
	defer rows.Close()

	logs := []SyncLog{}
	for rows.Next() {
		var log SyncLog
		if err := rows.Scan(
			&log.ID, &log.SyncType, &log.BranchCode, &log.YearMonth, &log.FiscalYear, &log.DebtYM,
			&log.Status, &log.StartedAt, &log.FinishedAt, &log.DurationMs,
			&log.RecordsUpserted, &log.RecordsZeroed, &log.ErrorMessage,
			&log.TriggeredBy, &log.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan sync log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
