-- Migration: extend bm_custcode_init to store richer snapshot fields from Oracle minimal query
\echo 'Altering bm_custcode_init to add more detail columns'

BEGIN;

ALTER TABLE bm_custcode_init
  ADD COLUMN IF NOT EXISTS org_name   TEXT,
  ADD COLUMN IF NOT EXISTS use_name   TEXT,
  ADD COLUMN IF NOT EXISTS cust_name  TEXT,
  ADD COLUMN IF NOT EXISTS address    TEXT,
  ADD COLUMN IF NOT EXISTS route_code TEXT,
  ADD COLUMN IF NOT EXISTS meter_size TEXT,
  ADD COLUMN IF NOT EXISTS meter_brand TEXT;

COMMIT;

