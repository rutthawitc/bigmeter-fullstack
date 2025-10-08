package notify

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken          string
	ChatID            int64
	Enabled           bool
	YearlyPrefix      string
	MonthlyPrefix     string
	YearlySuccessMsg  string
	YearlyFailureMsg  string
	MonthlySuccessMsg string
	MonthlyFailureMsg string
}

// TelegramNotifier sends notifications to Telegram
type TelegramNotifier struct {
	bot    *tgbotapi.BotAPI
	config TelegramConfig
}

// NewTelegramNotifier creates a new Telegram notifier
func NewTelegramNotifier(config TelegramConfig) (*TelegramNotifier, error) {
	if !config.Enabled {
		return &TelegramNotifier{config: config}, nil
	}

	if config.BotToken == "" {
		return nil, fmt.Errorf("telegram bot token is required when enabled")
	}

	if config.ChatID == 0 {
		return nil, fmt.Errorf("telegram chat ID is required when enabled")
	}

	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &TelegramNotifier{
		bot:    bot,
		config: config,
	}, nil
}

// NotifyYearlySuccess sends a notification for successful yearly sync
func (tn *TelegramNotifier) NotifyYearlySuccess(fiscalYear int, branches []string, duration time.Duration) {
	if !tn.config.Enabled {
		return
	}

	message := tn.buildMessage(
		tn.config.YearlyPrefix,
		tn.config.YearlySuccessMsg,
		map[string]string{
			"{fiscal_year}": fmt.Sprintf("%d", fiscalYear),
			"{branches}":    strings.Join(branches, ", "),
			"{count}":       fmt.Sprintf("%d", len(branches)),
			"{duration}":    formatDuration(duration),
			"{timestamp}":   time.Now().Format("2006-01-02 15:04:05"),
		},
	)

	tn.sendMessage(message)
}

// NotifyYearlyFailure sends a notification for failed yearly sync
func (tn *TelegramNotifier) NotifyYearlyFailure(fiscalYear int, branches []string, failedBranches []string, err error) {
	if !tn.config.Enabled {
		return
	}

	message := tn.buildMessage(
		tn.config.YearlyPrefix,
		tn.config.YearlyFailureMsg,
		map[string]string{
			"{fiscal_year}":     fmt.Sprintf("%d", fiscalYear),
			"{branches}":        strings.Join(branches, ", "),
			"{failed_branches}": strings.Join(failedBranches, ", "),
			"{error}":           err.Error(),
			"{timestamp}":       time.Now().Format("2006-01-02 15:04:05"),
		},
	)

	tn.sendMessage(message)
}

// NotifyMonthlySuccess sends a notification for successful monthly sync
func (tn *TelegramNotifier) NotifyMonthlySuccess(yearMonth string, branches []string, duration time.Duration) {
	if !tn.config.Enabled {
		return
	}

	message := tn.buildMessage(
		tn.config.MonthlyPrefix,
		tn.config.MonthlySuccessMsg,
		map[string]string{
			"{year_month}": yearMonth,
			"{branches}":   strings.Join(branches, ", "),
			"{count}":      fmt.Sprintf("%d", len(branches)),
			"{duration}":   formatDuration(duration),
			"{timestamp}":  time.Now().Format("2006-01-02 15:04:05"),
		},
	)

	tn.sendMessage(message)
}

// NotifyMonthlyFailure sends a notification for failed monthly sync
func (tn *TelegramNotifier) NotifyMonthlyFailure(yearMonth string, branches []string, failedBranches []string, err error) {
	if !tn.config.Enabled {
		return
	}

	message := tn.buildMessage(
		tn.config.MonthlyPrefix,
		tn.config.MonthlyFailureMsg,
		map[string]string{
			"{year_month}":      yearMonth,
			"{branches}":        strings.Join(branches, ", "),
			"{failed_branches}": strings.Join(failedBranches, ", "),
			"{error}":           err.Error(),
			"{timestamp}":       time.Now().Format("2006-01-02 15:04:05"),
		},
	)

	tn.sendMessage(message)
}

// buildMessage constructs the final message by replacing placeholders
func (tn *TelegramNotifier) buildMessage(prefix, template string, replacements map[string]string) string {
	message := template
	for key, value := range replacements {
		message = strings.ReplaceAll(message, key, value)
	}

	if prefix != "" {
		return prefix + "\n" + message
	}
	return message
}

// sendMessage sends a message to Telegram
func (tn *TelegramNotifier) sendMessage(text string) {
	if tn.bot == nil {
		log.Printf("telegram: bot not initialized, skipping notification")
		return
	}

	msg := tgbotapi.NewMessage(tn.config.ChatID, text)
	msg.ParseMode = "HTML"

	_, err := tn.bot.Send(msg)
	if err != nil {
		log.Printf("telegram: failed to send message: %v", err)
	} else {
		log.Printf("telegram: notification sent successfully")
	}
}

// SendTestMessage sends a test notification to verify Telegram integration
func (tn *TelegramNotifier) SendTestMessage() error {
	if !tn.config.Enabled {
		return fmt.Errorf("telegram notifications are disabled")
	}

	if tn.bot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}

	message := fmt.Sprintf("🧪 <b>Big Meter - Test Notification</b>\n\n"+
		"✅ Telegram integration is working correctly!\n"+
		"Time: %s", time.Now().Format("2006-01-02 15:04:05"))

	msg := tgbotapi.NewMessage(tn.config.ChatID, message)
	msg.ParseMode = "HTML"

	_, err := tn.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send test message: %w", err)
	}

	log.Printf("telegram: test notification sent successfully")
	return nil
}

// SendAlertMessage sends an alert notification message
func (tn *TelegramNotifier) SendAlertMessage(message string) error {
	if !tn.config.Enabled {
		return fmt.Errorf("telegram notifications are disabled")
	}

	if tn.bot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}

	msg := tgbotapi.NewMessage(tn.config.ChatID, message)
	msg.ParseMode = "HTML"

	_, err := tn.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send alert message: %w", err)
	}

	log.Printf("telegram: alert notification sent successfully")
	return nil
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}
