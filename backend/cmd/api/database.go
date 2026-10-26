package main

import (
	"context"      // Carries deadlines and cancellation signals to operations.
	"database/sql" // Standard database interface and connection-pool type.
	"fmt"          // Formatting tools; Errorf creates an error rather than printing it.
	"time"         // Supplies durations such as time.Second; this code does not sleep.
)

// openDatabase takes DSN configuration and returns a pool pointer plus an error.
func openDatabase(dsn string) (*sql.DB, error) {
	// Use the registered pgx driver to prepare a pool; this does not verify connectivity.
	db, err := sql.Open("pgx", dsn)
	// A non-nil error means this operation failed.
	if err != nil {
		// No usable pool; %w wraps the original error with our own description.
		// The new error is returned to the caller, not stored in fmt or printed here.
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Background supplies a base context with no deadline; WithTimeout adds five seconds.
	// The deadline starts now. ctx holds the control information; cancel is a function.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// Run on function exit to release context/timer resources; this does not close db.
	defer cancel()

	// Actually verify connectivity, creating a connection if needed and respecting ctx.
	if err := db.PingContext(ctx); err != nil {
		// The failed pool is discarded; cancel() separately cleans up the timeout control.
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Success returns the pool and no error; the pool stays open for its caller.
	return db, nil
}
