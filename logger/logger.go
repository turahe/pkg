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

var (
	logger     *slog.Logger
	cachedWriter io.Writer
)

// getCallerSource returns file, line, and function name for the caller at skip (1 = direct caller).
func getCallerSource(skip int) (file string, line int, function string) {
	pc, path, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0, ""
	}
	file = filepath.Base(path)
	if fn := runtime.FuncForPC(pc); fn != nil {
		function = fn.Name()
	}
	return file, line, function
}

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

// contextKey is an unexported type for context keys to avoid collisions.
type contextKey string

const (
	contextKeyTraceID       contextKey = "trace_id"
	contextKeyCorrelationID contextKey = "correlation_id"
)

// WithTraceID returns a copy of ctx with the given trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, contextKeyTraceID, traceID)
}

// WithCorrelationID returns a copy of ctx with the given correlation ID.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, contextKeyCorrelationID, correlationID)
}

// GetTraceID returns the trace ID from ctx if set, otherwise empty string.
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(contextKeyTraceID).(string); ok {
		return v
	}
	return ""
}

// GetCorrelationID returns the correlation ID from ctx if set, otherwise empty string.
func GetCorrelationID(ctx context.Context) string {
	if v, ok := ctx.Value(contextKeyCorrelationID).(string); ok {
		return v
	}
	return ""
}

// getTraceAndCorrelationID returns both IDs from ctx in one pass (used by handler).
func getTraceAndCorrelationID(ctx context.Context) (traceID, correlationID string) {
	if v, ok := ctx.Value(contextKeyTraceID).(string); ok {
		traceID = v
	}
	if v, ok := ctx.Value(contextKeyCorrelationID).(string); ok {
		correlationID = v
	}
	return traceID, correlationID
}

// gcpLogEntry represents a Google Cloud Logging log entry
type gcpLogEntry struct {
	Severity       string                 `json:"severity"`
	Time           string                 `json:"time"`
	Message        string                 `json:"message"`
	TraceID        string                 `json:"trace_id,omitempty"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
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
	traceID, correlationID := getTraceAndCorrelationID(ctx)
	entry := gcpLogEntry{
		Severity:      gcpSeverity(record.Level),
		Time:          record.Time.Format(time.RFC3339Nano),
		Message:       record.Message,
		TraceID:       traceID,
		CorrelationID: correlationID,
	}

	// Collect attributes; extract file/line/function into sourceLocation (from *f caller)
	// and do not put them in fields to avoid duplication and wrong PC-based source.
	fields := make(map[string]interface{}, 8)
	var locFromAttrs sourceLocation
	record.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "file":
			if s, ok := a.Value.Any().(string); ok {
				locFromAttrs.File = s
			}
		case "line":
			switch v := a.Value.Any().(type) {
			case int:
				locFromAttrs.Line = v
			case int64:
				locFromAttrs.Line = int(v)
			}
		case "function":
			if s, ok := a.Value.Any().(string); ok {
				locFromAttrs.Function = s
			}
		case "trace_id", "correlation_id":
			// Use context for these; do not duplicate in fields
		default:
			fields[a.Key] = a.Value.Any()
		}
		return true
	})

	if locFromAttrs.File != "" || locFromAttrs.Line != 0 || locFromAttrs.Function != "" {
		entry.SourceLocation = &locFromAttrs
	} else if record.PC != 0 {
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
	cachedWriter = getWriter()
	handler := newGCPHandler(cachedWriter, slog.LevelInfo)
	logger = slog.New(handler)
}

// SetLogLevel sets the log level for the logger
func SetLogLevel(level slog.Level) {
	if cachedWriter == nil {
		cachedWriter = getWriter()
	}
	handler := newGCPHandler(cachedWriter, level)
	logger = slog.New(handler)
}

// Fields is a map of key-value pairs for structured logging
type Fields map[string]interface{}

// logf writes a formatted log at the given level with caller source (skip: 1 = caller of logf wrapper).
func logf(ctx context.Context, level slog.Level, skip int, format string, args ...interface{}) {
	if !logger.Enabled(ctx, level) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	file, line, function := getCallerSource(skip)
	if file != "" || line != 0 || function != "" {
		logger.Log(ctx, level, msg,
			slog.String("file", file),
			slog.Int("line", line),
			slog.String("function", function),
		)
	} else {
		logger.Log(ctx, level, msg)
	}
}

// logAttrs writes a log at the given level with structured fields.
func logAttrs(ctx context.Context, level slog.Level, msg string, fields Fields) {
	if !logger.Enabled(ctx, level) {
		return
	}
	attrs := make([]slog.Attr, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	logger.LogAttrs(ctx, level, msg, attrs...)
}

// callerSkip is the depth to get the user's call site from logf (logf -> *f -> user = 3).
const callerSkip = 3

// DebugfContext logs at level Debug with trace_id/correlation_id from ctx.
func DebugfContext(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, slog.LevelDebug, callerSkip, format, args...)
}

// InfofContext logs at level Info with trace_id/correlation_id from ctx.
func InfofContext(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, slog.LevelInfo, callerSkip, format, args...)
}

// WarnfContext logs at level Warn with trace_id/correlation_id from ctx.
func WarnfContext(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, slog.LevelWarn, callerSkip, format, args...)
}

// ErrorfContext logs at level Error with trace_id/correlation_id from ctx.
func ErrorfContext(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, slog.LevelError, callerSkip, format, args...)
}

// Debugf logs at level Debug (no trace/correlation).
func Debugf(format string, args ...interface{}) {
	logf(context.Background(), slog.LevelDebug, callerSkip, format, args...)
}

// Infof logs at level Info (no trace/correlation).
func Infof(format string, args ...interface{}) {
	logf(context.Background(), slog.LevelInfo, callerSkip, format, args...)
}

// Warnf logs at level Warn (no trace/correlation).
func Warnf(format string, args ...interface{}) {
	logf(context.Background(), slog.LevelWarn, callerSkip, format, args...)
}

// Errorf logs at level Error (no trace/correlation).
func Errorf(format string, args ...interface{}) {
	logf(context.Background(), slog.LevelError, callerSkip, format, args...)
}

// Fatalf logs at level Error and exits with code 1.
func Fatalf(format string, args ...interface{}) {
	logf(context.Background(), slog.LevelError, callerSkip, format, args...)
	os.Exit(1)
}

// Debug logs at level Debug with fields.
func Debug(msg string, fields Fields) {
	logAttrs(context.Background(), slog.LevelDebug, msg, fields)
}

// Info logs at level Info with fields.
func Info(msg string, fields Fields) {
	logAttrs(context.Background(), slog.LevelInfo, msg, fields)
}

// Warn logs at level Warn with fields.
func Warn(msg string, fields Fields) {
	logAttrs(context.Background(), slog.LevelWarn, msg, fields)
}

// Error logs at level Error with fields.
func Error(msg string, fields Fields) {
	logAttrs(context.Background(), slog.LevelError, msg, fields)
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

// Ctx is a logger bound to a context. Use WithContext to create it; then Infof, Debugf, etc.
// automatically include trace_id and correlation_id from that context in the JSON output.
type Ctx struct {
	ctx context.Context
}

// WithContext returns a context-bound logger. Calls to Infof, Debugf, Warnf, Errorf (and Debug, Info, Warn, Error)
// on the returned value will automatically include trace_id and correlation_id from ctx in the log JSON.
// Use this in request handlers after trace middleware has set the context, e.g.:
//
//	log := logger.WithContext(c.Request.Context())
//	log.Infof("user %s logged in", userID)
func WithContext(ctx context.Context) *Ctx {
	return &Ctx{ctx: ctx}
}

func (c *Ctx) logf(level slog.Level, format string, args ...interface{}) {
	if !logger.Enabled(c.ctx, level) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	// Ctx.logf <- Infof <- user => skip 3 for user call site
	file, line, function := getCallerSource(3)
	if file != "" || line != 0 || function != "" {
		logger.Log(c.ctx, level, msg,
			slog.String("file", file),
			slog.Int("line", line),
			slog.String("function", function),
		)
	} else {
		logger.Log(c.ctx, level, msg)
	}
}

// Debugf logs at level Debug with trace_id/correlation_id from the bound context.
func (c *Ctx) Debugf(format string, args ...interface{}) { c.logf(slog.LevelDebug, format, args...) }

// Infof logs at level Info with trace_id/correlation_id from the bound context.
func (c *Ctx) Infof(format string, args ...interface{}) { c.logf(slog.LevelInfo, format, args...) }

// Warnf logs at level Warn with trace_id/correlation_id from the bound context.
func (c *Ctx) Warnf(format string, args ...interface{}) { c.logf(slog.LevelWarn, format, args...) }

// Errorf logs at level Error with trace_id/correlation_id from the bound context.
func (c *Ctx) Errorf(format string, args ...interface{}) { c.logf(slog.LevelError, format, args...) }

// Fatalf logs at level Error and exits with code 1; includes trace_id/correlation_id from the bound context.
func (c *Ctx) Fatalf(format string, args ...interface{}) {
	c.logf(slog.LevelError, format, args...)
	os.Exit(1)
}

func (c *Ctx) logAttrs(level slog.Level, msg string, fields Fields) {
	logAttrs(c.ctx, level, msg, fields)
}

// Debug logs at level Debug with fields; includes trace_id/correlation_id from the bound context.
func (c *Ctx) Debug(msg string, fields Fields) { c.logAttrs(slog.LevelDebug, msg, fields) }

// Info logs at level Info with fields; includes trace_id/correlation_id from the bound context.
func (c *Ctx) Info(msg string, fields Fields) { c.logAttrs(slog.LevelInfo, msg, fields) }

// Warn logs at level Warn with fields; includes trace_id/correlation_id from the bound context.
func (c *Ctx) Warn(msg string, fields Fields) { c.logAttrs(slog.LevelWarn, msg, fields) }

// Error logs at level Error with fields; includes trace_id/correlation_id from the bound context.
func (c *Ctx) Error(msg string, fields Fields) { c.logAttrs(slog.LevelError, msg, fields) }
