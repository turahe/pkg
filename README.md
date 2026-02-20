# Go Package Collection

[![Go Reference](https://pkg.go.dev/badge/github.com/turahe/pkg.svg)](https://pkg.go.dev/github.com/turahe/pkg)
[![Test](https://github.com/turahe/pkg/actions/workflows/test.yml/badge.svg)](https://github.com/turahe/pkg/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/turahe/pkg)](https://goreportcard.com/report/github.com/turahe/pkg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A collection of production-ready Go packages for building web services: database, Redis (standard + cluster), JWT, crypto, GCS, structured logging, HTTP middleware, Prometheus metrics, graceful shutdown, and utilities. Follows clean architecture boundaries — domain, use-case, and infrastructure are separate.

## Contents

- [Installation](#installation)
- [Packages](#packages)
  - [config](#config)
  - [database](#database)
  - [redis](#redis)
  - [logger](#logger)
  - [middlewares](#middlewares)
  - [handler](#handler)
  - [response](#response)
  - [jwt](#jwt)
  - [crypto](#crypto)
  - [gcs](#gcs)
  - [types](#types)
  - [util](#util)
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

Requires Go 1.21+.

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

Structured logging built on `log/slog`. Outputs Google Cloud Logging-compatible JSON with `severity`, `time`, `message`, `trace_id`, `correlation_id`, `sourceLocation`, and optional `fields`. The underlying writer is lazy-initialized on first log write (no I/O at startup).

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
logger.GetWriter() io.Writer         // file + stderr multi-writer
```

---

### `middlewares`

Gin middleware collection. Designed to be composed in a specific order for correct behaviour.

**Recommended middleware order:**
```go
router := gin.New()
router.Use(
    middlewares.RecoveryHandler,              // 1. catch panics first
    middlewares.RequestID(),                  // 2. inject trace/correlation IDs
    middlewares.LoggerMiddleware(),           // 3. log with IDs in context
    middlewares.Metrics(),                    // 4. Prometheus instrumentation
    middlewares.RequestTimeout(10*time.Second), // 5. bound all downstream handlers
    middlewares.CORS(),                       // 6. CORS headers
    middlewares.AuthMiddleware(),             // 7. JWT auth (protected routes)
    middlewares.RateLimiter(),               // 8. rate limit (requires Redis)
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
| `CORS()` | `gin.HandlerFunc` | CORS headers; global or per-origin from config |
| `AuthMiddleware()` | `gin.HandlerFunc` | Validates `Bearer` JWT; sets `user_id` in Gin context |
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

**Service codes:** `ServiceCodeCommon`, `ServiceCodeAuth`, `ServiceCodeTransaction`, `ServiceCodeWallet`, `ServiceCodeUser`, `ServiceCodeAdmin`, `ServiceCodeMerchant`, `ServiceCodeRole`, `ServiceCodePermission`, `ServiceCodeNotification`, `ServiceCodeApiKey`, `ServiceCodeDeposit`, and more.

---

### `jwt`

HS256 JWT tokens with configurable expiry.

```go
jwt.Init()                                              // reads SERVER_SECRET from config
jwt.GenerateToken(id uuid.UUID, email, name string) (string, error)
jwt.GenerateTokenWithExpiry(id uuid.UUID, email, name string, expiry time.Duration) (string, error)
jwt.GenerateRefreshToken(id uuid.UUID, email, name string) (string, error)
jwt.ValidateToken(tokenString string) (*Claims, error)  // returns Claims{UUID string}
jwt.ComparePassword(hashed, plain string) bool          // bcrypt comparison
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

```go
// Conditions is the WHERE clause map passed to repository methods.
type Conditions map[string]interface{}

// TimeRange holds a start/end pair (IANA or RFC3339 strings).
type TimeRange struct {
    Start string
    End   string
}
```

---

### `util`

```go
util.IsEmpty(value interface{}) bool
util.InAnySlice[T comparable](haystack []T, needle T) bool
util.RemoveDuplicates[T comparable](haystack []T) []T
util.FormatPhoneNumber(number, defaultRegion string) (string, error)  // E.164
```

---

## Usage

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
    pkgredis "github.com/turahe/pkg/redis"
    "gorm.io/gorm/logger"
)

func main() {
    cfg := config.GetConfig()

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
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

Copy `env.example` to `.env` and call `config.Setup("")`:

```bash
cp env.example .env
```

### Server

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP listen port |
| `SERVER_SECRET` | — | JWT signing secret (**required**) |
| `SERVER_MODE` | `debug` | Gin mode: `debug`, `release`, `test` |
| `SERVER_ACCESS_TOKEN_EXPIRY` | `1` | Access token lifetime (hours) |
| `SERVER_REFRESH_TOKEN_EXPIRY` | `7` | Refresh token lifetime (days) |
| `SERVER_SESSION_EXPIRY` | `24` | Session lifetime (hours) |
| `CORS_GLOBAL` | `true` | Allow all origins |

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
| `DATABASE_MAX_IDLE_CONNS` | `0` (→ 10) | Connection pool idle size |
| `DATABASE_MAX_OPEN_CONNS` | `0` (→ 30) | Connection pool max open |
| `DATABASE_CONN_MAX_LIFETIME` | `0` (→ 30 min) | Connection lifetime (minutes) |
| `DATABASE_CLOUD_SQL_INSTANCE` | — | `project:region:instance` for Cloud SQL |
| `DATABASE_*_SITE` | — | Same keys with `_SITE` suffix for secondary DB |

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

---

## Production Wiring Example

See [`cmd/example/main.go`](cmd/example/main.go) for a complete wiring of:

- Dependency-injected database with health check
- Optional Redis setup and graceful `Close()`
- `gin.New()` with explicit middleware stack (recovery → trace → logging → metrics → timeout)
- `/live` (liveness), `/ready` (readiness with component checks), `/metrics` (Prometheus)
- HTTP server with all timeouts set
- Graceful shutdown: readiness gate → `srv.Shutdown(25s)` → Redis close → DB close

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

The multi-stage `Dockerfile` uses `golang:1.25-alpine` to build and `gcr.io/distroless/base-debian12:nonroot` as the runtime image. Runs as UID 65532 (non-root).

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

### Integration tests (Redis · MySQL · Postgres)

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

Integration tests skip automatically when services are unavailable. CI runs the full matrix (Go 1.21–1.25.4) with these services via GitHub Actions.

**Packages with tests:** `config`, `crypto`, `database`, `gcs`, `handler`, `jwt`, `logger`, `middlewares`, `redis`, `repositories`, `response`, `types`, `util`.

---

## License

MIT
