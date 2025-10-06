-- Sync logs table for tracking yearly init and monthly sync operations
CREATE TABLE IF NOT EXISTS bm_sync_logs (
    id SERIAL PRIMARY KEY,
    sync_type VARCHAR(20) NOT NULL,  -- 'yearly_init' or 'monthly_sync'
    branch_code VARCHAR(10) NOT NULL,
    year_month VARCHAR(6),            -- YYYYMM (null for yearly_init)
    fiscal_year INTEGER,              -- fiscal year (null for monthly_sync)
    debt_ym VARCHAR(6),               -- Thai Buddhist YYYYMM (for yearly_init)
    status VARCHAR(20) NOT NULL,      -- 'success', 'error', 'in_progress'
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    duration_ms INTEGER,              -- calculated: finished - started
    records_upserted INTEGER,
    records_zeroed INTEGER,
    error_message TEXT,
    triggered_by VARCHAR(50),         -- 'scheduler', 'manual', 'api'
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for querying by branch and type
CREATE INDEX idx_sync_logs_branch_type ON bm_sync_logs(branch_code, sync_type);

-- Index for sorting by date (most common query)
CREATE INDEX idx_sync_logs_created_at ON bm_sync_logs(created_at DESC);

-- Index for filtering by status
CREATE INDEX idx_sync_logs_status ON bm_sync_logs(status);
