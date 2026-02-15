package database

import (
	"testing"
	"time"

	"gorm.io/gorm/logger"
)

func TestOptions_ApplyDefaults(t *testing.T) {
	opts := &Options{}
	opts.applyDefaults()
	if opts.MaxOpenConns != defaultMaxOpenConns {
		t.Errorf("MaxOpenConns = %d, want %d", opts.MaxOpenConns, defaultMaxOpenConns)
	}
	if opts.MaxIdleConns != defaultMaxIdleConns {
		t.Errorf("MaxIdleConns = %d, want %d", opts.MaxIdleConns, defaultMaxIdleConns)
	}
	if opts.ConnMaxLife != defaultConnMaxLifetime {
		t.Errorf("ConnMaxLife = %v, want %v", opts.ConnMaxLife, defaultConnMaxLifetime)
	}
	if opts.ConnMaxIdle != defaultConnMaxIdleTime {
		t.Errorf("ConnMaxIdle = %v, want %v", opts.ConnMaxIdle, defaultConnMaxIdleTime)
	}
	if opts.SlowThreshold != defaultSlowThreshold {
		t.Errorf("SlowThreshold = %v, want %v", opts.SlowThreshold, defaultSlowThreshold)
	}
	if opts.PingTimeout != defaultPingTimeout {
		t.Errorf("PingTimeout = %v, want %v", opts.PingTimeout, defaultPingTimeout)
	}
	if opts.LogLevel != logger.Warn {
		t.Errorf("LogLevel = %v, want Warn", opts.LogLevel)
	}
}

func TestOptions_ApplyDefaults_RespectsExplicit(t *testing.T) {
	opts := &Options{
		MaxOpenConns:    50,
		MaxIdleConns:    20,
		ConnMaxLife:     time.Hour,
		ConnMaxIdle:     15 * time.Minute,
		SlowThreshold:   time.Second,
		PingTimeout:     10 * time.Second,
		LogLevel:        logger.Info,
	}
	opts.applyDefaults()
	if opts.MaxOpenConns != 50 {
		t.Errorf("MaxOpenConns = %d, want 50", opts.MaxOpenConns)
	}
	if opts.MaxIdleConns != 20 {
		t.Errorf("MaxIdleConns = %d, want 20", opts.MaxIdleConns)
	}
	if opts.ConnMaxLife != time.Hour {
		t.Errorf("ConnMaxLife = %v, want 1h", opts.ConnMaxLife)
	}
	if opts.LogLevel != logger.Info {
		t.Errorf("LogLevel = %v, want Info", opts.LogLevel)
	}
}

func TestOption_WithIAM(t *testing.T) {
	opts := &Options{}
	WithIAM(true)(opts)
	if !opts.UseIAM {
		t.Error("UseIAM should be true")
	}
}

func TestOption_WithPrivateIP(t *testing.T) {
	opts := &Options{}
	WithPrivateIP(true)(opts)
	if !opts.UsePrivateIP {
		t.Error("UsePrivateIP should be true")
	}
}

func TestOption_WithLogLevel(t *testing.T) {
	opts := &Options{}
	WithLogLevel(logger.Info)(opts)
	if opts.LogLevel != logger.Info {
		t.Errorf("LogLevel = %v, want Info", opts.LogLevel)
	}
}

func TestOption_WithMaxOpenConns(t *testing.T) {
	opts := &Options{}
	WithMaxOpenConns(100)(opts)
	if opts.MaxOpenConns != 100 {
		t.Errorf("MaxOpenConns = %d, want 100", opts.MaxOpenConns)
	}
}
