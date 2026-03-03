package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LevelCritical is the slog level for CRITICAL severity (e.g. panic). Use with GetLogger().LogAttrs.
const LevelCritical = slog.LevelError + 1

var (
	globalLogger *slog.Logger
	globalCfg    configHolder
	cachedWriter io.Writer
)

type configHolder struct {
	mu   sync.RWMutex
	cfg  Config
	init bool
}

func (h *configHolder) get() Config {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.cfg
}

func (h *configHolder) set(cfg Config) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cfg = cfg
	h.init = true
}

// lazyWriter defers resolution of the writer until first Write.
type lazyWriter struct {
	once sync.Once
	real io.Writer
}

func (w *lazyWriter) Write(p []byte) (n int, err error) {
	w.once.Do(func() { w.real = getWriter() })
	return w.real.Write(p)
}

func getWriter() io.Writer {
	return os.Stderr
}

// contextKey is an unexported type for context keys.
type contextKey string

const (
	contextKeyTraceID       contextKey = "trace_id"
	contextKeySpanID        contextKey = "span_id"
	contextKeyCorrelationID contextKey = "correlation_id"
	contextKeyHTTPRequest   contextKey = "http_request"
)

// WithTraceID returns a copy of ctx with the given trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, contextKeyTraceID, traceID)
}

// WithSpanID returns a copy of ctx with the given span ID (16-char hex for Cloud Trace).
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, contextKeySpanID, spanID)
}

// WithCorrelationID returns a copy of ctx with the given correlation ID.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, contextKeyCorrelationID, correlationID)
}

// WithHTTPRequest returns a copy of ctx with the given GCP httpRequest for inclusion in logs.
func WithHTTPRequest(ctx context.Context, req *HTTPRequest) context.Context {
	return context.WithValue(ctx, contextKeyHTTPRequest, req)
}

// GetTraceID returns the trace ID from ctx if set.
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(contextKeyTraceID).(string); ok {
		return v
	}
	return ""
}

// GetSpanID returns the span ID from ctx if set.
func GetSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(contextKeySpanID).(string); ok {
		return v
	}
	return ""
}

// GetCorrelationID returns the correlation ID from ctx if set.
func GetCorrelationID(ctx context.Context) string {
	if v, ok := ctx.Value(contextKeyCorrelationID).(string); ok {
		return v
	}
	return ""
}

// GetHTTPRequest returns the GCP httpRequest from ctx if set.
func GetHTTPRequest(ctx context.Context) *HTTPRequest {
	if v, ok := ctx.Value(contextKeyHTTPRequest).(*HTTPRequest); ok {
		return v
	}
	return nil
}

// HTTPRequest is the GCP LogEntry httpRequest shape for structured logging.
// See: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
type HTTPRequest struct {
	RequestMethod string `json:"requestMethod,omitempty"`
	RequestURL    string `json:"requestUrl,omitempty"`
	RequestSize   int64  `json:"requestSize,omitempty"`
	Status        int    `json:"status,omitempty"`
	ResponseSize  int64  `json:"responseSize,omitempty"`
	UserAgent     string `json:"userAgent,omitempty"`
	RemoteIP      string `json:"remoteIp,omitempty"`
	Latency       string `json:"latency,omitempty"` // Duration in seconds, e.g. "0.123s"
}

func getTraceSpanCorrelation(ctx context.Context) (traceID, spanID, correlationID string) {
	if v, ok := ctx.Value(contextKeyTraceID).(string); ok {
		traceID = v
	}
	if v, ok := ctx.Value(contextKeySpanID).(string); ok {
		spanID = v
	}
	if v, ok := ctx.Value(contextKeyCorrelationID).(string); ok {
		correlationID = v
	}
	return traceID, spanID, correlationID
}

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

// sourceLocation for logging.googleapis.com/sourceLocation
type sourceLocation struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

// gcpLogEntry is the JSON shape written to stdout for Cloud Logging.
type gcpLogEntry struct {
	Severity       string                 `json:"severity"`
	Message        string                 `json:"message"`
	Time           string                 `json:"time"`
	Trace          string                 `json:"logging.googleapis.com/trace,omitempty"`
	SpanID         string                 `json:"logging.googleapis.com/spanId,omitempty"`
	SourceLocation *sourceLocation        `json:"logging.googleapis.com/sourceLocation,omitempty"`
	HTTPRequest    *HTTPRequest           `json:"httpRequest,omitempty"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
	Fields         map[string]interface{} `json:"-"`
}

// MarshalJSON merges Fields into the same JSON object.
func (e gcpLogEntry) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, 12)
	m["severity"] = e.Severity
	m["message"] = e.Message
	m["time"] = e.Time
	if e.Trace != "" {
		m["logging.googleapis.com/trace"] = e.Trace
	}
	if e.SpanID != "" {
		m["logging.googleapis.com/spanId"] = e.SpanID
	}
	if e.SourceLocation != nil {
		m["logging.googleapis.com/sourceLocation"] = e.SourceLocation
	}
	if e.HTTPRequest != nil {
		m["httpRequest"] = e.HTTPRequest
	}
	if e.CorrelationID != "" {
		m["correlation_id"] = e.CorrelationID
	}
	for k, v := range e.Fields {
		m[k] = v
	}
	return json.Marshal(m)
}

var (
	entryPool = sync.Pool{
		New: func() interface{} {
			return &gcpLogEntry{Fields: make(map[string]interface{}, 8)}
		},
	}
	bufPool = sync.Pool{
		New: func() interface{} { return new(bytes.Buffer) },
	}
)

func getEntryFromPool() *gcpLogEntry {
	e := entryPool.Get().(*gcpLogEntry)
	if e.Fields == nil {
		e.Fields = make(map[string]interface{}, 8)
	}
	return e
}

func putEntryToPool(e *gcpLogEntry) {
	for k := range e.Fields {
		delete(e.Fields, k)
	}
	e.Severity = ""
	e.Message = ""
	e.Time = ""
	e.Trace = ""
	e.SpanID = ""
	e.SourceLocation = nil
	e.HTTPRequest = nil
	e.CorrelationID = ""
	entryPool.Put(e)
}

// gcpHandler implements slog.Handler for Google Cloud structured logging.
type gcpHandler struct {
	writer io.Writer
	level  slog.Level
}

func newGCPHandler(writer io.Writer, level slog.Level) *gcpHandler {
	return &gcpHandler{writer: writer, level: level}
}

func (h *gcpHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *gcpHandler) Handle(ctx context.Context, record slog.Record) error {
	cfg := globalCfg.get()
	traceID, spanID, correlationID := getTraceSpanCorrelation(ctx)

	entry := getEntryFromPool()
	defer putEntryToPool(entry)

	entry.Severity = gcpSeverity(record.Level)
	entry.Message = record.Message
	entry.Time = record.Time.Format(time.RFC3339Nano)
	entry.CorrelationID = correlationID

	projectID := cfg.ProjectID
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if traceID != "" && projectID != "" {
		entry.Trace = "projects/" + projectID + "/traces/" + traceID
	}
	if spanID != "" {
		entry.SpanID = spanID
	}

	if cfg.EnableHTTPLogging {
		if hr := GetHTTPRequest(ctx); hr != nil {
			entry.HTTPRequest = hr
		}
	}

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
		case "trace_id", "correlation_id", "span_id":
		default:
			v := a.Value.Any()
			if cfg.Redact != nil {
				v = cfg.Redact(a.Key, v)
			}
			entry.Fields[a.Key] = v
		}
		return true
	})

	if locFromAttrs.File != "" || locFromAttrs.Line != 0 || locFromAttrs.Function != "" {
		entry.SourceLocation = &locFromAttrs
	} else if cfg.EnableCaller && record.PC != 0 {
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

	buf := bufPool.Get().(*bytes.Buffer)
	defer func() { buf.Reset(); bufPool.Put(buf) }()
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(entry); err != nil {
		return err
	}
	_, err := h.writer.Write(buf.Bytes())
	return err
}

func (h *gcpHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *gcpHandler) WithGroup(name string) slog.Handler       { return h }

// Init initializes the global logger with the given config. Call once at startup.
func Init(cfg Config) {
	globalCfg.set(cfg)
	if cachedWriter == nil {
		cachedWriter = &lazyWriter{}
	}
	globalLogger = slog.New(newGCPHandler(cachedWriter, cfg.LogLevel))
}

func init() {
	cachedWriter = &lazyWriter{}
	globalCfg.set(Config{LogLevel: slog.LevelInfo})
	globalLogger = slog.New(newGCPHandler(cachedWriter, slog.LevelInfo))
}

// SetLogLevel updates the log level. Kept for backward compatibility.
func SetLogLevel(level slog.Level) {
	if cachedWriter == nil {
		cachedWriter = &lazyWriter{}
	}
	cfg := globalCfg.get()
	cfg.LogLevel = level
	globalCfg.set(cfg)
	globalLogger = slog.New(newGCPHandler(cachedWriter, level))
}

// GetLogger returns the global slog.Logger.
func GetLogger() *slog.Logger {
	if globalLogger == nil {
		globalLogger = slog.New(newGCPHandler(getWriter(), slog.LevelInfo))
	}
	return globalLogger
}

// GetWriter returns the io.Writer used for logging (e.g. for GORM).
func GetWriter() io.Writer {
	if cachedWriter == nil {
		cachedWriter = getWriter()
	}
	return cachedWriter
}

// Fields is a map of key-value pairs for structured logging.
type Fields map[string]interface{}

const callerSkip = 3

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

func logf(ctx context.Context, level slog.Level, skip int, format string, args ...interface{}) {
	l := GetLogger()
	if !l.Enabled(ctx, level) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	cfg := globalCfg.get()
	if cfg.EnableCaller {
		file, line, function := getCallerSource(skip)
		l.Log(ctx, level, msg,
			slog.String("file", file),
			slog.Int("line", line),
			slog.String("function", function),
		)
	} else {
		l.Log(ctx, level, msg)
	}
}

func logAttrs(ctx context.Context, level slog.Level, msg string, fields Fields) {
	l := GetLogger()
	if !l.Enabled(ctx, level) {
		return
	}
	attrs := make([]slog.Attr, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	l.LogAttrs(ctx, level, msg, attrs...)
}

// Debugf logs at level Debug (no context).
func Debugf(format string, args ...interface{}) {
	logf(context.Background(), slog.LevelDebug, callerSkip, format, args...)
}

// Infof logs at level Info (no context).
func Infof(format string, args ...interface{}) {
	logf(context.Background(), slog.LevelInfo, callerSkip, format, args...)
}

// Warnf logs at level Warn (no context).
func Warnf(format string, args ...interface{}) {
	logf(context.Background(), slog.LevelWarn, callerSkip, format, args...)
}

// Errorf logs at level Error (no context).
func Errorf(format string, args ...interface{}) {
	logf(context.Background(), slog.LevelError, callerSkip, format, args...)
}

// Fatalf logs at level Error and exits with code 1. Use only for startup failures; in request handlers use Errorf + abort.
func Fatalf(format string, args ...interface{}) {
	logf(context.Background(), slog.LevelError, callerSkip, format, args...)
	os.Exit(1)
}

// DebugfContext logs at level Debug with trace/span/correlation from ctx.
func DebugfContext(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, slog.LevelDebug, callerSkip, format, args...)
}

// InfofContext logs at level Info with trace/span/correlation from ctx.
func InfofContext(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, slog.LevelInfo, callerSkip, format, args...)
}

// WarnfContext logs at level Warn with trace/span/correlation from ctx.
func WarnfContext(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, slog.LevelWarn, callerSkip, format, args...)
}

// ErrorfContext logs at level Error with trace/span/correlation from ctx.
func ErrorfContext(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, slog.LevelError, callerSkip, format, args...)
}

// Debug logs at level Debug with fields (no context).
func Debug(msg string, fields Fields) {
	logAttrs(context.Background(), slog.LevelDebug, msg, fields)
}

// Info logs at level Info with fields (no context).
func Info(msg string, fields Fields) {
	logAttrs(context.Background(), slog.LevelInfo, msg, fields)
}

// Warn logs at level Warn with fields (no context).
func Warn(msg string, fields Fields) {
	logAttrs(context.Background(), slog.LevelWarn, msg, fields)
}

// Error logs at level Error with fields (no context).
func Error(msg string, fields Fields) {
	logAttrs(context.Background(), slog.LevelError, msg, fields)
}

// Ctx is a logger bound to a context (trace, span, correlation, httpRequest).
type Ctx struct {
	ctx context.Context
}

// WithContext returns a context-bound logger. Use in handlers after trace/HTTP middleware.
func WithContext(ctx context.Context) *Ctx {
	return &Ctx{ctx: ctx}
}

func (c *Ctx) logf(level slog.Level, format string, args ...interface{}) {
	logf(c.ctx, level, callerSkip, format, args...)
}

func (c *Ctx) logAttrs(level slog.Level, msg string, fields Fields) {
	logAttrs(c.ctx, level, msg, fields)
}

// Debugf logs at level Debug with context.
func (c *Ctx) Debugf(format string, args ...interface{}) { c.logf(slog.LevelDebug, format, args...) }

// Infof logs at level Info with context.
func (c *Ctx) Infof(format string, args ...interface{}) { c.logf(slog.LevelInfo, format, args...) }

// Warnf logs at level Warn with context.
func (c *Ctx) Warnf(format string, args ...interface{}) { c.logf(slog.LevelWarn, format, args...) }

// Errorf logs at level Error with context.
func (c *Ctx) Errorf(format string, args ...interface{}) { c.logf(slog.LevelError, format, args...) }

// Fatalf logs at Error and exits. Do not use in request context; use Errorf + abort.
func (c *Ctx) Fatalf(format string, args ...interface{}) {
	c.logf(slog.LevelError, format, args...)
	os.Exit(1)
}

// Debug logs at Debug with fields.
func (c *Ctx) Debug(msg string, fields Fields) { c.logAttrs(slog.LevelDebug, msg, fields) }

// Info logs at Info with fields.
func (c *Ctx) Info(msg string, fields Fields) { c.logAttrs(slog.LevelInfo, msg, fields) }

// Warn logs at Warn with fields.
func (c *Ctx) Warn(msg string, fields Fields) { c.logAttrs(slog.LevelWarn, msg, fields) }

// Error logs at Error with fields.
func (c *Ctx) Error(msg string, fields Fields) { c.logAttrs(slog.LevelError, msg, fields) }