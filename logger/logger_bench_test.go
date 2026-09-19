package logger

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func benchInit(b *testing.B, cfg Config) {
	b.Helper()
	cachedWriter = io.Discard
	Init(cfg)
}

func BenchmarkLogf_NoCaller(b *testing.B) {
	benchInit(b, Config{LogLevel: slog.LevelInfo, EnableCaller: false})
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		InfofContext(ctx, "bench message")
	}
}

func BenchmarkLogf_WithCaller(b *testing.B) {
	benchInit(b, Config{LogLevel: slog.LevelInfo, EnableCaller: true})
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		InfofContext(ctx, "bench message")
	}
}

func BenchmarkLogAttrs_NoFields(b *testing.B) {
	benchInit(b, Config{LogLevel: slog.LevelInfo, EnableCaller: false})
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		WithContext(ctx).Info("bench message", nil)
	}
}

func BenchmarkLogAttrs_WithFields(b *testing.B) {
	benchInit(b, Config{LogLevel: slog.LevelInfo, EnableCaller: false})
	ctx := context.Background()
	fields := Fields{"a": 1, "b": "two", "c": true}
	b.ReportAllocs()
	for b.Loop() {
		WithContext(ctx).Info("bench message", fields)
	}
}
