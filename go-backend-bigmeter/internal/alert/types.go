package alert

import "time"

// BranchAlert represents alert statistics for a single branch
type BranchAlert struct {
	BranchCode string
	BranchName string
	Count      int
}

// AlertStats represents overall alert statistics
type AlertStats struct {
	YM                  string
	PrevYM              string
	Threshold           float64
	TotalBranches       int
	BranchesWithAlerts  int
	TotalCustomers      int
	BranchAlerts        []BranchAlert
	GeneratedAt         time.Time
}

// CustomerUsage represents a customer's usage data for percentage calculation
type CustomerUsage struct {
	CustCode      string
	BranchCode    string
	CurrentUsage  float64
	PreviousUsage float64
	Percentage    float64
}
