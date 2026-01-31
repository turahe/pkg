package logger

import (
	"context"
	"log/slog"
	"testing"
)

func TestDebugf(t *testing.T) {
	SetLogLevel(slog.LevelDebug)
	Debugf("debug %s", "msg")
}

func TestInfof(t *testing.T) {
	Infof("info %s", "msg")
}

func TestWarnf(t *testing.T) {
	Warnf("warn %s", "msg")
}

func TestErrorf(t *testing.T) {
	Errorf("error %s", "msg")
}

func TestGetLogger(t *testing.T) {
	l := GetLogger()
	if l == nil {
		t.Error("GetLogger must not return nil")
	}
}

func TestSetLogLevel(t *testing.T) {
	SetLogLevel(slog.LevelWarn)
	SetLogLevel(slog.LevelInfo)
}

func TestWithContext(t *testing.T) {
	SetLogLevel(slog.LevelDebug)
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")
	ctx = WithCorrelationID(ctx, "corr-456")

	log := WithContext(ctx)
	log.Infof("context-bound info %s", "msg")
	log.Debugf("context-bound debug %s", "msg")
	log.Warnf("context-bound warn %s", "msg")
	log.Errorf("context-bound error %s", "msg")
	log.Info("context-bound info with fields", Fields{"key": "value"})
	// Fatalf would exit; skip in test
}
