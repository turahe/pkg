# Documentation

Guides for [github.com/turahe/pkg](https://github.com/turahe/pkg) — a collection of production-ready Go packages for building web services.

API reference lives on [pkg.go.dev](https://pkg.go.dev/github.com/turahe/pkg). Each package also has a `doc.go` describing role, responsibilities, and constraints.

## Guides

| Guide | Description |
|-------|-------------|
| [Getting Started](getting-started.md) | Install, configure, and wire a minimal Gin server |
| [Architecture](architecture.md) | Clean architecture layers and dependency rules |
| [Environment Variables](environment.md) | Full env var reference (also see [`.env.example`](../.env.example)) |
| [Contributing / Development](development.md) | Tests, lint, Docker, Makefile targets |

## Packages

| Package | Layer | Description |
|---------|-------|-------------|
| [config](packages/config.md) | Infrastructure | Env / `.env` configuration |
| [database](packages/database.md) | Infrastructure | GORM + Cloud SQL, pool, health |
| [redis](packages/redis.md) | Infrastructure | Standalone / cluster Redis client |
| [logger](packages/logger.md) | Infrastructure | Structured slog + Cloud Trace fields |
| [jwt](packages/jwt.md) | Infrastructure | JWT sign / verify (HS256 / RS256 / ES256) |
| [crypto](packages/crypto.md) | Infrastructure | bcrypt password helpers |
| [gcs](packages/gcs.md) | Infrastructure | Google Cloud Storage client |
| [otelx](packages/otelx.md) | Infrastructure | OpenTelemetry tracing (OTLP / GCP) |
| [sentryx](packages/sentryx.md) | Infrastructure | Sentry init / flush |
| [mtls](packages/mtls.md) | Infrastructure | Mutual TLS server & client |
| [middlewares](packages/middlewares.md) | Adapter | Gin middleware stack |
| [handler](packages/handler.md) | Adapter | Base Gin handler helpers |
| [response](packages/response.md) | Adapter | Standardized JSON responses |
| [repositories](packages/repositories.md) | Adapter | Generic GORM base repository |
| [domain](packages/domain.md) | Domain | Shared sentinel errors |
| [domain/port](packages/domain.md#domainport) | Domain | Use-case ports (interfaces) |
| [usecase](packages/usecase.md) | Application | Use-case scaffolding |
| [types](packages/types.md) | Shared | Pagination / query helper types |
| [util](packages/util.md) | Shared | Generic helpers |

## Quick links

- [Root README](../README.md) — package overviews and production wiring
- [Changelog](../CHANGELOG.md)
- [License](../LICENSE) (MIT)
