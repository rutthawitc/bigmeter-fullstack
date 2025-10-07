-- Complete database initialization script for Big Meter
-- This combines all migrations in order: 0001 through 0005
-- Run this on a fresh database to set up all tables and indexes

-- =============================================================================
-- 0001_init.sql - Main tables
-- =============================================================================

CREATE TABLE IF NOT EXISTS bm_branches (
  code VARCHAR(10) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bm_custcode_init (
  fiscal_year INTEGER NOT NULL,
  branch_code VARCHAR(10) NOT NULL,
  cust_code VARCHAR(50) NOT NULL,
  org_name VARCHAR(255),
  use_type VARCHAR(20),
  meter_no VARCHAR(50),
  meter_state VARCHAR(20),
  debt_ym VARCHAR(6),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (fiscal_year, branch_code, cust_code)
);

CREATE TABLE IF NOT EXISTS bm_meter_details (
  year_month VARCHAR(6) NOT NULL,
  branch_code VARCHAR(10) NOT NULL,
  cust_code VARCHAR(50) NOT NULL,
  org_name VARCHAR(255),
  use_type VARCHAR(20),
  use_name VARCHAR(100),
  cust_name VARCHAR(255),
  address TEXT,
  meter_no VARCHAR(50),
  water_vol NUMERIC(18,2),
  present_water_usg NUMERIC(18,2),
  average NUMERIC(18,2),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (year_month, branch_code, cust_code)
);

-- =============================================================================
-- 0002_branches.sql - Branch metadata
-- =============================================================================

COMMENT ON TABLE bm_branches IS 'Provincial Waterworks Authority branches';
COMMENT ON TABLE bm_custcode_init IS 'Yearly initialized top-200 customer cohorts';
COMMENT ON TABLE bm_meter_details IS 'Monthly water usage details';

-- =============================================================================
-- 0003_indexes.sql - Performance indexes
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_custcode_init_branch
  ON bm_custcode_init(branch_code);

CREATE INDEX IF NOT EXISTS idx_custcode_init_fiscal_branch
  ON bm_custcode_init(fiscal_year, branch_code);

CREATE INDEX IF NOT EXISTS idx_meter_details_branch
  ON bm_meter_details(branch_code);

CREATE INDEX IF NOT EXISTS idx_meter_details_ym
  ON bm_meter_details(year_month);

CREATE INDEX IF NOT EXISTS idx_meter_details_ym_branch
  ON bm_meter_details(year_month, branch_code);

-- =============================================================================
-- 0004_extend_custcode_init.sql - Extended customer fields
-- =============================================================================

ALTER TABLE bm_custcode_init
  ADD COLUMN IF NOT EXISTS org_name VARCHAR(255);

ALTER TABLE bm_custcode_init
  ADD COLUMN IF NOT EXISTS use_type VARCHAR(20);

ALTER TABLE bm_custcode_init
  ADD COLUMN IF NOT EXISTS meter_no VARCHAR(50);

ALTER TABLE bm_custcode_init
  ADD COLUMN IF NOT EXISTS meter_state VARCHAR(20);

ALTER TABLE bm_custcode_init
  ADD COLUMN IF NOT EXISTS debt_ym VARCHAR(6);

-- =============================================================================
-- 0005_sync_logs.sql - Sync operation logging
-- =============================================================================

CREATE TABLE IF NOT EXISTS bm_sync_logs (
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

CREATE INDEX IF NOT EXISTS idx_sync_logs_branch_type
  ON bm_sync_logs(branch_code, sync_type);

CREATE INDEX IF NOT EXISTS idx_sync_logs_created_at
  ON bm_sync_logs(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_sync_logs_status
  ON bm_sync_logs(status);

COMMENT ON TABLE bm_sync_logs IS 'Audit log for sync operations';
COMMENT ON COLUMN bm_sync_logs.sync_type IS 'Type: yearly_init or monthly_sync';
COMMENT ON COLUMN bm_sync_logs.status IS 'Status: success, error, or in_progress';
COMMENT ON COLUMN bm_sync_logs.triggered_by IS 'Source: api, scheduler, or manual';

-- =============================================================================
-- Verification
-- =============================================================================

-- Display all tables
\dt bm_*

-- Display table counts
SELECT 'bm_branches' as table_name, COUNT(*) as row_count FROM bm_branches
UNION ALL
SELECT 'bm_custcode_init', COUNT(*) FROM bm_custcode_init
UNION ALL
SELECT 'bm_meter_details', COUNT(*) FROM bm_meter_details
UNION ALL
SELECT 'bm_sync_logs', COUNT(*) FROM bm_sync_logs;
