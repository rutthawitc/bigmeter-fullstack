package alert

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	dbpkg "go-backend-bigmeter/internal/database"
	"go-backend-bigmeter/internal/notify"
)

// Service handles alert calculation and notification logic
type Service struct {
	repo      *Repository
	notifier  *notify.TelegramNotifier
	botToken  string
	threshold float64
	chatID    int64
	link      string
}

// NewService creates a new alert service
func NewService(pg *dbpkg.Postgres, botToken string, chatID int64, threshold float64, link string) *Service {
	return &Service{
		repo:      NewRepository(pg),
		botToken:  botToken,
		chatID:    chatID,
		threshold: threshold,
		link:      link,
	}
}

// CalculateAlerts computes alert statistics for a given year-month
func (s *Service) CalculateAlerts(ctx context.Context, ym string, threshold float64) (*AlertStats, error) {
	// Calculate previous month
	prevYM, err := getPreviousMonth(ym)
	if err != nil {
		return nil, fmt.Errorf("invalid year-month format: %w", err)
	}

	// Calculate fiscal year from current month
	fiscalYear := fiscalYearFromYM(ym)

	// Get all branches
	branches, err := s.repo.GetAllBranches(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	stats := &AlertStats{
		YM:             ym,
		PrevYM:         prevYM,
		Threshold:      threshold,
		TotalBranches:  len(branches),
		BranchAlerts:   make([]BranchAlert, 0),
		GeneratedAt:    time.Now(),
	}

	// Process each branch
	for _, branch := range branches {
		count, err := s.calculateBranchAlerts(ctx, branch.Code, ym, prevYM, fiscalYear, threshold)
		if err != nil {
			log.Printf("alert: failed to calculate for branch %s: %v", branch.Code, err)
			continue
		}

		if count > 0 {
			stats.BranchAlerts = append(stats.BranchAlerts, BranchAlert{
				BranchCode: branch.Code,
				BranchName: branch.Name,
				Count:      count,
			})
			stats.BranchesWithAlerts++
			stats.TotalCustomers += count
		}
	}

	return stats, nil
}

// calculateBranchAlerts calculates the number of customers in a branch that meet the threshold
func (s *Service) calculateBranchAlerts(ctx context.Context, branchCode, ym, prevYM string, fiscalYear int, threshold float64) (int, error) {
	// Get current month usage
	currentData, err := s.repo.GetMonthUsage(ctx, branchCode, ym, fiscalYear)
	if err != nil {
		return 0, err
	}

	// Get previous month usage
	previousData, err := s.repo.GetMonthUsage(ctx, branchCode, prevYM, fiscalYear)
	if err != nil {
		return 0, err
	}

	// Create map for quick lookup of previous month data
	prevMap := make(map[string]float64)
	for _, data := range previousData {
		prevMap[data.CustCode] = data.PresentWaterUsage
	}

	// Count customers that meet threshold
	count := 0
	for _, curr := range currentData {
		prev, exists := prevMap[curr.CustCode]
		if !exists || prev == 0 {
			// Skip if no previous data or previous usage is 0
			continue
		}

		// Calculate percentage change
		pct := ((curr.PresentWaterUsage - prev) / prev) * 100

		// Check if decrease meets threshold (e.g., pct <= -20)
		if pct <= -threshold {
			count++
		}
	}

	return count, nil
}

// RunDaily runs the daily alert check and sends notification
func (s *Service) RunDaily(ctx context.Context, now time.Time) error {
	// Calculate current year-month
	ym := fmt.Sprintf("%04d%02d", now.Year(), now.Month())

	log.Printf("alert: running daily check for ym=%s threshold=%.1f", ym, s.threshold)

	// Calculate alerts
	stats, err := s.CalculateAlerts(ctx, ym, s.threshold)
	if err != nil {
		return fmt.Errorf("failed to calculate alerts: %w", err)
	}

	// Send notification
	return s.SendNotification(stats)
}

// SendNotification sends alert notification via Telegram
func (s *Service) SendNotification(stats *AlertStats) error {
	if s.botToken == "" || s.chatID == 0 {
		log.Printf("alert: telegram not configured, skipping notification")
		return nil
	}

	// Initialize notifier if needed
	if s.notifier == nil {
		var err error
		s.notifier, err = notify.NewTelegramNotifier(notify.TelegramConfig{
			Enabled:  true,
			BotToken: s.botToken,
			ChatID:   s.chatID,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize telegram notifier: %w", err)
		}
	}

	// Format and send message
	message := FormatAlertMessage(stats, s.link)
	return s.notifier.SendAlertMessage(message)
}

// Helper functions

// getPreviousMonth calculates the previous month from YYYYMM format
func getPreviousMonth(ym string) (string, error) {
	if len(ym) != 6 {
		return "", fmt.Errorf("invalid ym format: %s", ym)
	}

	year, err := strconv.Atoi(ym[:4])
	if err != nil {
		return "", fmt.Errorf("invalid year in ym: %s", ym)
	}

	month, err := strconv.Atoi(ym[4:])
	if err != nil || month < 1 || month > 12 {
		return "", fmt.Errorf("invalid month in ym: %s", ym)
	}

	month--
	if month == 0 {
		month = 12
		year--
	}

	return fmt.Sprintf("%04d%02d", year, month), nil
}

// fiscalYearFromYM calculates fiscal year from YYYYMM
// Fiscal year: Oct-Dec (months 10-12) = year+1, Jan-Sep (months 1-9) = year
func fiscalYearFromYM(ym string) int {
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
