// Package sentryx wraps github.com/getsentry/sentry-go with configuration from config
// so the rest of the service doesn't need to know about Sentry's surface area.
//
// Settings are loaded via config.GetConfig().Sentry (or config.Setup).
// SENTRY_DSN being empty is the explicit "Sentry is off" signal — Init becomes a
// no-op and Flush returns immediately.
package sentryx

import (
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/logger"
)

const defaultServerName = "app"

// LoadConfig returns Sentry settings from the global config package.
func LoadConfig() config.SentryConfiguration {
	return config.GetConfig().Sentry
}

// Init configures the global Sentry hub. Returning a boolean (enabled) instead of an
// error keeps the caller simple — when DSN is empty we report "off" and proceed
// silently, when DSN is set but invalid Sentry's own Init returns an error which we
// log and treat as "off" (the service must keep running either way).
func Init(cfg config.SentryConfiguration) bool {
	if strings.TrimSpace(cfg.DSN) == "" {
		logger.Infof("sentry: disabled (SENTRY_DSN is empty)")
		return false
	}

	serverName := cfg.ServerName
	if strings.TrimSpace(serverName) == "" {
		serverName = defaultServerName
	}

	sampleRate := cfg.SampleRate
	if sampleRate < 0 || sampleRate > 1 {
		sampleRate = 1.0
	}

	tracesSampleRate := cfg.TracesSampleRate
	if tracesSampleRate < 0 || tracesSampleRate > 1 {
		tracesSampleRate = 0.0
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		ServerName:       serverName,
		Debug:            cfg.Debug,
		AttachStacktrace: cfg.AttachStacktrace,
		SampleRate:       sampleRate,
		TracesSampleRate: tracesSampleRate,
	})
	if err != nil {
		logger.Errorf("sentry: init failed: %v", err)
		return false
	}

	logger.Infof("sentry: enabled environment=%s release=%s server_name=%s traces_sample_rate=%.2f",
		cfg.Environment, cfg.Release, serverName, tracesSampleRate)
	return true
}

// Flush drains any buffered events. Safe to call when Sentry was never initialised
// (returns immediately). Use defer sentryx.Flush(cfg.FlushTimeout) right after Init
// so SIGTERM doesn't drop the panic that caused the shutdown.
func Flush(timeout time.Duration) {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	sentry.Flush(timeout)
}
