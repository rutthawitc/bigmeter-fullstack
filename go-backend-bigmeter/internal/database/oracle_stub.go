//go:build !oracle
// +build !oracle

package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Oracle stub when oracle build tag is not present
type Oracle struct {
	DB *sql.DB
}

func NewOracle(dsn string) (*Oracle, error) {
	return nil, fmt.Errorf("oracle support not compiled (build with -tags oracle)")
}

func (o *Oracle) Ping(ctx context.Context) error {
	return fmt.Errorf("oracle not available")
}

func (o *Oracle) Close() {}
