# Go Package Collection

[![Go Reference](https://pkg.go.dev/badge/github.com/turahe/pkg.svg)](https://pkg.go.dev/github.com/turahe/pkg)
[![Test](https://github.com/turahe/pkg/actions/workflows/test.yml/badge.svg)](https://github.com/turahe/pkg/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/turahe/pkg)](https://goreportcard.com/report/github.com/turahe/pkg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A collection of production-ready Go packages for building web services: database, Redis (standard + cluster), JWT, crypto, GCS, structured logging, HTTP middleware, OpenTelemetry (`otelx`), Sentry (`sentryx`), mutual TLS (`mtls`), Prometheus metrics, graceful shutdown, and utilities. Follows clean architecture boundaries — domain, use-case, and infrastructure are separate.

## Contents

- [Installation](#installation)
- [Packages](#packages)
  - [config](#config)
  - [database](#database)
  - [redis](#redis)
  - [logger](#logger)
  - [middlewares](#middlewares)
  - [otelx](#otelx)
  - [sentryx](#sentryx)
  - [mtls](#mtls)
  - [handler](#handler)
  - [response](#response)
  - [jwt](#jwt)
  - [crypto](#crypto)
  - [gcs](#gcs)
  - [types](#types)
  - [util](#util)
  - [domain](#domain)
  - [usecase](#usecase)
  - [repositories](#repositories)
- [Documentation](#documentation)
- [Usage](#usage)
- [Environment Variables](#environment-variables)
- [Production Wiring Example](#production-wiring-example)
- [Testing](#testing)
- [License](#license)

---

## Installation

```bash
go get github.com/turahe/pkg
```

Requires Go 1.26+.

---

## Packages

### `config`

Loads all configuration from environment variables (or a `.env` file via [godotenv](https://github.com/joho/godotenv)). All packages read from the global `Configuration` singleton.

**API:**
```go
config.Setup(configPath string) error   // load .env; "" uses env vars only
config.GetConfig() *Configuration       // get global config (builds from env if needed)
config.SetConfig(cfg *Configuration)    // override config (testing / manual wiring)
```

**Configuration struct (abbreviated):**
```go
type Configuration struct {
    Server      ServerConfiguration
    Cors        CorsConfiguration
    Database    DatabaseConfiguration
    DatabaseSite DatabaseConfiguration  // optional second DB
    Redis       RedisConfiguration
    GCS         GCSConfiguration
    RateLimiter RateLimiterConfiguration
    Timezone    TimezoneConfiguration
    OpenTelemetry OpenTelemetryConfiguration
    Sentry        SentryConfiguration
    MTLS          MTLSConfiguration
}
```

---

### `database`

Enterprise-grade GORM database layer. Supports MySQL, Postgres, SQLite, SQL Server, and Google Cloud SQL (Postgres/MySQL) with IAM auth and Private IP. Includes SQL redaction in logs (passwords, tokens, card numbers).

**Drivers:** `mysql` · `postgres` · `sqlite` · `sqlserver` · `cloudsql-mysql` · `cloudsql-postgres`

**New API (recommended — dependency injection):**
```go
db, err := database.New(&cfg.Database, database.Options{
    LogLevel: logger.Warn,
})
// or with Cloud SQL options:
db, err := database.New(&cfg.Database, database.Options{
    UseIAM:       true,
    UsePrivateIP: true,
    LogLevel:     logger.Warn,
})

gormDB  := db.DB()                    // *gorm.DB
err     = db.Health(ctx)             // ping with PingTimeout-bounded context
err     = db.Close()                  // close connections + Cloud SQL connector
```

**Options:**

| Option | Default | Description |
|--------|---------|-------------|
| `MaxOpenConns` | 30 | Max open DB connections |
| `MaxIdleConns` | 10 | Max idle DB connections |
| `ConnMaxLife` | 30 min | Max connection lifetime |
| `ConnMaxIdle` | 10 min | Max connection idle time |
| `SlowThreshold` | 500 ms | GORM slow query log threshold |
| `PingTimeout` | 5 s | Health check ping timeout |
| `LogLevel` | Warn | GORM log level |
| `UseIAM` | false | Cloud SQL IAM auth |
| `UsePrivateIP` | false | Cloud SQL Private IP |
| `EnableOpenTelemetry` | auto | GORM OTel plugin; auto when tracing enabled (`OTEL_GORM_ENABLED`) |

**Connection timezone:** set `DATABASE_TIMEZONE` (IANA name) or leave empty to use `SERVER_TIMEZONE` (default `UTC`). Applied to MySQL `loc`, Postgres `timezone`, and Cloud SQL drivers.

**OpenTelemetry GORM:** when `otelx.Init` succeeds (OTLP or GCP exporter), GORM spans are registered automatically unless disabled with `database.WithOpenTelemetry(false)`.

**High-RPS pool preset:**
```go
db, err := database.New(&cfg.Database, database.Options{}, database.WithProductionPoolDefaults)
// sets MaxOpenConns=150, MaxIdleConns=50
```

**Legacy API (global, backward-compatible):**
```go
database.Setup()          // init from config env vars
database.GetDB()          // *gorm.DB (panics if not initialized)
database.GetDBSite()      // secondary *gorm.DB (falls back to primary)
database.HealthCheck(ctx) // check both primary and site DB
database.IsAlive()        // bool
database.Cleanup()        // close all connections
```

---

### `redis`

Redis client wrapper for both standalone Redis and Redis Cluster (Google Cloud Memorystore). Connection pooling, timeouts, and all common commands included.

**Setup:**
```go
redis.Setup() error          // reads config; no-op if REDIS_ENABLED=false
redis.Close() error          // close client; call on shutdown
redis.IsAlive() bool         // ping check
redis.GetUniversalClient() redis.Cmdable  // works with both modes
redis.GetRedis() *redis.Client            // standard mode only
redis.GetRedisCluster() *redis.ClusterClient  // cluster mode only
```

**String operations:**
```go
redis.Get(key string) (string, error)
redis.Set(key string, value interface{}, expiration time.Duration) error
redis.Delete(keys ...string) error
redis.MGet(keys ...string) ([]interface{}, error)
redis.MSet(pairs ...interface{}) error
```

**Hash operations:**
```go
redis.HGet(key, field string) (string, error)
redis.HGetAll(key string) (map[string]string, error)
redis.HSet(key, field string, value interface{}) error
redis.HSetMap(key string, fields map[string]interface{}) error
```

**List operations:**
```go
redis.LPush(key string, values ...interface{}) error
redis.RPop(key string) (string, error)
redis.LRange(key string, start, stop int64) ([]string, error)
```

**Set operations:**
```go
redis.SAdd(key string, members ...interface{}) error
redis.SMembers(key string) ([]string, error)
redis.SRem(key string, members ...interface{}) error
```

**Distributed lock:**
```go
redis.AcquireLock(key string, expiration time.Duration) (bool, error)
redis.ExtendLock(key string, expiration time.Duration) (bool, error)
redis.ReleaseLock(key string) error
```

**Pipeline, Pub/Sub, Scan:**
```go
redis.Pipeline(fn func(redis.Pipeliner) error) error
redis.PipelineSet(pairs map[string]interface{}, expiration time.Duration) error
redis.PublishMessage(channel string, message interface{}) error
redis.SubscribeToChannel(channel string) *redis.PubSub
redis.ScanKeys(pattern string, count int64) ([]string, error)  // cluster-aware
```

---

### `logger`

Structured logging built on `log/slog`. Outputs Google Cloud Logging-compatible JSON with `severity`, `time`, `message`, `trace_id`, `correlation_id`, `sourceLocation`, and optional `fields`. Logs to stderr (no log file). The underlying writer is lazy-initialized on first log write (no I/O at startup).

**Plain functions:**
```go
logger.Debugf(format string, args ...interface{})
logger.Infof(format string, args ...interface{})
logger.Warnf(format string, args ...interface{})
logger.Errorf(format string, args ...interface{})
logger.Fatalf(format string, args ...interface{})  // calls os.Exit(1)
```

**Structured fields:**
```go
logger.Debug(msg string, fields logger.Fields)
logger.Info(msg string, fields logger.Fields)
logger.Warn(msg string, fields logger.Fields)
logger.Error(msg string, fields logger.Fields)
```

**Context-bound (includes `trace_id` / `correlation_id` automatically):**
```go
// Store IDs in context (done by RequestID / TraceMiddleware):
ctx = logger.WithTraceID(ctx, traceID)
ctx = logger.WithCorrelationID(ctx, correlationID)

// Read IDs from context:
logger.GetTraceID(ctx)
logger.GetCorrelationID(ctx)

// Context-bound logger — all log calls carry the IDs from ctx:
log := logger.WithContext(ctx)
log.Infof("user %s logged in", userID)
log.Errorf("operation failed: %v", err)

// Or use context-aware free functions:
logger.InfofContext(ctx, "processed %d records", n)
logger.ErrorfContext(ctx, "failed: %v", err)
```

**Configuration:**
```go
logger.SetLogLevel(slog.LevelDebug)  // default: Info
logger.GetLogger() *slog.Logger
logger.GetWriter() io.Writer         // stderr (used e.g. by GORM logger)
```

---

### `middlewares`

Gin middleware collection. Designed to be composed in a specific order for correct behaviour.

**Recommended middleware order:**
```go
router := gin.New()

// Create JWT verifier (Manager for sign+verify, or Verifier for verify-only)
jwtManager, err := jwt.NewManager(context.Background(), config.GetConfig())
if err != nil {
    log.Fatal(err)
}
// Or for API-only services: jwtVerifier, _ := jwt.NewVerifier(ctx, config.GetConfig())

router.Use(
    middlewares.RecoveryHandler,              // 1. catch panics first
    middlewares.RequestID(),                  // 2. inject trace/correlation IDs
    middlewares.LoggerMiddleware(),            // 3. log with IDs in context
    middlewares.Metrics(),                    // 4. Prometheus instrumentation
    middlewares.RequestTimeout(10*time.Second), // 5. bound all downstream handlers
    middlewares.CORS(),                       // 6. CORS headers (incl. X-2FA-Code)
    middlewares.MTLSMiddleware(),           // 7. client cert check (when MTLS_ENABLED)
    middlewares.AuthMiddleware(jwtManager),   // 8. JWT auth (pass *Manager or *Verifier)
    middlewares.RateLimiter(),                // 9. rate limit (requires Redis)
)
router.NoMethod(middlewares.NoMethodHandler())
router.NoRoute(middlewares.NoRouteHandler())
```

**Reference:**

| Middleware | Signature | Description |
|------------|-----------|-------------|
| `RecoveryHandler` | `gin.HandlerFunc` | Catches panics; returns structured JSON 500 with stack trace in logs |
| `RequestID()` | `gin.HandlerFunc` | Reads `X-Request-ID` / `X-Trace-ID` or generates UUID; injects into context and response headers |
| `TraceMiddleware()` | `gin.HandlerFunc` | Reads `X-Trace-Id`, `X-Correlation-Id`, `X-Request-Id` separately; use when upstream sends distinct trace and correlation IDs |
| `LoggerMiddleware()` | `gin.HandlerFunc` | Structured request log (method, path, status, latency, IP, user-agent, trace IDs) |
| `Metrics()` | `gin.HandlerFunc` | Prometheus counters, histogram, and in-flight gauge; uses route pattern to avoid high-cardinality labels |
| `RequestTimeout(d)` | `gin.HandlerFunc` | Adds `context.WithTimeout` to every request; no-op when `d <= 0` |
| `CORS()` | `gin.HandlerFunc` | CORS headers; global or per-origin (`CORS_FRONTEND` / `CORS_IPS`) from config |
| `AuthMiddleware(verifier)` | `jwt.TokenVerifier` → `gin.HandlerFunc` | Validates `Bearer` JWT; sets `user_id`, `original_user_id`, `is_impersonating` in context. Pass *jwt.Manager or *jwt.Verifier. |
| `MTLSMiddleware()` | `gin.HandlerFunc` | Requires verified TLS client cert when `MTLS_ENABLED`; sets `mtls_client_cn` in context |
| `RateLimiter()` | `gin.HandlerFunc` | Redis Lua single-round-trip rate limiter; sets `X-RateLimit-*` headers; supports IP or user keying and skip-paths |
| `NoMethodHandler()` | `gin.HandlerFunc` | 405 JSON response |
| `NoRouteHandler()` | `gin.HandlerFunc` | 404 JSON response |

**Prometheus metrics exposed by `Metrics()`:**

| Metric | Type | Labels |
|--------|------|--------|
| `http_requests_total` | Counter | `method`, `path`, `status` |
| `http_request_duration_seconds` | Histogram | `method`, `path`, `status` |
| `http_requests_in_flight` | Gauge | — |

Register the scrape endpoint separately:
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

---

### `otelx`

OpenTelemetry trace setup wired to `config.OpenTelemetry`. Supports **OTLP HTTP** or **Google Cloud Trace** (`OTEL_TRACES_EXPORTER=gcp`).

```go
import "github.com/turahe/pkg/otelx"

shutdown, enabled := otelx.Init(ctx, config.GetConfig().OpenTelemetry)
if enabled {
    defer shutdown(ctx)
}

// GORM instrumentation (auto when tracing enabled + OTEL_GORM_ENABLED=true)
db, _ := database.New(&cfg.Database, database.Options{})
```

**API:**

| Function | Description |
|----------|-------------|
| `otelx.LoadConfig()` | Returns `config.OpenTelemetryConfiguration` |
| `otelx.Init(ctx, cfg)` | Configures global TracerProvider; returns shutdown func + enabled flag |
| `otelx.RegisterGORM(db, opts)` | Manual GORM OTel plugin registration |
| `otelx.TracingEnabled(cfg)` | True when OTLP endpoint or `gcp` exporter is configured |

**GCP mode:** set `OTEL_TRACES_EXPORTER=gcp` and `GOOGLE_CLOUD_PROJECT` (or `OTEL_GCP_PROJECT_ID`). Uses Application Default Credentials and enables `X-Cloud-Trace-Context` propagation automatically.

---

### `sentryx`

Sentry error reporting wired to `config.Sentry`. Empty `SENTRY_DSN` disables Sentry with zero overhead.

```go
import "github.com/turahe/pkg/sentryx"

if sentryx.Init(config.GetConfig().Sentry) {
    defer sentryx.Flush(config.GetConfig().Sentry.FlushTimeout)
}
```

| Function | Description |
|----------|-------------|
| `sentryx.LoadConfig()` | Returns `config.SentryConfiguration` |
| `sentryx.Init(cfg)` | Initializes Sentry; returns enabled flag |
| `sentryx.Flush(timeout)` | Drains buffered events on shutdown |

---

### `mtls`

Mutual TLS helpers for backend HTTPS (require client cert) and gateway outbound connections. Config from `config.MTLS`.

**Server (backend):**
```go
import "github.com/turahe/pkg/mtls"

srv := &http.Server{Addr: ":" + cfg.Server.Port, Handler: router}
_ = mtls.ConfigureServer(srv)
go mtls.ListenConfigured(srv)

router.Use(middlewares.MTLSMiddleware()) // HTTP-layer client cert enforcement
```

**Client (gateway outbound):**
```go
transport, err := mtls.NewTransport(config.GetConfig().MTLS)
client := &http.Client{Transport: transport}
```

| Function | Description |
|----------|-------------|
| `mtls.LoadConfig()` | Returns `config.MTLSConfiguration` |
| `mtls.ServerTLSConfig(cfg)` | TLS config with `RequireAndVerifyClientCert` |
| `mtls.ClientTLSConfig(cfg)` | TLS config with client cert for outbound calls |
| `mtls.ConfigureServer(srv)` | Applies server TLS when `MTLS_ENABLED` |
| `mtls.ListenConfigured(srv)` | `ListenAndServeTLS` or plain HTTP |
| `mtls.RunGin(engine, addr)` | Gin drop-in for `engine.Run` with optional mTLS |
| `mtls.NewTransport(cfg)` | `http.Transport` for gateway upstream HTTPS |

When `MTLS_ENABLED=false`, all helpers fall back to plain HTTP.

---

### `handler`

Base handler for Gin with binding, pagination, error routing, and role checks.

```go
type BaseHandler struct{}

// Bind request body / query / URI by content-type and method.
func (c *BaseHandler) ValidateReqParams(ctx *gin.Context, params interface{}) error

// 422 validation error response (Laravel-style field map).
func (c *BaseHandler) HandleValidationError(ctx *gin.Context, serviceCode string, err error)

// Clamp page/size to defaults (1, 10) and max (100).
func (c *BaseHandler) NormalizePagination(pageNumber, pageSize int) (int, int)

// Extract and validate URL param; writes 422 and returns false if missing.
func (c *BaseHandler) GetIDFromParam(ctx *gin.Context, paramName, serviceCode string) (string, bool)

// Prefer reqID; fall back to URL param.
func (c *BaseHandler) GetIDFromRequestOrParam(ctx *gin.Context, reqID, paramName, serviceCode string) (string, bool)

// Route domain errors to correct HTTP status. Returns true if handled.
// Checks errors.Is(ErrNotFound) → 404, errors.Is(ErrUnauthorized) → 401,
// then notFoundMessages list, then falls back to 500.
func (c *BaseHandler) HandleServiceError(ctx *gin.Context, serviceCode string, err error, notFoundMessages ...string) bool

// Build SimplePaginationResponse from a slice, page info, and total.
func (c *BaseHandler) BuildPaginationResponse(data []interface{}, pageNumber, pageSize int, total int64) response.SimplePaginationResponse

// Read "user_id" string from Gin context (set by AuthMiddleware).
func (c *BaseHandler) GetCurrentUserID(ctx *gin.Context) (string, bool)

// Return true if any userRole matches any requiredRole.
func (c *BaseHandler) CheckUserHasRole(userRoles, requiredRoles []string) bool
```

---

### `response`

Standardised JSON responses with a 7-digit composite code (`HTTP(3) + Service(2) + Case(2)`).

**Success responses:**
```go
response.Ok(ctx)
response.OkWithMessage(ctx, message)
response.OkWithData(ctx, data)
response.Created(ctx, data)
response.Updated(ctx, data)
response.Deleted(ctx)
```

**Error responses:**
```go
response.FailWithMessage(ctx, message)
response.FailWithDetailed(ctx, httpStatus, serviceCode, caseCode, data, message)
response.ValidationError(ctx, serviceCode, err)          // 422 with field map
response.ValidationErrorSimple(ctx, serviceCode, field, message)
response.NotFoundError(ctx, serviceCode, caseCode, message)   // 404
response.UnauthorizedError(ctx, message)                       // 401
response.ForbiddenError(ctx, message)                          // 403
response.ConflictError(ctx, serviceCode, caseCode, message)    // 409
```

**Pagination responses:**
```go
response.SimplePaginated(ctx, data, pageNumber, pageSize, hasNext, hasPrev)
response.CursorPaginated(ctx, data, nextCursor, hasNext)
```

**Service codes:** `ServiceCodeCommon`, `ServiceCodeAuth`, `ServiceCodeTransaction`, `ServiceCodeWithdrawal`, `ServiceCodeUser`, `ServiceCodeAdmin`, `ServiceCodeMerchant`, `ServiceCodeSetting`, `ServiceCodeRole`, `ServiceCodePermission`, `ServiceCodeNotification`, `ServiceCodeIPWhitelist`, `ServiceCodeApiKey`, `ServiceCodeDeposit`, `ServiceCodeWallet`, `ServiceCodePhone`, `ServiceCodeEmail`, `ServiceCodeTwoFactor`, `ServiceCodeBank`, `ServiceCodeBankAccount`, and more. Case codes cover success, validation, auth, not-found, business logic, server errors, conflicts, email/phone change, 2FA, and bank/bank-account flows (see [response/codes.go](response/codes.go)).

---

### `jwt`

JWT token generation and validation with **HS256**, **RS256** (default), or **ES256**. Keys are loaded from config (env, file paths, or embedded PEM). No global state: use **Manager** (sign + verify), **Signer** (issue tokens only), or **Verifier** (validate only). Supports issuer, audience, key ID (kid), and token type (access, refresh, impersonation).

**All-in-one (single service):**
```go
manager, err := jwt.NewManager(ctx, config.GetConfig())
if err != nil {
    log.Fatal(err)
}
router.Use(middlewares.AuthMiddleware(manager))

token, _ := manager.GenerateToken(userID)
claims, _ := manager.ValidateToken(tokenString)
```

**Split services (auth server issues, API server only verifies):**
```go
// Auth/login service — needs private key or secret
signer, err := jwt.NewSigner(ctx, config.GetConfig())
token, _ := signer.GenerateToken(userID)
refresh, _ := signer.GenerateRefreshToken(userID)
impersonation, _ := signer.GenerateImpersonationToken(adminID, "admin", targetID, 15*time.Minute)

// API/gateway service — needs only public key or secret
verifier, err := jwt.NewVerifier(ctx, config.GetConfig())
router.Use(middlewares.AuthMiddleware(verifier))  // TokenVerifier interface
claims, _ := verifier.ValidateToken(tokenString)
```

**API summary:**
| Symbol | Description |
|--------|-------------|
| `jwt.NewManager(ctx, cfg)` | All-in-one; loads sign + verify keys; returns error |
| `jwt.NewSigner(ctx, cfg)` | Signing only (private key or secret) |
| `jwt.NewVerifier(ctx, cfg)` | Verification only (public key or secret) |
| `jwt.TokenVerifier` | Interface: `ValidateToken(string) (*Claims, error)`; implemented by *Manager and *Verifier |
| `manager.GenerateToken(id)` | Access token (default expiry) |
| `manager.GenerateTokenWithExpiry(id, expiry)` | Access token with custom expiry |
| `manager.GenerateRefreshToken(id)` | Refresh token |
| `manager.GenerateImpersonationToken(adminID, role, targetID, ttl)` | Short-lived impersonation token (max 30 min) |
| `manager.ValidateToken(token)` / `verifier.ValidateToken(token)` | Parse and verify; return *Claims or error |
| `jwt.ComparePassword(hashed, plain)` | bcrypt comparison |
| `jwt.GetCurrentUserUUID(ctx)` | Read `user_id` from Gin context (set by AuthMiddleware) |

**Config / env:** Default algorithm is **RS256**. For HS256 set `JWT_SIGNING_ALGORITHM=HS256` and `SERVER_SECRET`. For RS256/ES256 set `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY` (file path or inline PEM), or embed keys and assign to config before calling `NewManager`/`NewSigner`/`NewVerifier`. Optional: `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_KEY_ID`.

**Embed keys (no file paths):** Put PEM files in a package (e.g. `keys/`) and embed them; then set config so the JWT package uses the embedded bytes instead of paths:

```go
//go:embed keys/private.pem
var privateKeyPEM []byte

//go:embed keys/public.pem
var publicKeyPEM []byte

func main() {
    cfg := config.GetConfig()
    cfg.Server.JWTPrivateKeyPEM = privateKeyPEM
    cfg.Server.JWTPublicKeyPEM = publicKeyPEM
    manager, err := jwt.NewManager(ctx, cfg)
    // ...
}
```

---

### `crypto`

bcrypt password hashing.

```go
crypto.HashAndSalt(plainPassword []byte) string
crypto.ComparePassword(hashedPassword string, plainPassword []byte) bool
```

---

### `gcs`

Google Cloud Storage wrapper. Uses Application Default Credentials (ADC) or an explicit service account JSON file.

```go
gcs.Setup() error
gcs.GetClient() *storage.Client
gcs.GetBucket() *storage.BucketHandle
gcs.ReadObject(objectName string) ([]byte, error)
gcs.ReadObjectAsReader(objectName string) (io.ReadCloser, error)
gcs.WriteObject(objectName string, data []byte, contentType string) error
gcs.DeleteObject(objectName string) error
gcs.ObjectExists(objectName string) (bool, error)
gcs.ListObjects(prefix string) ([]string, error)
gcs.Close() error
```

---

### `types`

Shared types for handlers and repositories (no infrastructure dependencies).

```go
// Conditions is the WHERE clause map passed to repository methods.
type Conditions map[string]interface{}

// PageInfo holds offset-based pagination request (pageNumber, pageSize) with form/json tags.
type PageInfo struct {
    PageNumber int `form:"pageNumber" json:"pageNumber"`
    PageSize   int `form:"pageSize" json:"pageSize"`
}

// TimeRange holds a start/end pair (IANA or RFC3339 strings).
type TimeRange struct {
    Start string `json:"start"`
    End   string `json:"end"`
}
```

---

### `util`

```go
util.IsEmpty(value interface{}) bool
util.InAnySlice[T comparable](haystack []T, needle T) bool
util.RemoveDuplicates[T comparable](haystack []T) []T
util.FormatPhoneNumber(phone *string, countryCode *string) *string  // E.164 via nyaruka/phonenumbers
```

---

### `domain`

Domain errors and port interfaces. No dependencies on infrastructure.

```go
// Sentinel errors for handlers and use cases (use with errors.Is).
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
)
```

### `domain/port`

Port interfaces implemented by infrastructure (e.g. repositories).

```go
type GetByID interface {
    GetByID(ctx context.Context, id string) (interface{}, bool, error)
}
```

### `usecase`

Application service layer. Use cases depend only on domain and ports.

```go
type Runner interface { Run(ctx context.Context) error }
type Func func(ctx context.Context) error  // adapts func to Runner

// Example: get item by ID; returns domain.ErrNotFound when not found.
func GetItemByID(ctx context.Context, repo port.GetByID, id string) (interface{}, error)
```

### `repositories`

GORM-based base repository: CRUD, First, Find, Scan, SimplePagination, RawSQL, ExecSQL. Use with dependency injection (`NewBaseRepositoryWithDB`) or global DB (`NewBaseRepository`). Implements `IBaseRepository`; adapt to `domain/port` in your app.

```go
repo := repositories.NewBaseRepositoryWithDB(db.DB())
repo.Create(ctx, &model)
repo.First(ctx, &out, types.Conditions{"id = ?": id})
repo.SimplePagination(ctx, &model, &out, page, size, conditions, orders, "User", "Items")
```

---

## Documentation

All packages follow GoDoc conventions:

- Each package has a **`doc.go`** describing its role, responsibilities, constraints, and what it must not do.
- Exported symbols have comments that start with the symbol name and describe behavior clearly.
- Repository, adapter, and domain packages document contracts, error behavior, and context use.

Run `go doc` or view [pkg.go.dev](https://pkg.go.dev/github.com/turahe/pkg) for the full API.

---

### Minimal server

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "sync/atomic"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/turahe/pkg/config"
    "github.com/turahe/pkg/database"
    "github.com/turahe/pkg/middlewares"
    "github.com/turahe/pkg/mtls"
    "github.com/turahe/pkg/otelx"
    pkgredis "github.com/turahe/pkg/redis"
    "github.com/turahe/pkg/sentryx"
    "gorm.io/gorm/logger"
)

func main() {
    if err := config.Setup(""); err != nil {
        log.Fatal(err)
    }
    cfg := config.GetConfig()

    ctx := context.Background()
    otelShutdown, otelEnabled := otelx.Init(ctx, cfg.OpenTelemetry)
    sentryEnabled := sentryx.Init(cfg.Sentry)

    // Database
    db, err := database.New(&cfg.Database, database.Options{LogLevel: logger.Warn})
    if err != nil {
        log.Fatal(err)
    }

    // Redis (optional)
    if cfg.Redis.Enabled {
        if err := pkgredis.Setup(); err != nil {
            log.Printf("redis: %v", err)
        }
    }

    // Readiness gate
    var ready atomic.Bool
    ready.Store(true)

    // Router
    router := gin.New()
    router.Use(
        middlewares.RecoveryHandler,
        middlewares.RequestID(),
        middlewares.LoggerMiddleware(),
        middlewares.Metrics(),
        middlewares.RequestTimeout(10*time.Second),
        middlewares.CORS(),
        middlewares.MTLSMiddleware(),
    )
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))
    router.GET("/live", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })
    router.GET("/ready", func(c *gin.Context) {
        if !ready.Load() {
            c.JSON(http.StatusServiceUnavailable, gin.H{"status": "shutting_down"})
            return
        }
        if err := db.Health(c.Request.Context()); err != nil {
            c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // Application routes ...
    router.NoMethod(middlewares.NoMethodHandler())
    router.NoRoute(middlewares.NoRouteHandler())

    srv := &http.Server{
        Addr:              ":" + cfg.Server.Port,
        Handler:           router,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       10 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       60 * time.Second,
    }
    if err := mtls.ConfigureServer(srv); err != nil {
        log.Fatal(err)
    }
    go func() {
        if err := mtls.ListenConfigured(srv); err != nil && err != http.ErrServerClosed {
            log.Printf("server: %v", err)
        }
    }()

    // Graceful shutdown (25 s matches terminationGracePeriodSeconds: 30)
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    <-sig
    ready.Store(false) // signal readiness probe → 503 immediately

    shutCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
    defer cancel()
    srv.Shutdown(shutCtx)

    if sentryEnabled {
        sentryx.Flush(cfg.Sentry.FlushTimeout)
    }
    if otelEnabled {
        otelShutdown(shutCtx)
    }

    if cfg.Redis.Enabled {
        pkgredis.Close()
    }
    db.Close()
}
```

### Repository with dependency injection

```go
import (
    "github.com/turahe/pkg/database"
    "github.com/turahe/pkg/repositories"
    "github.com/turahe/pkg/types"
)

db, _ := database.New(&cfg.Database, database.Options{})
repo := repositories.NewBaseRepositoryWithDB(db.DB())

// Create
err := repo.Create(ctx, &myModel)

// Find with conditions
var results []MyModel
err = repo.Find(ctx, &results, types.Conditions{
    "status = ?": "active",
    "tenant_id = ?": tenantID,
})

// Pagination (pass preload names to avoid N+1)
total, err := repo.SimplePagination(ctx, &MyModel{}, &results, page, size,
    types.Conditions{"status = ?": "active"},
    []string{"created_at DESC"},
    "User", "Items",  // preloads
)
```

### Clean architecture wiring

```go
// domain/port/repository.go
type ItemRepository interface {
    GetByID(ctx context.Context, id string) (*Item, bool, error)
}

// infrastructure/repository/item.go
type itemRepo struct{ base repositories.IBaseRepository }

func (r *itemRepo) GetByID(ctx context.Context, id string) (*Item, bool, error) {
    var item Item
    notFound, err := r.base.First(ctx, &item, types.Conditions{"id = ?": id})
    return &item, !notFound, err
}

// usecase/get_item.go
func GetItem(ctx context.Context, repo port.ItemRepository, id string) (*Item, error) {
    item, found, err := repo.GetByID(ctx, id)
    if err != nil { return nil, err }
    if !found { return nil, domain.ErrNotFound }
    return item, nil
}
```

---

## Environment Variables

Copy `.env.example` to `.env` and call `config.Setup("")`:

```bash
cp .env.example .env
```

### Server

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP listen port |
| `SERVER_SECRET` | — | JWT secret for **HS256** only; required when `JWT_SIGNING_ALGORITHM=HS256` |
| `SERVER_MODE` | `debug` | Gin mode: `debug`, `release`, `test` |
| `SERVER_ACCESS_TOKEN_EXPIRY` | `1` | Access token lifetime (hours) |
| `SERVER_REFRESH_TOKEN_EXPIRY` | `7` | Refresh token lifetime (days) |
| `SERVER_SESSION_EXPIRY` | `24` | Session lifetime (hours) |
| `SERVER_TIMEZONE` | `UTC` | IANA timezone (also default for DB when `DATABASE_TIMEZONE` empty) |
| `APP_ENV` | — | Deployment environment; fallback for OTEL/Sentry when their env vars are empty |

### CORS

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_GLOBAL` | `true` | Allow all origins (`*`) |
| `CORS_FRONTEND` | — | Frontend origin URL when `CORS_GLOBAL=false` (e.g. `http://localhost:3000`) |
| `CORS_IPS` | — | Comma-separated allowed origins when `CORS_GLOBAL=false` |

Allowed headers include `Authorization`, `X-CSRF-Token`, and `X-2FA-Code`.

### JWT

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SIGNING_ALGORITHM` | **`RS256`** | `HS256` · `RS256` · `ES256` |
| `JWT_PRIVATE_KEY` | — | File path or inline PEM private key (RS256/ES256) |
| `JWT_PUBLIC_KEY` | — | File path or inline PEM public key (RS256/ES256) |
| `JWT_ISSUER` | — | Issuer (`iss`) claim |
| `JWT_AUDIENCE` | — | Audience (`aud`), comma-separated |
| `JWT_KEY_ID` | — | Key ID (`kid`) in JWT header |

Keys can be embedded via `config.Server.JWTPrivateKeyPEM` / `JWTPublicKeyPEM` (see [jwt](#jwt) section).

### Database

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_DRIVER` | `mysql` | `mysql` · `postgres` · `sqlite` · `sqlserver` · `cloudsql-mysql` · `cloudsql-postgres` |
| `DATABASE_HOST` | `127.0.0.1` | |
| `DATABASE_PORT` | `3306` | |
| `DATABASE_USERNAME` | — | |
| `DATABASE_PASSWORD` | — | |
| `DATABASE_DBNAME` | — | |
| `DATABASE_SSLMODE` | `false` | |
| `DATABASE_LOGMODE` | `false` | Enable GORM query logging |
| `DATABASE_TIMEZONE` | — | IANA timezone for DB session (`mysql`/`postgres`/Cloud SQL); empty uses `SERVER_TIMEZONE` |
| `DATABASE_MAX_IDLE_CONNS` | `0` (→ 10) | Connection pool idle size |
| `DATABASE_MAX_OPEN_CONNS` | `0` (→ 30) | Connection pool max open |
| `DATABASE_CONN_MAX_LIFETIME` | `0` (→ 30 min) | Connection lifetime (minutes) |
| `DATABASE_CLOUD_SQL_INSTANCE` | — | `project:region:instance` for Cloud SQL |
| `DATABASE_*_SITE` | — | Same keys with `_SITE` suffix for secondary DB |
| `DATABASE_TIMEZONE_SITE` | — | Site DB timezone; empty uses `SERVER_TIMEZONE` |

### Redis

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_ENABLED` | `false` | |
| `REDIS_HOST` | `127.0.0.1` | |
| `REDIS_PORT` | `6379` | |
| `REDIS_PASSWORD` | — | |
| `REDIS_DB` | `0` | Database index (ignored in cluster mode) |
| `REDIS_CLUSTER_MODE` | `false` | Enable cluster client |
| `REDIS_CLUSTER_NODES` | — | Comma-separated `host:port` list |
| `REDIS_POOL_SIZE` | `0` (client default) | Max connections per node |
| `REDIS_MIN_IDLE_CONNS` | `0` | Min idle connections |
| `REDIS_READ_TIMEOUT_SEC` | `0` (no timeout) | Read timeout |
| `REDIS_WRITE_TIMEOUT_SEC` | `0` (no timeout) | Write timeout |

### Rate Limiter

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMITER_ENABLED` | `false` | Requires Redis |
| `RATE_LIMITER_REQUESTS` | `100` | Requests allowed per window |
| `RATE_LIMITER_WINDOW` | `60` | Window size (seconds) |
| `RATE_LIMITER_KEY_BY` | `ip` | `ip` or `user` |
| `RATE_LIMITER_SKIP_PATHS` | — | Comma-separated paths (e.g. `/health,/metrics`) |

### GCS

| Variable | Default | Description |
|----------|---------|-------------|
| `GCS_ENABLED` | `false` | |
| `GCS_BUCKET_NAME` | — | |
| `GCS_CREDENTIALS_FILE` | — | Path to service account JSON; omit to use ADC |

### OpenTelemetry

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_TRACES_EXPORTER` | `otlp` | `otlp` or `gcp` (Google Cloud Trace via ADC) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | — | OTLP HTTP endpoint; required when exporter is `otlp` |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | — | Optional trace-specific endpoint override |
| `OTEL_EXPORTER_OTLP_INSECURE` | `true` | Use HTTP instead of HTTPS for OTLP export |
| `OTEL_EXPORTER_OTLP_HEADERS` | — | Comma-separated `key=value` OTLP headers |
| `OTEL_SERVICE_NAME` | — | Service name resource attribute; `otelx` defaults to `app` |
| `OTEL_ENVIRONMENT` | — | Deployment environment; falls back to `APP_ENV` |
| `OTEL_SERVICE_VERSION` | — | Service version; falls back to `SENTRY_RELEASE` |
| `OTEL_TRACES_SAMPLER_ARG` | `1.0` | Parent-based ratio sampler (0..1) |
| `OTEL_SHUTDOWN_TIMEOUT` | `5s` | Tracer provider shutdown timeout |
| `OTEL_GORM_ENABLED` | `true` | Register GORM OTel plugin when tracing is enabled |
| `OTEL_GCP_PROJECT_ID` | — | GCP project for Cloud Trace; falls back to `GOOGLE_CLOUD_PROJECT` |
| `OTEL_GCP_PROPAGATOR` | `false` | Enable `X-Cloud-Trace-Context` propagation (auto-on for `gcp` exporter) |

Initialize tracing with `otelx.Init(ctx, config.GetConfig().OpenTelemetry)` after `config.Setup`. For Google Cloud Trace set `OTEL_TRACES_EXPORTER=gcp`. GORM is instrumented automatically when tracing is enabled and `OTEL_GORM_ENABLED=true` (default).

### Sentry

| Variable | Default | Description |
|----------|---------|-------------|
| `SENTRY_DSN` | — | Sentry DSN; empty disables Sentry |
| `SENTRY_ENVIRONMENT` | — | Deployment environment; falls back to `APP_ENV` |
| `SENTRY_RELEASE` | — | Release identifier (git SHA or semver) |
| `SENTRY_SERVER_NAME` | — | Server name tag; `sentryx` defaults to `app` |
| `SENTRY_DEBUG` | `false` | Enable Sentry SDK debug logging |
| `SENTRY_ATTACH_STACKTRACE` | `true` | Attach stack traces to non-panic messages |
| `SENTRY_SAMPLE_RATE` | `1.0` | Error event sample rate (0..1) |
| `SENTRY_TRACES_SAMPLE_RATE` | `0.0` | Performance transaction sample rate (0..1) |
| `SENTRY_FLUSH_TIMEOUT` | `2s` | Flush timeout at shutdown |

Initialize Sentry with `sentryx.Init(config.GetConfig().Sentry)` after `config.Setup`.

### mTLS

| Variable | Default | Description |
|----------|---------|-------------|
| `MTLS_ENABLED` | `false` | Enable mTLS server listen and `MTLSMiddleware` |
| `MTLS_CA_CERT` | `/etc/mtls/ca.crt` | Shared CA bundle |
| `MTLS_SERVER_CERT` | `/etc/mtls/server.crt` | Backend server certificate |
| `MTLS_SERVER_KEY` | `/etc/mtls/server.key` | Backend server private key |
| `MTLS_CLIENT_CERT` | `/etc/mtls/gateway.crt` | Gateway/client certificate (outbound) |
| `MTLS_CLIENT_KEY` | `/etc/mtls/gateway.key` | Gateway/client private key |
| `MTLS_SKIP_PATHS` | `/live,/ready,/metrics` | Paths bypassing `MTLSMiddleware` |

See [`mtls`](#mtls) for server/client wiring examples.

---

## Production Wiring Example

See [`cmd/example/main.go`](cmd/example/main.go) for a complete wiring of:

- `config.Setup` and optional `otelx` / `sentryx` initialization
- Dependency-injected database with health check and optional GORM tracing
- Optional Redis setup and graceful `Close()`
- Optional mTLS via `mtls.ConfigureServer` + `MTLSMiddleware`
- `gin.New()` with explicit middleware stack (recovery → trace → logging → metrics → timeout → CORS → mTLS)
- `/live` (liveness), `/ready` (readiness with component checks), `/metrics` (Prometheus)
- HTTP server with all timeouts set
- Graceful shutdown: readiness gate → `srv.Shutdown(25s)` → Sentry flush → OTel shutdown → Redis close → DB close

### Kubernetes probes

```yaml
livenessProbe:
  httpGet:
    path: /live
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5

terminationGracePeriodSeconds: 30
```

### Docker

```bash
docker build -t myapp .
docker run --env-file .env -p 8080:8080 myapp
```

The multi-stage `Dockerfile` uses `golang:1.26-alpine` to build and `gcr.io/distroless/base-debian12:nonroot` as the runtime image. Runs as UID 65532 (non-root).

---

## Testing

Run all tests:
```bash
go test ./...
```

With race detector (requires CGO):
```bash
CGO_ENABLED=1 go test -race ./...
```

With coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Run tests with Docker Compose

Run the full test suite (including integration tests) without local Go or databases:

```bash
make test-docker
```

Or with docker compose directly:

```bash
docker compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from test
docker compose -f docker-compose.test.yml down -v
```

The test runner waits for Redis, MySQL, and Postgres to be healthy, then runs `go test -v -race -count=1 ./...`. Coverage is written to `coverage.out` and a summary is printed.

### Integration tests (Redis · MySQL · Postgres) — local

Start services with Docker Compose:
```bash
docker compose up -d
```

Run with environment variables:
```bash
REDIS_ENABLED=true REDIS_HOST=127.0.0.1 REDIS_PORT=6379 \
DATABASE_DRIVER=mysql DATABASE_HOST=127.0.0.1 DATABASE_PORT=3306 \
DATABASE_USERNAME=root DATABASE_PASSWORD=root DATABASE_DBNAME=testdb \
go test ./...
```

Integration tests skip automatically when services are unavailable. CI runs the full matrix with these services via GitHub Actions.

**Packages with tests:** `config`, `crypto`, `database`, `gcs`, `handler`, `jwt`, `logger`, `middlewares`, `mtls`, `otelx`, `redis`, `repositories`, `response`, `sentryx`, `types`, `util`.

---

## License

MIT
