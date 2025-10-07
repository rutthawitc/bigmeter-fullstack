-- Reset all data in Big Meter database
-- WARNING: This will delete all customer data, usage details, and sync logs
-- Branch data (bm_branches) is preserved
--
-- Use this script when you need to clean all synced data but keep branch configuration
--
-- Usage:
--   docker compose exec -T postgres psql -U postgres -d bigmeter < migrations/reset_data.sql

BEGIN;

-- Disable triggers temporarily for faster deletion
SET session_replication_role = 'replica';

-- Delete all sync logs
TRUNCATE TABLE bm_sync_logs RESTART IDENTITY CASCADE;

-- Delete all meter details
TRUNCATE TABLE bm_meter_details CASCADE;

-- Delete all customer code initializations
TRUNCATE TABLE bm_custcode_init CASCADE;

-- Re-enable triggers
SET session_replication_role = 'origin';

-- Display current state
SELECT
    'Reset completed' as status,
    NOW() as timestamp;

-- Show table counts after reset
SELECT 'bm_branches' as table_name, COUNT(*) as row_count FROM bm_branches
UNION ALL
SELECT 'bm_custcode_init', COUNT(*) FROM bm_custcode_init
UNION ALL
SELECT 'bm_meter_details', COUNT(*) FROM bm_meter_details
UNION ALL
SELECT 'bm_sync_logs', COUNT(*) FROM bm_sync_logs;

-- Show preserved branches
SELECT
    'Preserved branches:' as info,
    COUNT(*) as count
FROM bm_branches;

COMMIT;

-- Vacuum tables to reclaim space
VACUUM ANALYZE bm_custcode_init;
VACUUM ANALYZE bm_meter_details;
VACUUM ANALYZE bm_sync_logs;
