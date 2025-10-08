package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go-backend-bigmeter/internal/alert"
	"go-backend-bigmeter/internal/config"
	dbpkg "go-backend-bigmeter/internal/database"
	"go-backend-bigmeter/internal/notify"
	syncsvc "go-backend-bigmeter/internal/sync"
)

type Server struct {
	cfg     config.Config
	pg      *dbpkg.Postgres
	ora     *dbpkg.Oracle
	syncSvc *syncsvc.Service
}

func NewServer(cfg config.Config, pg *dbpkg.Postgres, ora *dbpkg.Oracle) *Server {
	var syncService *syncsvc.Service
	if ora != nil {
		syncService = syncsvc.NewService(ora, pg)
	}
	return &Server{
		cfg:     cfg,
		pg:      pg,
		ora:     ora,
		syncSvc: syncService,
	}
}

// Router constructs a Gin engine with routes.
func (s *Server) Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	// Minimal CORS + headers
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Writer.Header().Set("Cache-Control", "no-store")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/healthz", s.gHealth)
		v1.GET("/version", s.gVersion)
		v1.GET("/branches", s.gBranches)
		v1.GET("/custcodes", s.gCustcodes)
		v1.GET("/details", s.gDetails)
		v1.GET("/details/summary", s.gDetailsSummary)
		v1.GET("/custcodes/:cust_code/details", s.gCustcodeDetails)
		// Admin/stub endpoints for frontend integration
		v1.POST("/sync/init", s.pSyncInit)
		v1.POST("/sync/monthly", s.pSyncMonthly)
		v1.GET("/sync/logs", s.gSyncLogs)
		v1.GET("/config", s.gConfig)
		// Telegram test endpoint
		v1.POST("/telegram/test", s.pTelegramTest)
		// Alert test endpoint
		v1.POST("/alerts/test", s.pAlertTest)
	}
	return r
}

func (s *Server) gHealth(c *gin.Context) {
	// Report time in configured local timezone
	loc, err := time.LoadLocation(s.cfg.Timezone)
	if err != nil {
		loc = time.Local
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().In(loc).Format(time.RFC3339),
	})
}

func (s *Server) gVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "bigmeter-sync-api",
		"version": getenv("VERSION", "0.1.0"),
		"commit":  getenv("GIT_SHA", "dev"),
	})
}

func (s *Server) gBranches(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	// Prefer DB source if table exists; otherwise fallback to config list
	type row struct{ Code, Name string }
	var rows []row
	// Attempt DB; ignore error and fallback
	const sqlList = `SELECT code, COALESCE(name,'') FROM bm_branches ORDER BY code`
	if s.pg != nil {
		if r, err := s.pg.Pool.Query(c.Request.Context(), sqlList); err == nil {
			defer r.Close()
			for r.Next() {
				var rr row
				if err := r.Scan(&rr.Code, &rr.Name); err != nil {
					break
				}
				rows = append(rows, rr)
			}
		}
	}
	items := make([]map[string]string, 0)
	if len(rows) > 0 {
		for _, r := range rows {
			if q == "" || strings.Contains(strings.ToLower(r.Code), strings.ToLower(q)) || strings.Contains(strings.ToLower(r.Name), strings.ToLower(q)) {
				m := map[string]string{"code": r.Code}
				if r.Name != "" {
					m["name"] = r.Name
				}
				items = append(items, m)
			}
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items), "limit": 0, "offset": 0})
		return
	}
	// Fallback to env/CSV branches
	for _, b := range s.cfg.Branches {
		if q == "" || strings.Contains(strings.ToLower(b), strings.ToLower(q)) {
			items = append(items, map[string]string{"code": b})
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items), "limit": 0, "offset": 0})
}

func (s *Server) gCustcodes(c *gin.Context) {
	ctx := c.Request.Context()
	branch := strings.TrimSpace(c.Query("branch"))
	if branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch is required"})
		return
	}
	fiscalYear, err := parseFiscalOrYM(c.Query("fiscal_year"), c.Query("ym"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit, offset := parseLimitOffset(c.Query("limit"), c.Query("offset"))
    orderBy := sanitizeOrderBy(c.Query("order_by"), map[string]string{
        "cust_code":  "cust_code",
        "meter_no":   "meter_no",
        "use_type":   "use_type",
        "created_at": "created_at",
        // new sortable fields
        "org_name":   "org_name",
        "use_name":   "use_name",
        "cust_name":  "cust_name",
        "address":    "address",
        "route_code": "route_code",
        "meter_size": "meter_size",
        "meter_brand": "meter_brand",
        "meter_state": "meter_state",
        "debt_ym":     "debt_ym",
    }, "cust_code")
	sortDir := sanitizeSort(c.Query("sort"))
	search := strings.TrimSpace(c.Query("q"))

	base := `SELECT fiscal_year, branch_code, org_name, cust_code, use_type, use_name, cust_name, address, route_code,
                     meter_no, meter_size, meter_brand, meter_state, debt_ym, created_at
             FROM bm_custcode_init WHERE branch_code=$1 AND fiscal_year=$2`
	args := []any{branch, fiscalYear}
    if search != "" {
        // Use the same placeholder $3 for all OR terms (same value)
        base += ` AND (
            cust_code ILIKE $3 OR meter_no ILIKE $3 OR use_type ILIKE $3 OR org_name ILIKE $3 OR
            use_name ILIKE $3 OR cust_name ILIKE $3 OR address ILIKE $3 OR route_code ILIKE $3 OR
            meter_size ILIKE $3 OR meter_brand ILIKE $3 OR meter_state ILIKE $3 OR debt_ym ILIKE $3
        )`
        args = append(args, "%"+search+"%")
    }
	countSQL := "SELECT COUNT(1) FROM (" + base + ") t"
	listSQL := base + fmt.Sprintf(" ORDER BY %s %s LIMIT %d OFFSET %d", orderBy, sortDir, limit, offset)

	var total int
	if err := s.pg.Pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := s.pg.Pool.Query(ctx, listSQL, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

    type item struct {
        FiscalYear int       `json:"fiscal_year"`
        BranchCode string    `json:"branch_code"`
        OrgName    *string   `json:"org_name,omitempty"`
        CustCode   string    `json:"cust_code"`
        UseType    *string   `json:"use_type,omitempty"`
        UseName    *string   `json:"use_name,omitempty"`
        CustName   *string   `json:"cust_name,omitempty"`
        Address    *string   `json:"address,omitempty"`
        RouteCode  *string   `json:"route_code,omitempty"`
        MeterNo    *string   `json:"meter_no,omitempty"`
        MeterSize  *string   `json:"meter_size,omitempty"`
        MeterBrand *string   `json:"meter_brand,omitempty"`
        MeterState *string   `json:"meter_state,omitempty"`
        DebtYM     *string   `json:"debt_ym,omitempty"`
        CreatedAt  time.Time `json:"created_at"`
    }
	var items []item
	for rows.Next() {
		var it item
		var org, ut, uname, cname, addr, route, mn, msize, mbrand, mstate, dym sql.NullString
		if err := rows.Scan(
			&it.FiscalYear, &it.BranchCode, &org, &it.CustCode, &ut, &uname, &cname, &addr, &route,
			&mn, &msize, &mbrand, &mstate, &dym, &it.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		it.OrgName = stringPtr(org)
		it.UseType = stringPtr(ut)
		it.UseName = stringPtr(uname)
		it.CustName = stringPtr(cname)
		it.Address = stringPtr(addr)
		it.RouteCode = stringPtr(route)
		it.MeterNo = stringPtr(mn)
		it.MeterSize = stringPtr(msize)
		it.MeterBrand = stringPtr(mbrand)
		it.MeterState = stringPtr(mstate)
		it.DebtYM = stringPtr(dym)
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (s *Server) gDetails(c *gin.Context) {
	ctx := c.Request.Context()
	ym := strings.TrimSpace(c.Query("ym"))
	branch := strings.TrimSpace(c.Query("branch"))
	if ym == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ym and branch are required"})
		return
	}

	// Get fiscal year from query param if provided, otherwise calculate from ym
	// This allows frontend to specify fiscal year for historical months that belong to different cohorts
	var fiscal int
	if fyParam := strings.TrimSpace(c.Query("fiscal_year")); fyParam != "" {
		if fy, err := strconv.Atoi(fyParam); err == nil && fy > 2000 && fy < 3000 {
			fiscal = fy
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fiscal_year parameter"})
			return
		}
	} else {
		// Default: calculate from year_month (YYYYMM format)
		// Fiscal year: Oct-Dec = year+1, Jan-Sep = year
		fiscal = fiscalYearFromYM(ym)
	}

	limit, offset := parseLimitOffset(c.Query("limit"), c.Query("offset"))
    orderBy := sanitizeOrderBy(c.Query("order_by"), map[string]string{
        "cust_code":           "cust_code",
        "present_water_usg":   "present_water_usg",
        "present_meter_count": "present_meter_count",
        "average":             "average",
        "created_at":          "created_at",
        // optional sort on descriptive fields
        "org_name":   "org_name",
        "use_type":   "use_type",
        "use_name":   "use_name",
        "cust_name":  "cust_name",
        "address":    "address",
        "route_code": "route_code",
        "meter_no":   "meter_no",
        "meter_size": "meter_size",
        "meter_brand": "meter_brand",
        "meter_state": "meter_state",
        "debt_ym":     "debt_ym",
    }, "cust_code")
	sortDir := sanitizeSort(c.Query("sort"))
	search := strings.TrimSpace(c.Query("q"))

	base := `SELECT year_month, branch_code, org_name, cust_code, use_type, use_name, cust_name, address, route_code,
                    meter_no, meter_size, meter_brand, meter_state, average, present_meter_count, present_water_usg,
                    debt_ym, created_at
             FROM bm_meter_details WHERE fiscal_year=$1 AND year_month=$2 AND branch_code=$3`
	args := []any{fiscal, ym, branch}

	custs := multiValues(c.Request.URL.Query(), "cust_code")
	if len(custs) > 0 {
		ph := make([]string, len(custs))
		for i := range custs {
			ph[i] = fmt.Sprintf("$%d", len(args)+i+1)
		}
		base += " AND cust_code IN (" + strings.Join(ph, ",") + ")"
		for _, cc := range custs {
			args = append(args, cc)
		}
	}
    if search != "" {
        args = append(args, "%"+search+"%")
        // one placeholder index for all OR-ed columns
        p := len(args)
        base += fmt.Sprintf(" AND (cust_code ILIKE $%d OR meter_no ILIKE $%d OR cust_name ILIKE $%d OR address ILIKE $%d OR route_code ILIKE $%d OR org_name ILIKE $%d OR use_type ILIKE $%d OR use_name ILIKE $%d)", p, p, p, p, p, p, p, p)
    }
	countSQL := "SELECT COUNT(1) FROM (" + base + ") t"
	listSQL := base + fmt.Sprintf(" ORDER BY %s %s LIMIT %d OFFSET %d", orderBy, sortDir, limit, offset)

	var total int
	if err := s.pg.Pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := s.pg.Pool.Query(ctx, listSQL, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

    type item struct {
        YearMonth         string    `json:"year_month"`
        BranchCode        string    `json:"branch_code"`
        OrgName           *string   `json:"org_name,omitempty"`
        CustCode          string    `json:"cust_code"`
        UseType           *string   `json:"use_type,omitempty"`
        UseName           *string   `json:"use_name,omitempty"`
        CustName          *string   `json:"cust_name,omitempty"`
        Address           *string   `json:"address,omitempty"`
        RouteCode         *string   `json:"route_code,omitempty"`
        MeterNo           *string   `json:"meter_no,omitempty"`
        MeterSize         *string   `json:"meter_size,omitempty"`
        MeterBrand        *string   `json:"meter_brand,omitempty"`
        MeterState        *string   `json:"meter_state,omitempty"`
        Average           float64   `json:"average"`
        PresentMeterCount float64   `json:"present_meter_count"`
        PresentWaterUsg   float64   `json:"present_water_usg"`
        DebtYM            *string   `json:"debt_ym,omitempty"`
        CreatedAt         time.Time `json:"created_at"`
        IsZeroed          bool      `json:"is_zeroed"`
    }
	var items []item
	for rows.Next() {
		var it item
		var org, ut, un, cn, ad, rc, mn, ms, mb, mst, dym *string
		if err := rows.Scan(&it.YearMonth, &it.BranchCode, &org, &it.CustCode, &ut, &un, &cn, &ad, &rc,
			&mn, &ms, &mb, &mst, &it.Average, &it.PresentMeterCount, &it.PresentWaterUsg, &dym, &it.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		it.OrgName = org
		it.UseType, it.UseName, it.CustName, it.Address, it.RouteCode = ut, un, cn, ad, rc
		it.MeterNo, it.MeterSize, it.MeterBrand, it.MeterState, it.DebtYM = mn, ms, mb, mst, dym
		it.IsZeroed = (it.PresentWaterUsg == 0 && it.PresentMeterCount == 0 && (it.OrgName == nil || *it.OrgName == ""))
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (s *Server) gCustcodeDetails(c *gin.Context) {
	custCode := strings.TrimSpace(c.Param("cust_code"))
	if custCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cust_code is required in path"})
		return
	}
	branch := strings.TrimSpace(c.Query("branch"))
	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))
	if branch == "" || from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch, from, to are required"})
		return
	}
	ctx := c.Request.Context()
	sql := `SELECT year_month, present_water_usg, present_meter_count, org_name
            FROM bm_meter_details
            WHERE cust_code=$1 AND branch_code=$2 AND year_month BETWEEN $3 AND $4
            ORDER BY year_month`
	rows, err := s.pg.Pool.Query(ctx, sql, custCode, branch, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	type point struct {
		YM                string  `json:"ym"`
		PresentWaterUsg   float64 `json:"present_water_usg"`
		PresentMeterCount float64 `json:"present_meter_count"`
		IsZeroed          bool    `json:"is_zeroed"`
	}
	var series []point
	for rows.Next() {
		var ym string
		var org *string
		var usg, cnt float64
		if err := rows.Scan(&ym, &usg, &cnt, &org); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		zero := ""
		if org != nil {
			zero = *org
		}
		series = append(series, point{YM: ym, PresentWaterUsg: usg, PresentMeterCount: cnt, IsZeroed: (usg == 0 && cnt == 0 && zero == "")})
	}
	c.JSON(http.StatusOK, gin.H{"cust_code": custCode, "branch_code": branch, "from": from, "to": to, "series": series})
}

func (s *Server) gDetailsSummary(c *gin.Context) {
	ctx := c.Request.Context()
	ym := strings.TrimSpace(c.Query("ym"))
	branch := strings.TrimSpace(c.Query("branch"))
	if ym == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ym and branch are required"})
		return
	}
	var total, zeroed int
	var sum float64
	err := s.pg.Pool.QueryRow(ctx,
		`SELECT COUNT(1) AS total,
                COALESCE(SUM(CASE WHEN present_water_usg=0 AND present_meter_count=0 AND org_name='' THEN 1 ELSE 0 END), 0) AS zeroed,
                COALESCE(SUM(present_water_usg), 0) AS sum_usg
         FROM bm_meter_details WHERE year_month=$1 AND branch_code=$2`, ym, branch,
	).Scan(&total, &zeroed, &sum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ym": ym, "branch": branch, "total": total, "zeroed": zeroed, "active": total - zeroed, "sum_present_water_usg": sum})
}

// pSyncInit triggers yearly initialization sync for specified branches.
func (s *Server) pSyncInit(c *gin.Context) {
	var req struct {
		Branches []string `json:"branches"`
		DebtYM   string   `json:"debt_ym"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	// Check if sync service is available
	if s.syncSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync service not available (Oracle not configured)"})
		return
	}

	// Default branches from config if not provided
	branches := req.Branches
	if len(branches) == 0 {
		branches = s.cfg.Branches
	}
	if len(branches) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branches are required"})
		return
	}

	// Default DEBT_YM to October of current year
	debtYM := strings.TrimSpace(req.DebtYM)
	if debtYM == "" {
		debtYM = fmt.Sprintf("%04d10", time.Now().Year())
	}

	// Normalize to Gregorian YM
	ymGreg, err := normalizeGregorianYM(debtYM)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt_ym; expect YYYYMM"})
		return
	}

	// Convert to Thai YM for Oracle query
	thaiYM, err := toThaiYM(ymGreg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to convert to Thai calendar"})
		return
	}

	fiscal, err := parseFiscalOrYM("", ymGreg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt_ym"})
		return
	}

	started := time.Now()

	// Run sync in background to avoid HTTP timeout issues
	// User can monitor progress via sync logs table
	go func() {
		// Use background context instead of request context
		ctx := context.Background()

		log.Printf("yearly init: starting background sync for %d branches", len(branches))
		totalUpserted := 0
		totalZeroed := 0
		failedCount := 0

		// Execute sync for each branch sequentially (one at a time)
		// This avoids Oracle connection pool exhaustion from concurrent queries
		for _, branch := range branches {
			b := strings.TrimSpace(branch)
			log.Printf("yearly init: processing branch=%s", b)
			upserted, zeroed, err := s.syncSvc.InitCustcodes(ctx, fiscal, b, thaiYM, "api")
			if err != nil {
				log.Printf("yearly init: branch=%s failed: %v", b, err)
				failedCount++
				// Continue with other branches even if one fails
			} else {
				log.Printf("yearly init: branch=%s completed (upserted=%d)", b, upserted)
				totalUpserted += upserted
				totalZeroed += zeroed
			}
		}

		elapsed := time.Since(started)
		log.Printf("yearly init: background sync completed (total branches=%d, failed=%d, upserted=%d, elapsed=%v)",
			len(branches), failedCount, totalUpserted, elapsed)
	}()

	// Return immediately with 202 Accepted
	c.JSON(http.StatusAccepted, gin.H{
		"message":     "Yearly initialization started in background",
		"fiscal_year": fiscal,
		"branches":    branches,
		"debt_ym":     debtYM,
		"started_at":  started.Format(time.RFC3339),
		"note":        "Monitor progress via sync logs table",
	})
}

// pSyncMonthly triggers monthly details sync for specified branches.
func (s *Server) pSyncMonthly(c *gin.Context) {
	var req struct {
		Branches  []string `json:"branches"`
		YM        string   `json:"ym"`
		BatchSize int      `json:"batch_size,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	// Check if sync service is available
	if s.syncSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync service not available (Oracle not configured)"})
		return
	}

	branches := req.Branches
	if len(branches) == 0 {
		branches = s.cfg.Branches
	}
	if len(branches) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branches are required"})
		return
	}

	ym := strings.TrimSpace(req.YM)
	if len(ym) != 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ym is required (YYYYMM)"})
		return
	}
	if _, err := strconv.Atoi(ym[:4]); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ym year"})
		return
	}
	if m, err := strconv.Atoi(ym[4:]); err != nil || m < 1 || m > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ym month"})
		return
	}

	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 100 // default
	}

	started := time.Now()

	// Run sync in background to avoid HTTP timeout issues
	// User can monitor progress via sync logs table
	go func() {
		// Use background context instead of request context
		ctx := context.Background()

		log.Printf("monthly sync: starting background sync for %d branches (ym=%s)", len(branches), ym)
		totalUpserted := 0
		totalZeroed := 0
		failedCount := 0

		// Execute sync for each branch sequentially (one at a time)
		// This avoids Oracle connection pool exhaustion from concurrent queries
		for _, branch := range branches {
			b := strings.TrimSpace(branch)
			log.Printf("monthly sync: processing branch=%s ym=%s", b, ym)
			upserted, zeroed, err := s.syncSvc.MonthlyDetails(ctx, ym, b, batchSize, "api")
			if err != nil {
				log.Printf("monthly sync: branch=%s ym=%s failed: %v", b, ym, err)
				failedCount++
				// Continue with other branches even if one fails
			} else {
				log.Printf("monthly sync: branch=%s ym=%s completed (upserted=%d, zeroed=%d)", b, ym, upserted, zeroed)
				totalUpserted += upserted
				totalZeroed += zeroed
			}
		}

		elapsed := time.Since(started)
		log.Printf("monthly sync: background sync completed (total branches=%d, failed=%d, upserted=%d, zeroed=%d, elapsed=%v)",
			len(branches), failedCount, totalUpserted, totalZeroed, elapsed)
	}()

	// Return immediately with 202 Accepted
	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Monthly sync started in background",
		"ym":         ym,
		"branches":   branches,
		"started_at": started.Format(time.RFC3339),
		"note":       "Monitor progress via sync logs table",
	})
}

// gSyncLogs returns sync operation logs with optional filtering
func (s *Server) gSyncLogs(c *gin.Context) {
	if s.syncSvc == nil || s.syncSvc.LogRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync logs not available"})
		return
	}

	// Parse query parameters
	branchCode := c.Query("branch")
	syncType := c.Query("sync_type")
	status := c.Query("status")

	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	// Build filter
	filter := syncsvc.ListSyncLogsFilter{
		Limit:  limit,
		Offset: offset,
	}
	if branchCode != "" {
		filter.BranchCode = &branchCode
	}
	if syncType != "" {
		filter.SyncType = &syncType
	}
	if status != "" {
		filter.Status = &status
	}

	logs, total, err := s.syncSvc.LogRepo.ListSyncLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// gConfig returns a read-only snapshot of key configuration values.
func (s *Server) gConfig(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "timezone":      s.cfg.Timezone,
        "cron_yearly":   s.cfg.YearlySpec,
        "cron_monthly":  s.cfg.MonthlySpec,
        "branches_count": len(s.cfg.Branches),
    })
}

// pTelegramTest sends a test notification to verify Telegram integration
func (s *Server) pTelegramTest(c *gin.Context) {
	// Check if Telegram is enabled
	if !s.cfg.Telegram.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Telegram notifications are not enabled",
			"enabled": false,
		})
		return
	}

	// Create TelegramNotifier instance
	notifier, err := notify.NewTelegramNotifier(notify.TelegramConfig{
		Enabled:           s.cfg.Telegram.Enabled,
		BotToken:          s.cfg.Telegram.BotToken,
		ChatID:            s.cfg.Telegram.ChatID,
		YearlyPrefix:      s.cfg.Telegram.YearlyPrefix,
		MonthlyPrefix:     s.cfg.Telegram.MonthlyPrefix,
		YearlySuccessMsg:  s.cfg.Telegram.YearlySuccessMsg,
		YearlyFailureMsg:  s.cfg.Telegram.YearlyFailureMsg,
		MonthlySuccessMsg: s.cfg.Telegram.MonthlySuccessMsg,
		MonthlyFailureMsg: s.cfg.Telegram.MonthlyFailureMsg,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to initialize Telegram bot: %v", err),
		})
		return
	}

	// Send test message
	if err := notifier.SendTestMessage(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to send test message: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test notification sent successfully",
		"enabled": true,
		"chat_id": s.cfg.Telegram.ChatID,
	})
}

// pAlertTest triggers an alert calculation and sends notification
func (s *Server) pAlertTest(c *gin.Context) {
	var req struct {
		YM        string  `json:"ym"`
		Threshold float64 `json:"threshold"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body, use defaults
		req.YM = ""
		req.Threshold = 0
	}

	// Default to current month if not specified
	ym := req.YM
	if ym == "" {
		now := time.Now()
		ym = fmt.Sprintf("%04d%02d", now.Year(), now.Month())
	}

	// Validate ym format
	if len(ym) != 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ym format, expect YYYYMM"})
		return
	}

	// Default to config threshold if not specified
	threshold := req.Threshold
	if threshold <= 0 {
		threshold = s.cfg.Alert.Threshold
	}

	// Create alert service
	alertService := alert.NewService(
		s.pg,
		s.cfg.Telegram.BotToken,
		s.cfg.Alert.ChatID,
		threshold,
		s.cfg.Alert.Link,
	)

	// Calculate alerts
	stats, err := alertService.CalculateAlerts(c.Request.Context(), ym, threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send notification if enabled
	if s.cfg.Alert.Enabled {
		if err := alertService.SendNotification(stats); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to send notification: %v", err),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":               "Alert calculated and sent successfully",
		"ym":                    stats.YM,
		"prev_ym":               stats.PrevYM,
		"threshold":             stats.Threshold,
		"total_branches":        stats.TotalBranches,
		"branches_with_alerts":  stats.BranchesWithAlerts,
		"total_customers":       stats.TotalCustomers,
		"notification_enabled":  s.cfg.Alert.Enabled,
	})
}

// helpers
func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func stringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}

func parseFiscalOrYM(fiscal string, ym string) (int, error) {
	if fiscal != "" {
		n, err := strconv.Atoi(fiscal)
		if err != nil {
			return 0, fmt.Errorf("invalid fiscal_year")
		}
		return n, nil
	}
	if ym == "" {
		return 0, fmt.Errorf("either fiscal_year or ym is required")
	}
	if len(ym) != 6 {
		return 0, fmt.Errorf("invalid ym format, expect YYYYMM")
	}
	y, err := strconv.Atoi(ym[:4])
	if err != nil {
		return 0, fmt.Errorf("invalid ym")
	}
	m, err := strconv.Atoi(ym[4:])
	if err != nil || m < 1 || m > 12 {
		return 0, fmt.Errorf("invalid ym")
	}
	if m >= 10 {
		return y + 1, nil
	}
	return y, nil
}

func parseLimitOffset(limStr, offStr string) (int, int) {
	limit := 50
	offset := 0
	if limStr != "" {
		if n, err := strconv.Atoi(limStr); err == nil && n > 0 {
			if n > 500 {
				n = 500
			}
			limit = n
		}
	}
	if offStr != "" {
		if n, err := strconv.Atoi(offStr); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}

func fiscalYearFromYM(ym string) int {
	// ym format: YYYYMM (e.g., "202410" for October 2024)
	// Fiscal year: Oct-Dec (months 10-12) = year+1, Jan-Sep (months 1-9) = year
	if len(ym) != 6 {
		return 0
	}
	year, _ := strconv.Atoi(ym[:4])
	month, _ := strconv.Atoi(ym[4:6])
	if month >= 10 {
		return year + 1
	}
	return year
}

func sanitizeOrderBy(v string, allow map[string]string, def string) string {
	if c, ok := allow[v]; ok {
		return c
	}
	return allow[def]
}
func sanitizeSort(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "DESC" {
		return "DESC"
	}
	return "ASC"
}
func multiValues(q map[string][]string, key string) []string {
	vs := q[key]
	var out []string
	for _, v := range vs {
		parts := strings.Split(v, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

// Helper functions for date conversion (from cmd/sync/main.go)

// normalizeGregorianYM converts a YYYYMM to Gregorian if it's Thai Buddhist calendar
func normalizeGregorianYM(ym string) (string, error) {
	if len(ym) != 6 {
		return "", fmt.Errorf("invalid ym; expect YYYYMM")
	}
	y, err := strconv.Atoi(ym[:4])
	if err != nil {
		return "", fmt.Errorf("invalid ym year")
	}
	m, err := strconv.Atoi(ym[4:])
	if err != nil || m < 1 || m > 12 {
		return "", fmt.Errorf("invalid ym month")
	}
	if y >= 2400 { // Thai -> convert to Gregorian
		y -= 543
	}
	return fmt.Sprintf("%04d%02d", y, m), nil
}

// toThaiYM converts a Gregorian YYYYMM to Thai (Buddhist) YYYYMM by adding 543 to the year
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
