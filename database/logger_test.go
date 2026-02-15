package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm/logger"
)

func TestRedactSQL(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		{
			name:     "password redaction",
			sql:      `SELECT * FROM users WHERE password='secret123'`,
			expected: `SELECT * FROM users WHERE password=[REDACTED]`,
		},
		{
			name:     "passwd redaction",
			sql:      `UPDATE users SET passwd='x' WHERE id=1`,
			expected: `UPDATE users SET passwd=[REDACTED] WHERE id=1`,
		},
		{
			name:     "token redaction",
			sql:      `SELECT * FROM sessions WHERE token='abc-token-xyz'`,
			expected: `SELECT * FROM sessions WHERE token=[REDACTED]`,
		},
		{
			name:     "api_key redaction",
			sql:      `SELECT * FROM api_keys WHERE api_key='sk-12345'`,
			expected: `SELECT * FROM api_keys WHERE api_key=[REDACTED]`,
		},
		{
			name:     "credit_card redaction",
			sql:      `SELECT * FROM payments WHERE credit_card='4111111111111111'`,
			expected: `SELECT * FROM payments WHERE credit_card=[REDACTED]`,
		},
		{
			name:     "ssn redaction",
			sql:      `UPDATE customers SET ssn='123-45-6789' WHERE id=1`,
			expected: `UPDATE customers SET ssn=[REDACTED] WHERE id=1`,
		},
		{
			name:     "pin redaction",
			sql:      `SELECT * FROM cards WHERE pin='1234'`,
			expected: `SELECT * FROM cards WHERE pin=[REDACTED]`,
		},
		{
			name:     "no sensitive data",
			sql:      `SELECT id, name FROM users WHERE id=1`,
			expected: `SELECT id, name FROM users WHERE id=1`,
		},
		{
			name:     "multiple sensitive fields",
			sql:      `SELECT * FROM users WHERE password='p' AND token='t'`,
			expected: `SELECT * FROM users WHERE password=[REDACTED] AND token=[REDACTED]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactSQL(tt.sql)
			if got != tt.expected {
				t.Errorf("redactSQL() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewFintechLogger(t *testing.T) {
	cfg := logger.Config{
		LogLevel:                  logger.Info,
		SlowThreshold:             100 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
	}
	l := NewFintechLogger(cfg)
	if l == nil {
		t.Fatal("NewFintechLogger() returned nil")
	}
}

func TestNewFintechLogger_DefaultSlowThreshold(t *testing.T) {
	cfg := logger.Config{
		LogLevel:                  logger.Info,
		SlowThreshold:             0, // zero value
		IgnoreRecordNotFoundError: true,
	}
	l := NewFintechLogger(cfg).(*fintechLogger)
	if l == nil {
		t.Fatal("NewFintechLogger() returned nil")
	}
	if l.slowThreshold != defaultSlowThreshold {
		t.Errorf("slowThreshold = %v, want %v", l.slowThreshold, defaultSlowThreshold)
	}
}

func TestFintechLogger_LogMode(t *testing.T) {
	l := NewFintechLogger(logger.Config{LogLevel: logger.Info}).(*fintechLogger)
	if l.level != logger.Info {
		t.Errorf("initial level = %v, want Info", l.level)
	}
	got := l.LogMode(logger.Warn)
	nl, ok := got.(*fintechLogger)
	if !ok {
		t.Fatalf("LogMode() returned %T, want *fintechLogger", got)
	}
	if nl.level != logger.Warn {
		t.Errorf("LogMode(Warn) level = %v, want Warn", nl.level)
	}
	// Original logger unchanged
	if l.level != logger.Info {
		t.Errorf("original logger level changed to %v", l.level)
	}
}

func TestFintechLogger_InfoWarnError_LevelFiltering(t *testing.T) {
	ctx := context.Background()
	// At Silent, Info/Warn/Error should not panic (they no-op)
	l := NewFintechLogger(logger.Config{LogLevel: logger.Silent}).(*fintechLogger)
	l.Info(ctx, "test %s", "info")
	l.Warn(ctx, "test %s", "warn")
	l.Error(ctx, "test %s", "error")
}

func TestFintechLogger_Trace_SilentLevel(t *testing.T) {
	ctx := context.Background()
	l := NewFintechLogger(logger.Config{LogLevel: logger.Silent}).(*fintechLogger)
	begin := time.Now()
	fc := func() (string, int64) { return "SELECT 1", 0 }
	l.Trace(ctx, begin, fc, nil)
	// Should not panic; Trace at Silent is a no-op
}

func TestFintechLogger_Trace_ErrRecordNotFound_Ignored(t *testing.T) {
	ctx := context.Background()
	l := NewFintechLogger(logger.Config{
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: true,
	}).(*fintechLogger)
	begin := time.Now()
	fc := func() (string, int64) { return "SELECT * FROM missing", 0 }
	l.Trace(ctx, begin, fc, logger.ErrRecordNotFound)
	// Should not panic; ErrRecordNotFound is ignored
}

func TestFintechLogger_Trace_ErrRecordNotFound_NotIgnored(t *testing.T) {
	ctx := context.Background()
	l := NewFintechLogger(logger.Config{
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: false,
	}).(*fintechLogger)
	begin := time.Now()
	fc := func() (string, int64) { return "SELECT * FROM missing", 0 }
	l.Trace(ctx, begin, fc, logger.ErrRecordNotFound)
	// Should log error; no panic
}

func TestFintechLogger_Trace_OtherError(t *testing.T) {
	ctx := context.Background()
	l := NewFintechLogger(logger.Config{LogLevel: logger.Info}).(*fintechLogger)
	begin := time.Now()
	fc := func() (string, int64) { return "SELECT 1", 0 }
	l.Trace(ctx, begin, fc, errors.New("connection refused"))
	// Should log error; no panic
}

func TestFintechLogger_Trace_SlowQuery(t *testing.T) {
	ctx := context.Background()
	l := NewFintechLogger(logger.Config{
		LogLevel:       logger.Info,
		SlowThreshold:  10 * time.Millisecond,
	}).(*fintechLogger)
	begin := time.Now().Add(-50 * time.Millisecond) // 50ms ago
	fc := func() (string, int64) { return "SELECT * FROM large_table", 1000 }
	l.Trace(ctx, begin, fc, nil)
	// Should log slow query warning; no panic
}

func TestFintechLogger_Trace_Success(t *testing.T) {
	ctx := context.Background()
	l := NewFintechLogger(logger.Config{
		LogLevel:       logger.Info,
		SlowThreshold:  time.Second,
	}).(*fintechLogger)
	begin := time.Now()
	fc := func() (string, int64) { return "SELECT id FROM users", 5 }
	l.Trace(ctx, begin, fc, nil)
	// Should log info; no panic
}

func TestFintechLogger_Trace_CallbackInvoked(t *testing.T) {
	ctx := context.Background()
	l := NewFintechLogger(logger.Config{LogLevel: logger.Info}).(*fintechLogger)
	begin := time.Now()
	called := false
	fc := func() (string, int64) {
		called = true
		return "SELECT 1", 1
	}
	l.Trace(ctx, begin, fc, nil)
	if !called {
		t.Error("Trace() did not invoke the callback")
	}
}
