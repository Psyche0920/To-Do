package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// openDatabase returns a connection pool after checking database connectivity.
func openDatabase(dsn string) (*sql.DB, error) {
	// Use the registered pgx driver to prepare a pool; this does not verify connectivity.
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		// No usable pool; %w wraps the original error with our own description.
		return nil, fmt.Errorf("open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Actually verify connectivity, creating a connection if needed and respecting ctx.
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
