-- Migration: Add fiscal_year column to bm_meter_details
-- This separates usage details by fiscal year, allowing multiple cohorts to coexist

BEGIN;

-- Add fiscal_year column (nullable initially)
ALTER TABLE bm_meter_details
ADD COLUMN IF NOT EXISTS fiscal_year INTEGER;

-- Populate fiscal_year for existing rows based on year_month
-- Fiscal year logic: Oct-Dec (months 10-12) = year+1, Jan-Sep (months 1-9) = year
UPDATE bm_meter_details
SET fiscal_year = CASE
    WHEN SUBSTRING(year_month, 5, 2)::INTEGER >= 10
    THEN SUBSTRING(year_month, 1, 4)::INTEGER + 1
    ELSE SUBSTRING(year_month, 1, 4)::INTEGER
END
WHERE fiscal_year IS NULL;

-- Make fiscal_year NOT NULL after populating
ALTER TABLE bm_meter_details
ALTER COLUMN fiscal_year SET NOT NULL;

-- Drop old unique constraint
ALTER TABLE bm_meter_details
DROP CONSTRAINT IF EXISTS bm_meter_details_year_month_branch_code_cust_code_key;

-- Add new unique constraint including fiscal_year
ALTER TABLE bm_meter_details
ADD CONSTRAINT bm_meter_details_fiscal_year_year_month_branch_code_cust_code_key
UNIQUE (fiscal_year, year_month, branch_code, cust_code);

-- Add index for better query performance
CREATE INDEX IF NOT EXISTS idx_meter_details_fiscal_year
ON bm_meter_details(fiscal_year);

COMMIT;

-- Verify the change
\echo 'Migration 0006 completed: fiscal_year added to bm_meter_details'
SELECT COUNT(*) as total_rows, COUNT(DISTINCT fiscal_year) as fiscal_years
FROM bm_meter_details;
