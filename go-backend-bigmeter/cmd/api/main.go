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

	srv := api.NewServer(cfg, pg)
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
