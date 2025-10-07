package sync

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbpkg "go-backend-bigmeter/internal/database"
)

// Service provides minimal sync capabilities: ora-test and init-once.
type Service struct {
	Oracle   *dbpkg.Oracle
	Postgres *dbpkg.Postgres
	LogRepo  *LogRepository
}

func NewService(ora *dbpkg.Oracle, pg *dbpkg.Postgres) *Service {
	return &Service{
		Oracle:   ora,
		Postgres: pg,
		LogRepo:  NewLogRepository(pg.Pool),
	}
}

// OraTest pings Oracle and logs a simple count to validate connectivity.
func (s *Service) OraTest(ctx context.Context, branch string, debtYM string) error {
	if err := s.Oracle.Ping(ctx); err != nil {
		return err
	}
	log.Printf("ora-test: ping ok")
	row := s.Oracle.DB.QueryRowContext(ctx, "SELECT banner FROM v$version WHERE ROWNUM=1")
	var banner string
	_ = row.Scan(&banner)
	if banner != "" {
		log.Printf("ora-test: version: %s", banner)
	}
	// Lightweight existence check (avoid full COUNT(*) which may be slow): fetch 1 row
	q := `SELECT 1 FROM PWACIS.TB_TR_DEBT_TRN trn
          WHERE trn.ORG_OWNER_ID = :ORG_OWNER_ID AND trn.DEBT_YM = :DEBT_YM AND ROWNUM=1`
	if r := s.Oracle.DB.QueryRowContext(ctx, q, sql.Named("ORG_OWNER_ID", branch), sql.Named("DEBT_YM", debtYM)); r != nil {
		var one int
		if err := r.Scan(&one); err != nil {
			return fmt.Errorf("ora-test: query failed: %w", err)
		}
	}
	log.Printf("ora-test: branch=%s debt_ym=%s ok", branch, debtYM)
	return nil
}

// InitCustcodes runs the minimal unique-200 SQL and upserts into bm_custcode_init.
func (s *Service) InitCustcodes(ctx context.Context, fiscalYear int, branch string, debtYM string, triggeredBy string) (int, int, error) {
	started := time.Now()
	status := "success"
	defer func() { observeJob("yearly_init", branch, status, started) }()

	// Record sync start
	var logID int64
	var logErr error
	if s.LogRepo != nil {
		logID, logErr = s.LogRepo.RecordSyncStart(ctx, "yearly_init", branch, triggeredBy, nil, &debtYM, &fiscalYear)
		if logErr != nil {
			log.Printf("warning: failed to record sync start: %v", logErr)
		}
	}

	q, err := os.ReadFile(filepath.Join("sqls", "200-meter-minimal.sql"))
	if err != nil {
		if s.LogRepo != nil && logID > 0 {
			s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
		}
		return 0, 0, fmt.Errorf("read minimal sql: %w", err)
	}
	rows, err := s.Oracle.DB.QueryContext(ctx, string(q), sql.Named("ORG_OWNER_ID", branch), sql.Named("DEBT_YM", debtYM))
	if err != nil {
		if s.LogRepo != nil && logID > 0 {
			s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
		}
		return 0, 0, fmt.Errorf("oracle query minimal: %w", err)
	}
	defer rows.Close()

	tx, err := s.Postgres.Pool.Begin(ctx)
	if err != nil {
		if s.LogRepo != nil && logID > 0 {
			s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
		}
		return 0, 0, fmt.Errorf("pg begin: %w", err)
	}
	defer tx.Rollback(ctx)

	insert := `INSERT INTO bm_custcode_init (
                    fiscal_year, branch_code, org_name, cust_code, use_type, use_name, cust_name, address, route_code,
                    meter_no, meter_size, meter_brand, meter_state, debt_ym)
               VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
               ON CONFLICT (fiscal_year, branch_code, cust_code) DO UPDATE SET
                    org_name=EXCLUDED.org_name,
                    use_type=EXCLUDED.use_type,
                    use_name=EXCLUDED.use_name,
                    cust_name=EXCLUDED.cust_name,
                    address=EXCLUDED.address,
                    route_code=EXCLUDED.route_code,
                    meter_no=EXCLUDED.meter_no,
                    meter_size=EXCLUDED.meter_size,
                    meter_brand=EXCLUDED.meter_brand,
                    meter_state=EXCLUDED.meter_state,
                    debt_ym=EXCLUDED.debt_ym`

	count := 0
	keep := make([]string, 0, 200)
	for rows.Next() {
		var (
			ba, orgName, custCode, useType, useName, custName, custAddress, routeCode sql.NullString
			meterNo, sizeName, brandName, meterState, debtYMCol                       sql.NullString
		)
		if err := rows.Scan(
			&ba, &orgName, &custCode, &useType, &useName, &custName, &custAddress, &routeCode,
			&meterNo, &sizeName, &brandName, &meterState, &debtYMCol,
		); err != nil {
			status = "error"
			if s.LogRepo != nil && logID > 0 {
				s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
			}
			return 0, 0, fmt.Errorf("scan minimal: %w", err)
		}
		if _, err := tx.Exec(ctx, insert,
			fiscalYear, branch, orgName.String, custCode.String, useType.String, useName.String, custName.String, custAddress.String, routeCode.String,
			meterNo.String, sizeName.String, brandName.String, meterState.String, debtYMCol.String,
		); err != nil {
			status = "error"
			if s.LogRepo != nil && logID > 0 {
				s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
			}
			return 0, 0, fmt.Errorf("pg insert minimal: %w", err)
		}
		count++
		keep = append(keep, custCode.String)
	}
	if err := rows.Err(); err != nil {
		status = "error"
		if s.LogRepo != nil && logID > 0 {
			s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
		}
		return 0, 0, err
	}
	// Prune extras not in current top-200 cohort for this branch+fiscal
	if len(keep) > 0 {
		// Build DELETE with NOT IN (...) placeholders
		ph := make([]string, len(keep))
		args := make([]any, 0, 2+len(keep))
		args = append(args, fiscalYear, branch)
		for i, c := range keep {
			ph[i] = fmt.Sprintf("$%d", i+3)
			args = append(args, c)
		}
		del := "DELETE FROM bm_custcode_init WHERE fiscal_year=$1 AND branch_code=$2 AND cust_code NOT IN (" + strings.Join(ph, ",") + ")"
		if ct, err := tx.Exec(ctx, del, args...); err != nil {
			status = "error"
			if s.LogRepo != nil && logID > 0 {
				s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
			}
			return 0, 0, fmt.Errorf("pg prune extras: %w", err)
		} else {
			if n := ct.RowsAffected(); n > 0 {
				log.Printf("init: branch=%s fiscal=%d pruned=%d extras", branch, fiscalYear, n)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		status = "error"
		if s.LogRepo != nil && logID > 0 {
			s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
		}
		return 0, 0, err
	}
	log.Printf("init: branch=%s fiscal=%d debt_ym=%s upserted=%d", branch, fiscalYear, debtYM, count)
	addRows("yearly_init", branch, "upserted", count)

	// Record sync success
	if s.LogRepo != nil && logID > 0 {
		if err := s.LogRepo.UpdateSyncSuccess(ctx, logID, count, 0); err != nil {
			log.Printf("warning: failed to update sync log: %v", err)
		}
	}

	// Auto-backfill last 3 months of usage details for the new cohort
	log.Printf("init: branch=%s auto-backfilling last 3 months of usage details", branch)
	if err := s.backfillRecentMonths(ctx, branch, fiscalYear, debtYM, 3, triggeredBy); err != nil {
		log.Printf("warning: backfill failed for branch=%s: %v", branch, err)
		// Don't fail the whole init if backfill fails
	}

	return count, 0, nil
}

// backfillRecentMonths syncs the last N months of usage details after yearly init.
// This provides historical context for the newly captured cohort.
func (s *Service) backfillRecentMonths(ctx context.Context, branch string, fiscalYear int, debtYM string, numMonths int, triggeredBy string) error {
	// Parse debt_ym to get the reference month (e.g., "202410" -> October 2024)
	if len(debtYM) != 6 {
		return fmt.Errorf("invalid debt_ym format: %s", debtYM)
	}

	year, err := strconv.Atoi(debtYM[:4])
	if err != nil {
		return fmt.Errorf("parse year from debt_ym: %w", err)
	}
	month, err := strconv.Atoi(debtYM[4:6])
	if err != nil {
		return fmt.Errorf("parse month from debt_ym: %w", err)
	}

	// Generate list of months to backfill (going backwards from debt_ym)
	months := make([]string, 0, numMonths)
	for i := 0; i < numMonths; i++ {
		// Go back i months
		m := month - i
		y := year
		for m <= 0 {
			m += 12
			y--
		}
		// Convert back to Thai Buddhist year for the sync
		thaiYear := y + 543
		ym := fmt.Sprintf("%d%02d", thaiYear, m)
		months = append(months, ym)
	}

	log.Printf("backfill: branch=%s months=%v", branch, months)

	// Sync each month using MonthlyDetails
	batchSize := 100 // Default batch size
	for _, ym := range months {
		log.Printf("backfill: branch=%s ym=%s starting", branch, ym)
		upserted, zeroed, err := s.MonthlyDetails(ctx, ym, branch, batchSize, triggeredBy)
		if err != nil {
			log.Printf("backfill: branch=%s ym=%s failed: %v", branch, ym, err)
			// Continue with other months even if one fails
			continue
		}
		log.Printf("backfill: branch=%s ym=%s completed (upserted=%d, zeroed=%d)", branch, ym, upserted, zeroed)
	}

	return nil
}

// MonthlyDetails loads monthly details for a given YYYYMM and branch, filtered to the
// cohort captured in bm_custcode_init for the fiscal year of that month.
// It batches cust_codes to avoid overly large IN clauses, upserts rows into bm_meter_details,
// and inserts "zeroed" rows for cohort custcodes that return no Oracle rows for the given month.
func (s *Service) MonthlyDetails(ctx context.Context, ym string, branch string, batchSize int, triggeredBy string) (int, int, error) {
	started := time.Now()
	status := "success"
	defer func() { observeJob("monthly_details", branch, status, started) }()
	if len(ym) != 6 {
		return 0, 0, fmt.Errorf("invalid ym; expect YYYYMM")
	}
	thaiYM, err := toThaiYM(ym)
	if err != nil {
		return 0, 0, err
	}
	fiscal := fiscalYearFromYM(ym)

	// Record sync start
	var logID int64
	var logErr error
	if s.LogRepo != nil {
		logID, logErr = s.LogRepo.RecordSyncStart(ctx, "monthly_sync", branch, triggeredBy, &ym, nil, &fiscal)
		if logErr != nil {
			log.Printf("warning: failed to record sync start: %v", logErr)
		}
	}

	// Load cohort from Postgres
	// Also keep snapshot text fields for zeroed rows (use_type, meter_no, meter_state)
	const qCohort = `SELECT cust_code, COALESCE(use_type,''), COALESCE(meter_no,''), COALESCE(meter_state,'')
                     FROM bm_custcode_init WHERE fiscal_year=$1 AND branch_code=$2`
	rows, err := s.Postgres.Pool.Query(ctx, qCohort, fiscal, branch)
	if err != nil {
		if s.LogRepo != nil && logID > 0 {
			s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
		}
		return 0, 0, fmt.Errorf("pg select cohort: %w", err)
	}
	defer rows.Close()
	var cohort []string
	snap := make(map[string][3]string)
	for rows.Next() {
		var cc, ut, mn, ms string
		if err := rows.Scan(&cc, &ut, &mn, &ms); err != nil {
			if s.LogRepo != nil && logID > 0 {
				s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
			}
			return 0, 0, fmt.Errorf("scan cohort: %w", err)
		}
		cohort = append(cohort, cc)
		snap[cc] = [3]string{ut, mn, ms}
	}
	if err := rows.Err(); err != nil {
		if s.LogRepo != nil && logID > 0 {
			s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
		}
		return 0, 0, err
	}
	if len(cohort) == 0 {
		log.Printf("month: ym=%s branch=%s fiscal=%d cohort=0 (skip)", ym, branch, fiscal)
		// Record success with 0 counts
		if s.LogRepo != nil && logID > 0 {
			s.LogRepo.UpdateSyncSuccess(ctx, logID, 0, 0)
		}
		return 0, 0, nil
	}

	// Prune any existing details rows for this ym+branch that are not in the cohort.
	// This ensures /details returns at most the cohort size (typically 200) and
	// removes leftovers from earlier oversized runs.
	{
		ph := make([]string, len(cohort))
		args := make([]any, 0, 2+len(cohort))
		args = append(args, ym, branch)
		for i, c := range cohort {
			ph[i] = fmt.Sprintf("$%d", i+3)
			args = append(args, c)
		}
		del := "DELETE FROM bm_meter_details WHERE year_month=$1 AND branch_code=$2 AND cust_code NOT IN (" + strings.Join(ph, ",") + ")"
		if ct, err := s.Postgres.Pool.Exec(ctx, del, args...); err != nil {
			status = "error"
			if s.LogRepo != nil && logID > 0 {
				s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
			}
			return 0, 0, fmt.Errorf("pg prune details extras: %w", err)
		} else if n := ct.RowsAffected(); n > 0 {
			log.Printf("month: ym=%s branch=%s pruned_details=%d", ym, branch, n)
		}
	}

	// Load SQL template and prepare base
	b, err := os.ReadFile(filepath.Join("sqls", "200-meter-details.sql"))
	if err != nil {
		if s.LogRepo != nil && logID > 0 {
			s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
		}
		return 0, 0, fmt.Errorf("read details sql: %w", err)
	}
	baseSQL := string(b)
	// Remove any FETCH FIRST ...
	baseSQL = removeFetchFirst(baseSQL)

	totalUpserts := 0
	totalZeroed := 0
	batchCount := 0

	for i := 0; i < len(cohort); i += max(1, batchSize) {
		end := i + max(1, batchSize)
		if end > len(cohort) {
			end = len(cohort)
		}
		batch := cohort[i:end]

		// Build IN clause placeholders
		ph := make([]string, len(batch))
		args := []any{sql.Named("ORG_OWNER_ID", branch), sql.Named("DEBT_YM", thaiYM)}
		for j, c := range batch {
			name := fmt.Sprintf("C%d", j)
			ph[j] = ":" + name
			args = append(args, sql.Named(name, c))
		}
		sqlText := strings.Replace(baseSQL, "/*__CUSTCODE_FILTER__*/", "AND trn.CUST_CODE IN ("+strings.Join(ph, ",")+")", 1)

		// Query Oracle
		orows, err := s.Oracle.DB.QueryContext(ctx, sqlText, args...)
		if err != nil {
			status = "error"
			if s.LogRepo != nil && logID > 0 {
				s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
			}
			return 0, 0, fmt.Errorf("oracle details batch %d-%d: %w", i, end, err)
		}

		// Track which custcodes returned data
		seen := make(map[string]bool, len(batch))

		// Upsert results
		tx, err := s.Postgres.Pool.Begin(ctx)
		if err != nil {
			orows.Close()
			status = "error"
			if s.LogRepo != nil && logID > 0 {
				s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
			}
			return 0, 0, fmt.Errorf("pg begin: %w", err)
		}

		upsert := `INSERT INTO bm_meter_details (
                        year_month, branch_code, org_name, cust_code, use_type, use_name, cust_name, address, route_code,
                        meter_no, meter_size, meter_brand, meter_state, average, present_meter_count, present_water_usg, debt_ym)
                    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
                    ON CONFLICT (year_month, branch_code, cust_code) DO UPDATE SET
                        org_name=EXCLUDED.org_name,
                        use_type=EXCLUDED.use_type,
                        use_name=EXCLUDED.use_name,
                        cust_name=EXCLUDED.cust_name,
                        address=EXCLUDED.address,
                        route_code=EXCLUDED.route_code,
                        meter_no=EXCLUDED.meter_no,
                        meter_size=EXCLUDED.meter_size,
                        meter_brand=EXCLUDED.meter_brand,
                        meter_state=EXCLUDED.meter_state,
                        average=EXCLUDED.average,
                        present_meter_count=EXCLUDED.present_meter_count,
                        present_water_usg=EXCLUDED.present_water_usg,
                        debt_ym=EXCLUDED.debt_ym`

		for orows.Next() {
			var cust, mtrNo, debt sql.NullString
			var avg, presentCnt, presentUSG sql.NullFloat64
			if err := orows.Scan(&cust, &mtrNo, &avg, &presentCnt, &presentUSG, &debt); err != nil {
				orows.Close()
				tx.Rollback(ctx)
				status = "error"
				if s.LogRepo != nil && logID > 0 {
					s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
				}
				return 0, 0, fmt.Errorf("scan details: %w", err)
			}
			seen[cust.String] = true
			if _, err := tx.Exec(ctx, upsert,
				ym, branch,
				nil,                     /* org_name */
				cust.String,             /* cust_code */
				nil, nil, nil, nil, nil, /* use_type, use_name, cust_name, address, route_code */
				nullableString(mtrNo), /* meter_no */
				nil, nil, nil,         /* meter_size, meter_brand, meter_state */
				zeroIfNull(avg), zeroIfNull(presentCnt), zeroIfNull(presentUSG), nullableString(debt),
			); err != nil {
				orows.Close()
				tx.Rollback(ctx)
				status = "error"
				if s.LogRepo != nil && logID > 0 {
					s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
				}
				return 0, 0, fmt.Errorf("pg upsert details: %w", err)
			}
			totalUpserts++
		}
		if err := orows.Err(); err != nil {
			orows.Close()
			tx.Rollback(ctx)
			status = "error"
			if s.LogRepo != nil && logID > 0 {
				s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
			}
			return 0, 0, err
		}
		orows.Close()

		// Insert zeroed rows for missing
		for _, c := range batch {
			if seen[c] {
				continue
			}
			snapv := snap[c]
			if _, err := tx.Exec(ctx, upsert,
				ym, branch, "", c, snapv[0], "", "", "", "", snapv[1], "", "", snapv[2],
				0.0, 0.0, 0.0, thaiYM,
			); err != nil {
				tx.Rollback(ctx)
				status = "error"
				if s.LogRepo != nil && logID > 0 {
					s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
				}
				return 0, 0, fmt.Errorf("pg upsert zeroed: %w", err)
			}
			totalZeroed++
		}

		if err := tx.Commit(ctx); err != nil {
			status = "error"
			if s.LogRepo != nil && logID > 0 {
				s.LogRepo.UpdateSyncError(ctx, logID, err.Error())
			}
			return 0, 0, err
		}
		batchCount++
		log.Printf("month: ym=%s branch=%s batch=%d-%d upserted=%d zeroed=%d", ym, branch, i, end-1, totalUpserts, totalZeroed)
	}
	log.Printf("month: ym=%s branch=%s completed upserted=%d zeroed=%d", ym, branch, totalUpserts, totalZeroed)
	addRows("monthly_details", branch, "upserted", totalUpserts)
	addRows("monthly_details", branch, "zeroed", totalZeroed)
	incBatches("monthly_details", branch, batchCount)

	// Record sync success
	if s.LogRepo != nil && logID > 0 {
		if err := s.LogRepo.UpdateSyncSuccess(ctx, logID, totalUpserts, totalZeroed); err != nil {
			log.Printf("warning: failed to update sync log: %v", err)
		}
	}

	return totalUpserts, totalZeroed, nil
}

// helpers for monthly
func toThaiYM(ym string) (string, error) {
	if len(ym) != 6 {
		return "", fmt.Errorf("invalid ym")
	}
	y, err := strconv.Atoi(ym[:4])
	if err != nil {
		return "", fmt.Errorf("invalid ym year")
	}
	mm := ym[4:]
	return fmt.Sprintf("%d%s", y+543, mm), nil
}

func fiscalYearFromYM(ym string) int {
	y, _ := strconv.Atoi(ym[:4])
	m, _ := strconv.Atoi(ym[4:])
	if m >= 10 {
		return y + 1
	}
	return y
}

func removeFetchFirst(s string) string {
	// very simple removal to be robust if template adds it; case-insensitive
	upper := strings.ToUpper(s)
	idx := strings.Index(upper, "FETCH FIRST 200 ROWS ONLY")
	if idx < 0 {
		return s
	}
	// remove that phrase only
	return s[:idx] + s[idx+len("FETCH FIRST 200 ROWS ONLY"):]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func nullableString(ns sql.NullString) any {
	if ns.Valid {
		return ns.String
	}
	return nil
}
func zeroIfNull(n sql.NullFloat64) float64 {
	if n.Valid {
		return n.Float64
	}
	return 0
}
