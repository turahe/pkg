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

type Option func(*Options)

func WithIAM(v bool) Option {
	return func(o *Options) { o.UseIAM = v }
}

func WithPrivateIP(v bool) Option {
	return func(o *Options) { o.UsePrivateIP = v }
}

func WithLogLevel(level logger.LogLevel) Option {
	return func(o *Options) { o.LogLevel = level }
}

func WithMaxOpenConns(n int) Option {
	return func(o *Options) { o.MaxOpenConns = n }
}

func WithMaxIdleConns(n int) Option {
	return func(o *Options) { o.MaxIdleConns = n }
}

func WithConnMaxLifetime(d time.Duration) Option {
	return func(o *Options) { o.ConnMaxLife = d }
}

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(o *Options) { o.ConnMaxIdle = d }
}

// WithProductionPoolDefaults sets higher connection pool limits for high-throughput production (e.g. 1000+ RPS).
// Call as an override: database.New(cfg, opts, database.WithProductionPoolDefaults).
func WithProductionPoolDefaults(o *Options) {
	o.MaxOpenConns = productionMaxOpenConns
	o.MaxIdleConns = productionMaxIdleConns
}
