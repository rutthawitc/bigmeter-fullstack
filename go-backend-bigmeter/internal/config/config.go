package config

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds runtime configuration loaded from env vars.
type Config struct {
	Timezone    string
	OracleDSN   string
	PostgresDSN string
	Branches    []string
	// Schedules use cron spec; timezone applied from Timezone.
	YearlySpec        string
	MonthlySpec       string
	EnableYearlyInit  bool
	EnableMonthlySync bool
	// Telegram notification settings
	Telegram TelegramConfig
}

// TelegramConfig holds Telegram notification settings
type TelegramConfig struct {
	Enabled           bool
	BotToken          string
	ChatID            int64
	YearlyPrefix      string
	MonthlyPrefix     string
	YearlySuccessMsg  string
	YearlyFailureMsg  string
	MonthlySuccessMsg string
	MonthlyFailureMsg string
}

// Load loads configuration from environment variables. It will read a local
// .env file if present, and applies sensible defaults for schedules.
func Load() (Config, error) {
	_ = godotenv.Load()

	tz := getEnv("TIMEZONE", "Asia/Bangkok")
	// validate timezone
	if _, err := time.LoadLocation(tz); err != nil {
		return Config{}, fmt.Errorf("invalid TIMEZONE %q: %w", tz, err)
	}

	cfg := Config{
		Timezone:          tz,
		OracleDSN:         os.Getenv("ORACLE_DSN"),
		PostgresDSN:       os.Getenv("POSTGRES_DSN"),
		YearlySpec:        getEnv("CRON_YEARLY", "0 0 22 15 10 *"), // 22:00 Oct 15 every year
		MonthlySpec:       getEnv("CRON_MONTHLY", "0 0 8 16 * *"),  // 08:00 on the 16th monthly
		EnableYearlyInit:  getBoolEnv("ENABLE_YEARLY_INIT", true),
		EnableMonthlySync: getBoolEnv("ENABLE_MONTHLY_SYNC", true),
		Telegram:          loadTelegramConfig(),
	}

	// Branch list as comma-separated codes, e.g. BA01,BA02,...
	if s := os.Getenv("BRANCHES"); s != "" {
		cfg.Branches = splitAndTrim(s, ",")
	} else {
		cfg.Branches = parseBranchesFromCSV()
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getBoolEnv(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v == "true" || v == "1" || v == "yes"
}

func getInt64Env(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func loadTelegramConfig() TelegramConfig {
	return TelegramConfig{
		Enabled:  getBoolEnv("TELEGRAM_ENABLED", false),
		BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID:   getInt64Env("TELEGRAM_CHAT_ID", 0),
		YearlyPrefix: getEnv("TELEGRAM_YEARLY_PREFIX",
			"🔄 <b>Big Meter - Yearly Sync</b>"),
		MonthlyPrefix: getEnv("TELEGRAM_MONTHLY_PREFIX",
			"📊 <b>Big Meter - Monthly Sync</b>"),
		YearlySuccessMsg: getEnv("TELEGRAM_YEARLY_SUCCESS",
			"✅ Yearly cohort init completed successfully\n"+
				"Fiscal Year: {fiscal_year}\n"+
				"Branches: {count} ({branches})\n"+
				"Duration: {duration}\n"+
				"Time: {timestamp}"),
		YearlyFailureMsg: getEnv("TELEGRAM_YEARLY_FAILURE",
			"❌ Yearly cohort init failed\n"+
				"Fiscal Year: {fiscal_year}\n"+
				"Failed Branches: {failed_branches}\n"+
				"Error: {error}\n"+
				"Time: {timestamp}"),
		MonthlySuccessMsg: getEnv("TELEGRAM_MONTHLY_SUCCESS",
			"✅ Monthly sync completed successfully\n"+
				"Year-Month: {year_month}\n"+
				"Branches: {count} ({branches})\n"+
				"Duration: {duration}\n"+
				"Time: {timestamp}"),
		MonthlyFailureMsg: getEnv("TELEGRAM_MONTHLY_FAILURE",
			"❌ Monthly sync failed\n"+
				"Year-Month: {year_month}\n"+
				"Failed Branches: {failed_branches}\n"+
				"Error: {error}\n"+
				"Time: {timestamp}"),
	}
}

func splitAndTrim(s, sep string) []string {
	var out []string
	cur := ""
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			if cur != "" {
				out = append(out, trimSpace(cur))
			}
			cur = ""
			continue
		}
		cur += string(s[i])
	}
	if cur != "" {
		out = append(out, trimSpace(cur))
	}
	return out
}

func trimSpace(s string) string {
	// simple trim without importing strings to keep deps minimal in this file
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}

// parseBranchesFromCSV loads docs/r6_branches.csv at runtime when BRANCHES env is not set.
func parseBranchesFromCSV() []string {
	p := filepath.Join("docs", "r6_branches.csv")
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil || len(rows) == 0 {
		return nil
	}
	var out []string
	for i, rec := range rows {
		if i == 0 {
			continue
		} // skip header
		if len(rec) == 0 {
			continue
		}
		code := trimSpace(rec[0])
		if code != "" {
			out = append(out, code)
		}
	}
	return out
}
