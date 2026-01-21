package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var logger *slog.Logger

// gcpSeverity maps slog levels to Google Cloud Logging severity levels
func gcpSeverity(level slog.Level) string {
	switch {
	case level < slog.LevelInfo:
		return "DEBUG"
	case level < slog.LevelWarn:
		return "INFO"
	case level < slog.LevelError:
		return "WARNING"
	case level < slog.LevelError+1:
		return "ERROR"
	default:
		return "CRITICAL"
	}
}

// sourceLocation represents the source location in Google Cloud Logging format
type sourceLocation struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

// gcpLogEntry represents a Google Cloud Logging log entry
type gcpLogEntry struct {
	Severity       string                 `json:"severity"`
	Time           string                 `json:"time"`
	Message        string                 `json:"message"`
	SourceLocation *sourceLocation        `json:"sourceLocation,omitempty"`
	Fields         map[string]interface{} `json:"fields,omitempty"`
}

// gcpHandler implements slog.Handler for Google Cloud Logging format
type gcpHandler struct {
	writer io.Writer
	level  slog.Level
}

func newGCPHandler(writer io.Writer, level slog.Level) *gcpHandler {
	return &gcpHandler{
		writer: writer,
		level:  level,
	}
}

func (h *gcpHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *gcpHandler) Handle(ctx context.Context, record slog.Record) error {
	entry := gcpLogEntry{
		Severity: gcpSeverity(record.Level),
		Time:     record.Time.Format(time.RFC3339Nano),
		Message:  record.Message,
	}

	// Add source location if available
	if record.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{record.PC})
		f, _ := fs.Next()
		if f.File != "" {
			entry.SourceLocation = &sourceLocation{
				File:     filepath.Base(f.File),
				Line:     f.Line,
				Function: f.Function,
			}
		}
	}

	// Collect attributes
	fields := make(map[string]interface{})
	record.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	if len(fields) > 0 {
		entry.Fields = fields
	}

	// Encode as JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	jsonData = append(jsonData, '\n')
	_, err = h.writer.Write(jsonData)
	return err
}

func (h *gcpHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, return the same handler
	// In a more complex implementation, you might want to store these attributes
	return h
}

func (h *gcpHandler) WithGroup(name string) slog.Handler {
	// For simplicity, return the same handler
	return h
}

func init() {
	writer := getWriter()
	handler := newGCPHandler(writer, slog.LevelInfo)
	logger = slog.New(handler)
}

// SetLogLevel sets the log level for the logger
func SetLogLevel(level slog.Level) {
	writer := getWriter()
	handler := newGCPHandler(writer, level)
	logger = slog.New(handler)
}

// Fields is a map of key-value pairs for structured logging
type Fields map[string]interface{}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelDebug) {
		// Get caller information for source location
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			fn := runtime.FuncForPC(pc)
			var fnName string
			if fn != nil {
				fnName = fn.Name()
			}
			logger.Log(ctx, slog.LevelDebug, fmt.Sprintf(format, args...),
				slog.String("file", filepath.Base(file)),
				slog.Int("line", line),
				slog.String("function", fnName),
			)
		} else {
			logger.Debug(fmt.Sprintf(format, args...))
		}
	}
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelInfo) {
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			fn := runtime.FuncForPC(pc)
			var fnName string
			if fn != nil {
				fnName = fn.Name()
			}
			logger.Log(ctx, slog.LevelInfo, fmt.Sprintf(format, args...),
				slog.String("file", filepath.Base(file)),
				slog.Int("line", line),
				slog.String("function", fnName),
			)
		} else {
			logger.Info(fmt.Sprintf(format, args...))
		}
	}
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelWarn) {
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			fn := runtime.FuncForPC(pc)
			var fnName string
			if fn != nil {
				fnName = fn.Name()
			}
			logger.Log(ctx, slog.LevelWarn, fmt.Sprintf(format, args...),
				slog.String("file", filepath.Base(file)),
				slog.Int("line", line),
				slog.String("function", fnName),
			)
		} else {
			logger.Warn(fmt.Sprintf(format, args...))
		}
	}
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelError) {
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			fn := runtime.FuncForPC(pc)
			var fnName string
			if fn != nil {
				fnName = fn.Name()
			}
			logger.Log(ctx, slog.LevelError, fmt.Sprintf(format, args...),
				slog.String("file", filepath.Base(file)),
				slog.Int("line", line),
				slog.String("function", fnName),
			)
		} else {
			logger.Error(fmt.Sprintf(format, args...))
		}
	}
}

// Fatalf logs a message at level Fatal on the standard logger and exits.
func Fatalf(format string, args ...interface{}) {
	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelError) {
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			fn := runtime.FuncForPC(pc)
			var fnName string
			if fn != nil {
				fnName = fn.Name()
			}
			logger.Log(ctx, slog.LevelError, fmt.Sprintf(format, args...),
				slog.String("file", filepath.Base(file)),
				slog.Int("line", line),
				slog.String("function", fnName),
			)
		} else {
			logger.Error(fmt.Sprintf(format, args...))
		}
		os.Exit(1)
	}
}

// Debug logs a message at level Debug with fields.
func Debug(msg string, fields Fields) {
	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelDebug) {
		attrs := make([]slog.Attr, 0, len(fields))
		for k, v := range fields {
			attrs = append(attrs, slog.Any(k, v))
		}
		logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
	}
}

// Info logs a message at level Info with fields.
func Info(msg string, fields Fields) {
	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelInfo) {
		attrs := make([]slog.Attr, 0, len(fields))
		for k, v := range fields {
			attrs = append(attrs, slog.Any(k, v))
		}
		logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
	}
}

// Warn logs a message at level Warn with fields.
func Warn(msg string, fields Fields) {
	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelWarn) {
		attrs := make([]slog.Attr, 0, len(fields))
		for k, v := range fields {
			attrs = append(attrs, slog.Any(k, v))
		}
		logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
	}
}

// Error logs a message at level Error with fields.
func Error(msg string, fields Fields) {
	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelError) {
		attrs := make([]slog.Attr, 0, len(fields))
		for k, v := range fields {
			attrs = append(attrs, slog.Any(k, v))
		}
		logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
	}
}

func getWriter() io.Writer {
	if _, err := os.Stat("./log"); os.IsNotExist(err) {
		os.MkdirAll("./log", os.ModePerm)
	}

	file, err := os.OpenFile("log/application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// Use stderr as fallback if we can't open the log file
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		return os.Stderr
	}

	// Write to both file and console (stderr for errors)
	return io.MultiWriter(file, os.Stderr)
}

// GetLogger returns the underlying slog.Logger instance
func GetLogger() *slog.Logger {
	return logger
}
