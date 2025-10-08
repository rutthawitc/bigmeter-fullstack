# Changelog

All notable changes to this project are documented in this file.

Format loosely follows Keep a Changelog. Dates use YYYY-MM-DD.

## 2025-10-08

### Added - Telegram Alert System
- **Alert notification system** for water usage decrease detection
  - Compares current month vs previous month usage for all branches
  - Default threshold: 20% decrease (configurable via `TELEGRAM_ALERT_THRESHOLD`)
  - Scheduled runs: 16th and 30th of each month at 09:10 AM (`CRON_ALERT`)
  - Sends formatted Thai message summary to dedicated chat (`TELEGRAM_ALERT_CHAT_ID`)
  - Calculation logic: Skip if previous = 0, include if `((current - previous) / previous) × 100 ≤ -threshold`
- **New API endpoints**:
  - `POST /api/v1/telegram/test` - Send test Telegram notification to verify bot integration
  - `POST /api/v1/alerts/test` - Trigger alert calculation with optional `ym` and `threshold` parameters
- **New internal packages**:
  - `internal/alert/types.go` - Data structures for alert statistics
  - `internal/alert/repository.go` - Database queries for branch and usage data
  - `internal/alert/service.go` - Core calculation and notification logic
  - `internal/alert/message.go` - Thai language message formatting with Buddhist calendar
- **Environment variables**:
  - `TELEGRAM_ALERT_ENABLED` - Enable/disable alert notifications (default: false)
  - `TELEGRAM_ALERT_CHAT_ID` - Telegram chat ID for alerts (separate from sync notifications)
  - `TELEGRAM_ALERT_THRESHOLD` - Alert threshold percentage (default: 20.0)
  - `TELEGRAM_ALERT_LINK` - Optional link to include in alert messages
  - `CRON_ALERT` - Cron schedule for alerts (default: `0 10 9 16,30 * *`)
  - `ENABLE_ALERT` - Feature flag for alert scheduler (default: true)
- **Telegram notifier enhancements**:
  - Added `SendAlertMessage()` method to `internal/notify/telegram.go`
  - Support for separate chat IDs for sync vs alert notifications

### Added - Frontend Current Month Banner
- **Current Month Banner** - Informational banner for data availability
  - Displays before 16th of each month with message: "ข้อมูลเดือนปัจจุบัน จะนำเข้าและแสดงผลได้ในวันที่ 16 [month] [year]"
  - Thai Buddhist calendar formatting (e.g., "ตุลาคม 2568")
  - Closeable with session-only dismissal (no localStorage persistence)
  - Yellow warning style with clock emoji (⏰)
  - Component: `big-meter-frontend/src/components/DetailPage/CurrentMonthBanner.tsx`
  - Positioned between FilterSection and data table card

### Changed
- Updated `cmd/sync/main.go` to integrate alert scheduler alongside yearly and monthly sync jobs
- Updated `docker-compose.yml` with new alert environment variables for both `api` and `sync` services

### Documentation
- Updated `CLAUDE.md` with alert notification system and current month banner documentation
- Updated `docs/API-Spec.md` with new endpoints (`/telegram/test`, `/alerts/test`)
- Updated `DEPLOYMENT.md` with new environment variables and configuration examples
- Updated `big-meter-frontend/CLAUDE.md` with current month banner implementation details

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

