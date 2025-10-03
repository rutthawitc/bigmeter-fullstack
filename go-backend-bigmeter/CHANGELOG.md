# Changelog

All notable changes to this project are documented in this file.

Format loosely follows Keep a Changelog. Dates use YYYY-MM-DD.

## 2025-09-20

### Added
- Yearly snapshot now persists richer fields to `bm_custcode_init`:
  `org_name, use_name, cust_name, address, route_code, meter_size, meter_brand`.
- New migration `0004_extend_custcode_init.sql` adding the above columns.
- API `/api/v1/custcodes` now exposes the new fields (nullable, omitted when null).
- API search/sort:
  - `/custcodes` search across `cust_code, meter_no, use_type, org_name, use_name, cust_name, address, route_code, meter_size, meter_brand, meter_state, debt_ym`.
  - `/custcodes` sortable fields: `cust_code, meter_no, use_type, created_at, org_name, use_name, cust_name, address, route_code, meter_size, meter_brand, meter_state, debt_ym`.
  - `/details` search across `cust_code, meter_no, cust_name, address, route_code, org_name, use_type, use_name`.
  - `/details` sortable fields include: `cust_code, present_water_usg, present_meter_count, average, created_at, org_name, use_type, use_name, cust_name, address, route_code, meter_no, meter_size, meter_brand, meter_state, debt_ym`.

### Changed
- `sqls/200-meter-minimal.sql` now includes additional columns and is refactored to
  deduplicate + limit to top-200 before joining heavy dimension tables (faster).
- `sqls/200-meter-details.sql` trimmed to the core fields to reduce Oracle load;
  descriptive columns not fetched will store as NULL in Postgres.
- API JSON omits nullable fields (uses `omitempty`) to avoid noisy `null` values.

### Performance
- Yearly init per-branch significantly faster due to limit-then-join strategy.

### Documentation
- Updated README, API spec, CLI cheat sheet, and Dev notes to reflect schema
  changes, search/sort behavior, and SQL optimizations.

### Compatibility Notes
- Clients that previously relied on presence of keys with explicit `null` values
  should note that nullable fields are now omitted from JSON responses.

