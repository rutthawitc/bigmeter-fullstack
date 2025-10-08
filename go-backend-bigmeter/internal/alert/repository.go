package alert

import (
	"context"
	"fmt"

	dbpkg "go-backend-bigmeter/internal/database"
)

// Repository handles database operations for alerts
type Repository struct {
	pg *dbpkg.Postgres
}

// NewRepository creates a new alert repository
func NewRepository(pg *dbpkg.Postgres) *Repository {
	return &Repository{pg: pg}
}

// Branch represents a branch from the database
type Branch struct {
	Code string
	Name string
}

// GetAllBranches retrieves all branches from the database
func (r *Repository) GetAllBranches(ctx context.Context) ([]Branch, error) {
	query := `SELECT code, COALESCE(name, '') as name FROM bm_branches ORDER BY code`
	rows, err := r.pg.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query branches: %w", err)
	}
	defer rows.Close()

	var branches []Branch
	for rows.Next() {
		var b Branch
		if err := rows.Scan(&b.Code, &b.Name); err != nil {
			return nil, fmt.Errorf("failed to scan branch: %w", err)
		}
		branches = append(branches, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating branches: %w", err)
	}

	return branches, nil
}

// UsageData represents usage data for a customer in a specific month
type UsageData struct {
	CustCode         string
	PresentWaterUsage float64
}

// GetMonthUsage retrieves usage data for a specific branch and month
func (r *Repository) GetMonthUsage(ctx context.Context, branchCode, ym string, fiscalYear int) ([]UsageData, error) {
	query := `
		SELECT cust_code, COALESCE(present_water_usg, 0) as present_water_usg
		FROM bm_meter_details
		WHERE branch_code = $1 AND year_month = $2 AND fiscal_year = $3
		ORDER BY cust_code
	`

	rows, err := r.pg.Pool.Query(ctx, query, branchCode, ym, fiscalYear)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage for branch=%s ym=%s: %w", branchCode, ym, err)
	}
	defer rows.Close()

	var usageData []UsageData
	for rows.Next() {
		var u UsageData
		if err := rows.Scan(&u.CustCode, &u.PresentWaterUsage); err != nil {
			return nil, fmt.Errorf("failed to scan usage data: %w", err)
		}
		usageData = append(usageData, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage data: %w", err)
	}

	return usageData, nil
}
