package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"go-backend-bigmeter/internal/config"
	dbpkg "go-backend-bigmeter/internal/database"
	"go-backend-bigmeter/internal/notify"
	syncsvc "go-backend-bigmeter/internal/sync"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	pg, err := dbpkg.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	ora, err := dbpkg.NewOracle(cfg.OracleDSN)
	if err != nil {
		log.Fatalf("oracle: %v", err)
	}
	defer ora.Close()

	svc := syncsvc.NewService(ora, pg)

	// Initialize Telegram notifier
	notifier, err := notify.NewTelegramNotifier(notify.TelegramConfig{
		BotToken:          cfg.Telegram.BotToken,
		ChatID:            cfg.Telegram.ChatID,
		Enabled:           cfg.Telegram.Enabled,
		YearlyPrefix:      cfg.Telegram.YearlyPrefix,
		MonthlyPrefix:     cfg.Telegram.MonthlyPrefix,
		YearlySuccessMsg:  cfg.Telegram.YearlySuccessMsg,
		YearlyFailureMsg:  cfg.Telegram.YearlyFailureMsg,
		MonthlySuccessMsg: cfg.Telegram.MonthlySuccessMsg,
		MonthlyFailureMsg: cfg.Telegram.MonthlyFailureMsg,
	})
	if err != nil {
		log.Fatalf("telegram notifier: %v", err)
	}
	if cfg.Telegram.Enabled {
		log.Printf("telegram notifications enabled (chat_id=%d)", cfg.Telegram.ChatID)
	}

	// Optional Prometheus metrics server
	if addr := strings.TrimSpace(os.Getenv("METRICS_ADDR")); addr != "" {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			log.Printf("metrics listening on %s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Printf("metrics server error: %v", err)
			}
		}()
	}

	mode := strings.ToLower(os.Getenv("MODE"))
	switch mode {
	case "ora-test":
		branches := cfg.Branches
		if b := strings.TrimSpace(os.Getenv("BRANCHES")); b != "" {
			branches = strings.Split(b, ",")
		}
		if len(branches) == 0 {
			log.Fatal("ora-test: BRANCHES is required")
		}
		// Accept Gregorian YM; if DEBT_YM is provided (Thai or Gregorian), normalize to Gregorian
		ymIn := strings.TrimSpace(os.Getenv("YM"))
		if ymIn == "" {
			ymIn = strings.TrimSpace(os.Getenv("DEBT_YM"))
		}
		if ymIn == "" {
			log.Fatal("ora-test: YM=YYYYMM (Gregorian) required")
		}
		ymGreg, err := normalizeGregorianYM(ymIn)
		if err != nil {
			log.Fatalf("ora-test YM: %v", err)
		}
		thaiYM, err := toThaiYM(ymGreg)
		if err != nil {
			log.Fatalf("ora-test Thai YM: %v", err)
		}
		if err := svc.OraTest(ctx, strings.TrimSpace(branches[0]), thaiYM); err != nil {
			log.Fatalf("ora-test: %v", err)
		}
	case "init-once":
		fiscal := fiscalYear(time.Now())
		// Accept Gregorian YM via YM env (preferred). If DEBT_YM is provided (Thai or Gregorian), normalize.
		ymIn := strings.TrimSpace(os.Getenv("YM"))
		if ymIn == "" {
			ymIn = strings.TrimSpace(os.Getenv("DEBT_YM"))
		}
		if ymIn == "" {
			ymIn = fmt.Sprintf("%04d10", time.Now().Year())
		}
		ymGreg, err := normalizeGregorianYM(ymIn)
		if err != nil {
			log.Fatalf("init-once YM: %v", err)
		}
		thaiYM, err := toThaiYM(ymGreg)
		if err != nil {
			log.Fatalf("init-once Thai YM: %v", err)
		}
		for _, b := range cfg.Branches {
			if err := svc.InitCustcodes(ctx, fiscal, strings.TrimSpace(b), thaiYM); err != nil {
				log.Printf("init %s: %v", b, err)
			}
		}
		log.Println("init-once completed")
	case "month-once":
		ym := strings.TrimSpace(os.Getenv("YM"))
		if ym == "" {
			log.Fatal("month-once: YM=YYYYMM is required")
		}
		bs := 100
		if v := strings.TrimSpace(os.Getenv("BATCH_SIZE")); v != "" {
			if n, err := fmt.Sscanf(v, "%d", &bs); n == 0 || err != nil {
				bs = 100
			}
		}
		for _, b := range cfg.Branches {
			if err := svc.MonthlyDetails(ctx, ym, strings.TrimSpace(b), bs); err != nil {
				log.Printf("month %s: %v", b, err)
			}
		}
		log.Println("month-once completed")
	default:
		// Scheduler mode (no MODE specified)
		loc, err := time.LoadLocation(cfg.Timezone)
		if err != nil {
			log.Fatalf("timezone: %v", err)
		}
		// Use seconds-field cron (6 fields) to match defaults like "0 0 22 15 10 *"
		cr := cron.New(cron.WithLocation(loc), cron.WithSeconds())

		// Yearly cohort init (optional)
		if cfg.EnableYearlyInit {
			_, err = cr.AddFunc(cfg.YearlySpec, func() {
			now := time.Now().In(loc)
			fiscal := fiscalYear(now)
			// Use Gregorian October of current year for YM; convert to Thai for Oracle
			ymGreg := fmt.Sprintf("%04d10", now.Year())
			thaiYM, _ := toThaiYM(ymGreg)
			log.Printf("cron yearly: start fiscal=%d debt_ym=%s branches=%d", fiscal, thaiYM, len(cfg.Branches))

			startTime := time.Now()
			var failedBranches []string
			var lastError error

			// Concurrency + retry controls
			conc := getEnvInt("SYNC_CONCURRENCY", 2)
			retries := getEnvInt("SYNC_RETRIES", 2)
			delay := getEnvDur("SYNC_RETRY_DELAY", 10*time.Second)
			runBranchesConcurrent(cfg.Branches, conc, func(branch string) {
				err := runWithRetry(retries, delay, func() error {
					return svc.InitCustcodes(context.Background(), fiscal, strings.TrimSpace(branch), thaiYM)
				}, func(attempt int, err error) {
					log.Printf("cron yearly init %s attempt=%d: %v", branch, attempt, err)
				})
				if err != nil {
					failedBranches = append(failedBranches, branch)
					lastError = err
				}
			})

			duration := time.Since(startTime)
			if len(failedBranches) > 0 {
				log.Printf("cron yearly: completed with errors (failed: %d/%d)", len(failedBranches), len(cfg.Branches))
				notifier.NotifyYearlyFailure(fiscal, cfg.Branches, failedBranches, lastError)
			} else {
				log.Printf("cron yearly: completed successfully")
				notifier.NotifyYearlySuccess(fiscal, cfg.Branches, duration)
			}
		})
		if err != nil {
			log.Fatalf("cron yearly add: %v", err)
		}
		} else {
			log.Printf("yearly init disabled (ENABLE_YEARLY_INIT=false)")
		}

		// Monthly details (optional)
		if cfg.EnableMonthlySync {
			_, err = cr.AddFunc(cfg.MonthlySpec, func() {
			now := time.Now().In(loc)
			ym := fmt.Sprintf("%04d%02d", now.Year(), int(now.Month()))
			log.Printf("cron monthly: start ym=%s branches=%d", ym, len(cfg.Branches))

			startTime := time.Now()
			var failedBranches []string
			var lastError error

			// Controls
			conc := getEnvInt("SYNC_CONCURRENCY", 2)
			retries := getEnvInt("SYNC_RETRIES", 2)
			delay := getEnvDur("SYNC_RETRY_DELAY", 10*time.Second)
			bs := getEnvInt("BATCH_SIZE", 100)
			runBranchesConcurrent(cfg.Branches, conc, func(branch string) {
				err := runWithRetry(retries, delay, func() error {
					return svc.MonthlyDetails(context.Background(), ym, strings.TrimSpace(branch), bs)
				}, func(attempt int, err error) {
					log.Printf("cron monthly %s attempt=%d: %v", branch, attempt, err)
				})
				if err != nil {
					failedBranches = append(failedBranches, branch)
					lastError = err
				}
			})

			duration := time.Since(startTime)
			if len(failedBranches) > 0 {
				log.Printf("cron monthly: completed with errors (failed: %d/%d)", len(failedBranches), len(cfg.Branches))
				notifier.NotifyMonthlyFailure(ym, cfg.Branches, failedBranches, lastError)
			} else {
				log.Printf("cron monthly: completed successfully ym=%s", ym)
				notifier.NotifyMonthlySuccess(ym, cfg.Branches, duration)
			}
		})
		if err != nil {
			log.Fatalf("cron monthly add: %v", err)
		}
		} else {
			log.Printf("monthly sync disabled (ENABLE_MONTHLY_SYNC=false)")
		}

		// Log scheduler status
		yearlyStatus := "disabled"
		if cfg.EnableYearlyInit {
			yearlyStatus = cfg.YearlySpec
		}
		monthlyStatus := "disabled"
		if cfg.EnableMonthlySync {
			monthlyStatus = cfg.MonthlySpec
		}
		log.Printf("scheduler running (TZ=%s) yearly='%s' monthly='%s'", cfg.Timezone, yearlyStatus, monthlyStatus)
		cr.Run()
	}
}

// helpers: concurrency & retry
func runWithRetry(retries int, delay time.Duration, fn func() error, onErr func(attempt int, err error)) error {
	if retries < 0 {
		retries = 0
	}
	attempt := 0
	for {
		err := fn()
		if err == nil {
			return nil
		}
		if attempt >= retries {
			return err
		}
		attempt++
		if onErr != nil {
			onErr(attempt, err)
		}
		time.Sleep(delay)
	}
}

func runBranchesConcurrent(branches []string, concurrency int, job func(branch string)) {
	if concurrency < 1 {
		concurrency = 1
	}
	if len(branches) == 0 {
		return
	}
	sem := make(chan struct{}, concurrency)
	done := make(chan struct{})
	go func() {
		for _, b := range branches {
			sem <- struct{}{}
			branch := b
			go func() {
				defer func() { <-sem }()
				job(branch)
			}()
		}
		// wait drain
		for i := 0; i < cap(sem); i++ {
			sem <- struct{}{}
		}
		close(done)
	}()
	<-done
}

func getEnvInt(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvDur(key string, def time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// normalizeGregorianYM accepts either Thai YYYYMM or Gregorian YYYYMM and returns Gregorian YYYYMM.
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

func fiscalYear(t time.Time) int {
	if int(t.Month()) >= 10 {
		return t.Year() + 1
	}
	return t.Year()
}

// toThaiYM converts a Gregorian YYYYMM to Thai (Buddhist) YYYYMM by adding 543 to the year.
// Expects input in the form YYYYMM and returns the same format with the adjusted year.
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
