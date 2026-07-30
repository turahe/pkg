# database

GORM-based database layer with MySQL, Postgres, SQLite, SQL Server, and Google Cloud SQL (Postgres/MySQL) including optional IAM auth and Private IP.

**Import:** `github.com/turahe/pkg/database`

## Role

Infrastructure: open connections, configure pools, expose `*gorm.DB`, health checks. No business queries.

## Recommended API (DI)

```go
db, err := database.New(&cfg.Database, database.Options{
    LogLevel: logger.Warn,
})
// Cloud SQL:
db, err := database.New(&cfg.Database, database.Options{
    UseIAM:       true,
    UsePrivateIP: true,
    LogLevel:     logger.Warn,
})

sqlDB := db.DB()           // *gorm.DB
err = db.Health(ctx)
err = db.Close()
```

Functional options: `WithIAM`, `WithPrivateIP`, `WithLogLevel`, `WithMaxOpenConns`, `WithMaxIdleConns`, `WithConnMaxLifetime`, `WithConnMaxIdleTime`, `WithOpenTelemetry`, `WithProductionPoolDefaults`.

## Legacy compat

`Setup`, `GetDB`, `GetDBSite`, `HealthCheck`, `Cleanup`, `IsAlive`, `CreateDatabaseConnection` — global singletons for migration; prefer `New`.

## Drivers

`mysql` · `postgres` · `sqlite` · `sqlserver` · `cloudsql-mysql` · `cloudsql-postgres`

## Features

- Connection pool from config (defaults: 10 idle, 30 open, 30 min lifetime)
- SQL redaction in logs (passwords, tokens, card numbers) via `NewFintechLogger`
- Optional OpenTelemetry GORM plugin (`OTEL_GORM_ENABLED`, or `WithOpenTelemetry`)
- Session timezone from `DATABASE_TIMEZONE` / `SERVER_TIMEZONE`

## Constraints

- Driver selected by config only; no runtime provider switching.
- No retry / circuit breaker.
- Cloud SQL IAM and Private IP via `Options`, not env.
- Must not define domain/use-case logic or run business queries (use repositories).

## See also

- [Environment — Database](../environment.md#database)
- [repositories](repositories.md)
- [otelx](otelx.md)
