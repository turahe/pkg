/*
Package middlewares provides Gin HTTP middleware for cross-cutting concerns: recovery, tracing, logging, metrics, timeout, CORS, auth, and rate limiting.

Role in architecture:
  - Adapters: sit between the HTTP server and handlers; no business logic, only request/response and infra calls.

Responsibilities:
  - Recovery: catch panics, log stack, return JSON 500.
  - Tracing: inject request/trace/correlation IDs from headers or generate UUIDs; store in context for logger.
  - Logging: log each request (method, path, status, latency, IP) with context-bound logger.
  - Metrics: expose Prometheus counters, histogram, and in-flight gauge (route pattern as label).
  - Timeout: set request context deadline so downstream DB/Redis respect it.
  - CORS: set Access-Control-* headers from config.
  - Auth: validate Bearer JWT via jwt.TokenVerifier (Manager or Verifier); set user_id and impersonation fields in context. Use AuthMiddleware(verifier) with jwt.NewManager or jwt.NewVerifier for verification-only services.
  - Rate limiting: Redis sliding-window (ZSET + Lua); 429 when exceeded; skip paths configurable.

Constraints:
  - Rate limiter requires Redis enabled and config.RateLimiter.Enabled; fails open on Redis error.
  - Auth requires a non-nil jwt.TokenVerifier (*jwt.Manager or *jwt.Verifier); pass it to AuthMiddleware(verifier).
  - No provider switching or fallbacks inside middleware; config is read once at middleware build.

This package must NOT:
  - Contain use-case or domain logic; only HTTP and infra concerns.
*/
package middlewares
