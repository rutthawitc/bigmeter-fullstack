package main

import (
	"context"
	"log"
	"os"
	"time"

	"go-backend-bigmeter/internal/api"
	"go-backend-bigmeter/internal/config"
	dbpkg "go-backend-bigmeter/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pg, err := dbpkg.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	// Initialize Oracle connection for sync operations
	// If Oracle DSN is not configured, sync endpoints will return errors
	var ora *dbpkg.Oracle
	if cfg.OracleDSN != "" {
		ora, err = dbpkg.NewOracle(cfg.OracleDSN)
		if err != nil {
			log.Printf("warning: oracle connection failed (sync endpoints disabled): %v", err)
			ora = nil
		} else {
			defer ora.Close()
			log.Printf("oracle connection initialized for sync operations")
		}
	} else {
		log.Printf("warning: ORACLE_DSN not configured (sync endpoints disabled)")
	}

	srv := api.NewServer(cfg, pg, ora)
	engine := srv.Router()

	addr := ":8089"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("api listening on %s (gin)", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}
