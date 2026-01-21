package logger

import (
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
