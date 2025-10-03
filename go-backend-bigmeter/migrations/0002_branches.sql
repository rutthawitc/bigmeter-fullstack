-- Migration: branches master data
\echo 'Creating bm_branches table'

BEGIN;

CREATE TABLE IF NOT EXISTS bm_branches (
    code TEXT PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;

