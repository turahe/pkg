package database

import (
	"context"
	"errors"
	"fmt"
)

// Health pings the database using a context with timeout d.opts.PingTimeout. Returns an error if db is nil, *sql.DB cannot be obtained, or ping fails.
func (d *Database) Health(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("database not initialized")
	}
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, d.opts.PingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}

// Close runs all registered cleanup functions (e.g. Cloud SQL connector close) and clears the list. Returns the first error if any cleanup fails.
func (d *Database) Close() error {
	var errs []error
	for _, fn := range d.cleanups {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}
	d.cleanups = nil
	if len(errs) > 0 {
		return fmt.Errorf("close: %w", errors.Join(errs...))
	}
	return nil
}
