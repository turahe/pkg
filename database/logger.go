package database

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"time"

	pkglogger "github.com/turahe/pkg/logger"
	"gorm.io/gorm/logger"
)

var (
	sensitivePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(password|passwd|pwd)\s*=\s*['"]?[^'"\s]+['"]?`),
		regexp.MustCompile(`(?i)(token|secret|api_key|apikey)\s*=\s*['"]?[^'"\s]+['"]?`),
		regexp.MustCompile(`(?i)(credit_card|card_number|cc_number)\s*=\s*['"]?[^'"\s]+['"]?`),
		regexp.MustCompile(`(?i)(ssn|social_security)\s*=\s*['"]?[^'"\s]+['"]?`),
		regexp.MustCompile(`(?i)(pin|otp)\s*=\s*['"]?[^'"\s]+['"]?`),
	}
	redactionPlaceholder = "[REDACTED]"
)

func redactSQL(sql string) string {
	s := sql
	for _, re := range sensitivePatterns {
		s = re.ReplaceAllString(s, "$1="+redactionPlaceholder)
	}
	for strings.Contains(s, redactionPlaceholder+" "+redactionPlaceholder) {
		s = strings.ReplaceAll(s, redactionPlaceholder+" "+redactionPlaceholder, redactionPlaceholder)
	}
	return s
}

type fintechLogger struct {
	level                   logger.LogLevel
	slowThreshold           time.Duration
	ignoreRecordNotFoundErr bool
}

func newFintechLogger(opts *Options) logger.Interface {
	return &fintechLogger{
		level:                   opts.LogLevel,
		slowThreshold:           opts.SlowThreshold,
		ignoreRecordNotFoundErr: true,
	}
}

func NewFintechLogger(cfg logger.Config) logger.Interface {
	slow := cfg.SlowThreshold
	if slow == 0 {
		slow = defaultSlowThreshold
	}
	return &fintechLogger{
		level:                   cfg.LogLevel,
		slowThreshold:           slow,
		ignoreRecordNotFoundErr: cfg.IgnoreRecordNotFoundError,
	}
}

func (l *fintechLogger) LogMode(level logger.LogLevel) logger.Interface {
	nl := *l
	nl.level = level
	return &nl
}

func (l *fintechLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Info {
		pkglogger.InfofContext(ctx, "[DB] "+msg, args...)
	}
}

func (l *fintechLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Warn {
		pkglogger.WarnfContext(ctx, "[DB] "+msg, args...)
	}
}

func (l *fintechLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Error {
		pkglogger.ErrorfContext(ctx, "[DB] "+msg, args...)
	}
}

func (l *fintechLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	redactedSQL := redactSQL(sql)
	attrs := []slog.Attr{
		slog.Int64("duration_ms", elapsed.Milliseconds()),
		slog.Int64("rows_affected", rows),
		slog.String("sql", redactedSQL),
	}
	if err != nil {
		if err == logger.ErrRecordNotFound && l.ignoreRecordNotFoundErr {
			return
		}
		attrs = append(attrs, slog.String("error", err.Error()))
		pkglogger.GetLogger().LogAttrs(ctx, slog.LevelError, "[DB] query failed", attrs...)
		return
	}
	if elapsed > l.slowThreshold {
		attrs = append(attrs, slog.Bool("slow", true), slog.Int64("threshold_ms", l.slowThreshold.Milliseconds()))
		pkglogger.GetLogger().LogAttrs(ctx, slog.LevelWarn, "[DB] slow query detected", attrs...)
		return
	}
	if l.level >= logger.Info {
		pkglogger.GetLogger().LogAttrs(ctx, slog.LevelInfo, "[DB] query executed", attrs...)
	}
}
