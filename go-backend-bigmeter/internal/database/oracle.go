//go:build oracle
// +build oracle

package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/godror/godror"
)

type Oracle struct {
	DB *sql.DB
}

func NewOracle(dsn string) (*Oracle, error) {
	// godror thick driver accepts EZCONNECT (USER/PASS@host:1521/SERVICE) or oracle:// URL.
	db, err := sql.Open("godror", dsn)
	if err != nil {
		return nil, fmt.Errorf("open oracle (godror): %w", err)
	}
	return &Oracle{DB: db}, nil
}

func (o *Oracle) Ping(ctx context.Context) error { return o.DB.PingContext(ctx) }
func (o *Oracle) Close() {
	if o.DB != nil {
		_ = o.DB.Close()
	}
}
