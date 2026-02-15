package database

import (
	"context"
	"fmt"
)

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

func (d *Database) Close() error {
	var errs []error
	for _, fn := range d.cleanups {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}
	d.cleanups = nil
	if len(errs) > 0 {
		return fmt.Errorf("close: %v", errs)
	}
	return nil
}
