package logger

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
)

// ErrorStructured logs an error with type, message, and optional stack/cause chain.
// Use from context-bound logger: logger.WithContext(ctx).ErrorStructured(err).
func (c *Ctx) ErrorStructured(err error) {
	if err == nil {
		return
	}
	cfg := globalCfg.get()
	fields := Fields{
		"error_type": typeOf(err),
		"error":       err.Error(),
	}
	if cfg.ErrorStacktrace {
		fields["stacktrace"] = string(debug.Stack())
	}
	if unwrap := errors.Unwrap(err); unwrap != nil {
		fields["cause"] = unwrap.Error()
		fields["cause_type"] = typeOf(unwrap)
	}
	c.Error("error", fields)
}

// ErrorStructuredContext logs an error with type, message, and optional stack/cause from context.
func ErrorStructuredContext(ctx context.Context, err error) {
	if err == nil {
		return
	}
	WithContext(ctx).ErrorStructured(err)
}

func typeOf(err error) string {
	if err == nil {
		return ""
	}
	t := fmt.Sprintf("%T", err)
	// Trim *package.Type to package.Type for readability
	if idx := strings.LastIndex(t, "."); idx >= 0 {
		return t[idx+1:]
	}
	return t
}
