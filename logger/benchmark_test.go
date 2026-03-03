package logger

import (
	"context"
	"log/slog"
	"testing"
)

func BenchmarkLogf_NoCaller(b *testing.B) {
	Init(Config{LogLevel: slog.LevelInfo, EnableCaller: false})
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InfofContext(ctx, "bench message %d", i)
	}
}

func BenchmarkLogf_WithCaller(b *testing.B) {
	Init(Config{LogLevel: slog.LevelInfo, EnableCaller: true})
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InfofContext(ctx, "bench message %d", i)
	}
}

func BenchmarkLogAttrs_NoFields(b *testing.B) {
	Init(Config{LogLevel: slog.LevelInfo, EnableCaller: false})
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithContext(ctx).Info("bench message", nil)
	}
}

func BenchmarkLogAttrs_WithFields(b *testing.B) {
	Init(Config{LogLevel: slog.LevelInfo, EnableCaller: false})
	ctx := context.Background()
	fields := Fields{"a": 1, "b": "two", "c": true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithContext(ctx).Info("bench message", fields)
	}
}
