package database

import (
	"time"

	"gorm.io/gorm/logger"
)

const (
	defaultMaxOpenConns    = 30
	defaultMaxIdleConns    = 10
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 10 * time.Minute
	defaultSlowThreshold   = 500 * time.Millisecond
	defaultPingTimeout     = 5 * time.Second

	// Production pool defaults for high RPS (e.g. 1000+). Use with WithProductionPoolDefaults().
	productionMaxOpenConns = 150
	productionMaxIdleConns = 50
)

// Options holds database connection and pool settings. applyDefaults fills zero values with package defaults.
type Options struct {
	UseIAM       bool
	UsePrivateIP bool
	LogLevel     logger.LogLevel
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
	ConnMaxIdle  time.Duration
	SlowThreshold time.Duration
	PingTimeout  time.Duration
}

func (o *Options) applyDefaults() {
	if o.MaxOpenConns <= 0 {
		o.MaxOpenConns = defaultMaxOpenConns
	}
	if o.MaxIdleConns <= 0 {
		o.MaxIdleConns = defaultMaxIdleConns
	}
	if o.ConnMaxLife <= 0 {
		o.ConnMaxLife = defaultConnMaxLifetime
	}
	if o.ConnMaxIdle <= 0 {
		o.ConnMaxIdle = defaultConnMaxIdleTime
	}
	if o.SlowThreshold <= 0 {
		o.SlowThreshold = defaultSlowThreshold
	}
	if o.PingTimeout <= 0 {
		o.PingTimeout = defaultPingTimeout
	}
	if o.LogLevel < logger.Silent || o.LogLevel > logger.Info {
		o.LogLevel = logger.Warn
	}
}

// Option is a functional option applied to Options (e.g. WithMaxOpenConns, WithProductionPoolDefaults).
type Option func(*Options)

// WithIAM sets whether to use IAM authentication for Cloud SQL.
func WithIAM(v bool) Option {
	return func(o *Options) { o.UseIAM = v }
}

// WithPrivateIP sets whether to use Private IP for Cloud SQL.
func WithPrivateIP(v bool) Option {
	return func(o *Options) { o.UsePrivateIP = v }
}

// WithLogLevel sets the GORM log level (e.g. logger.Warn, logger.Info).
func WithLogLevel(level logger.LogLevel) Option {
	return func(o *Options) { o.LogLevel = level }
}

// WithMaxOpenConns sets the maximum number of open connections to the database.
func WithMaxOpenConns(n int) Option {
	return func(o *Options) { o.MaxOpenConns = n }
}

// WithMaxIdleConns sets the maximum number of idle connections in the pool.
func WithMaxIdleConns(n int) Option {
	return func(o *Options) { o.MaxIdleConns = n }
}

// WithConnMaxLifetime sets the maximum lifetime of a connection.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(o *Options) { o.ConnMaxLife = d }
}

// WithConnMaxIdleTime sets the maximum time a connection may be idle.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(o *Options) { o.ConnMaxIdle = d }
}

// WithProductionPoolDefaults sets higher connection pool limits for high-throughput production (e.g. 1000+ RPS).
// Call as an override: database.New(cfg, opts, database.WithProductionPoolDefaults).
func WithProductionPoolDefaults(o *Options) {
	o.MaxOpenConns = productionMaxOpenConns
	o.MaxIdleConns = productionMaxIdleConns
}
