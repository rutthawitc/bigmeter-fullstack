package config

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
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
