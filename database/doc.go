/*
Package database provides a GORM-based database layer with support for MySQL, Postgres, SQLite, SQL Server, and Google Cloud SQL (Postgres/MySQL with optional IAM and Private IP).

Role in architecture:
  - Infrastructure: connects to the database, configures pool, exposes *gorm.DB. No business logic.

Responsibilities:
  - Build DSN and open connection (standard drivers or Cloud SQL connector).
  - Apply connection pool limits (MaxOpenConns, MaxIdleConns, ConnMaxLifetime, ConnMaxIdle).
  - Expose Health(ctx) with PingTimeout-bounded ping and Close() for cleanup (including Cloud SQL connector).
  - Optional GORM logger with SQL redaction (passwords, tokens, card numbers) and slow query logging.
  - Legacy compat: Setup/GetDB/GetDBSite/HealthCheck/Cleanup/IsAlive for global singleton usage.

Constraints:
  - Driver selection is by config only; no runtime provider switching.
  - No retry, circuit breaker, or fallback; single connection path per Database instance.
  - Cloud SQL IAM and Private IP are set via Options at construction, not from config.

This package must NOT:
  - Define use-case or domain logic; only connection and pool management.
  - Execute business queries; that belongs in repositories or use-case adapters.
*/
package database
