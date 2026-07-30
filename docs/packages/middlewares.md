# middlewares

Gin HTTP middleware for cross-cutting concerns: recovery, tracing, logging, metrics, timeout, CORS, auth, mTLS, and rate limiting.

**Import:** `github.com/turahe/pkg/middlewares`

## Role

Adapters between the HTTP server and handlers. No domain or use-case logic.

## Recommended stack order

```go
r := gin.New()
r.Use(
    middlewares.RecoveryHandler,
    middlewares.TraceMiddleware(),       // or CloudTraceMiddleware()
    middlewares.LoggerMiddleware(),
    middlewares.Metrics(),
    middlewares.RequestTimeout(30*time.Second),
    middlewares.CORS(),
    // middlewares.MTLSMiddleware(),
    // middlewares.AuthMiddleware(verifier),
    // middlewares.RateLimiter(),
)
r.NoMethod(middlewares.NoMethodHandler())
r.NoRoute(middlewares.NoRouteHandler())
```

## Middleware reference

| Middleware | Purpose |
|------------|---------|
| `RecoveryHandler` | Panic → log stack → JSON 500 |
| `TraceMiddleware` / `CloudTraceMiddleware` / `RequestID` | Request / trace / correlation IDs |
| `LoggerMiddleware` | Method, path, status, latency, IP |
| `Metrics` / `HTTPInstrumentation` | Prometheus counters, histogram, in-flight |
| `RequestTimeout` | Context deadline for downstream I/O |
| `CORS` | From `CORS_*` config |
| `AuthMiddleware(verifier)` | Bearer JWT via `jwt.TokenVerifier` |
| `MTLSMiddleware` | Require verified client cert when enabled |
| `RateLimiter` | Redis sliding-window (ZSET + Lua) |

## Auth

```go
verifier, err := jwt.NewVerifier(ctx, cfg.Server)
r.Use(middlewares.AuthMiddleware(verifier))
```

Sets `user_id` and impersonation fields in Gin context. Requires non-nil verifier.

## Rate limiter

Requires `REDIS_ENABLED` and `RATE_LIMITER_ENABLED`. **Fails open** on Redis errors. Key by `ip` or `user`; skip paths configurable.

## Headers / context

- `X-Cloud-Trace-Context`, trace ID, correlation ID, request ID constants
- `ContextMTLSClientCN` — client certificate CN after mTLS middleware
- Correlation ID is never overwritten once set

## Constraints

- Config is read once at middleware construction.
- Must not contain business rules.

## See also

- [jwt](jwt.md)
- [mtls](mtls.md)
- [redis](redis.md)
- [Getting Started](../getting-started.md)
