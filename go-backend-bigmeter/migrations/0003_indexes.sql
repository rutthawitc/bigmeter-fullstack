-- Migration: performance indexes for API and sync queries
\echo 'Creating performance indexes'

BEGIN;

-- bm_meter_details frequently filtered by (year_month, branch_code)
CREATE INDEX IF NOT EXISTS idx_bm_details_ym_branch
  ON bm_meter_details(year_month, branch_code);

-- For lookups by cust_code (e.g., series by customer)
CREATE INDEX IF NOT EXISTS idx_bm_details_cust
  ON bm_meter_details(cust_code);

-- Sorting/aggregating on present_water_usg within ym+branch filters
CREATE INDEX IF NOT EXISTS idx_bm_details_ym_branch_usg
  ON bm_meter_details(year_month, branch_code, present_water_usg);

-- bm_custcode_init list queries filter by (branch_code, fiscal_year)
CREATE INDEX IF NOT EXISTS idx_bm_cust_init_branch_fiscal
  ON bm_custcode_init(branch_code, fiscal_year);

COMMIT;

