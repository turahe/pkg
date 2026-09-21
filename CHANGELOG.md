# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`storage` package**: Multi-driver object storage (`STORAGE_DRIVER=gcs|s3|r2`) with singleton `Setup` / `SetupContext`, `ErrNotInitialized`, and provider-neutral operations. See [docs/packages/storage.md](docs/packages/storage.md) for env vars, RustFS/R2 examples, and migration steps.
- **RustFS in Docker Compose**: Local/test S3-compatible endpoint on `:9000` / console `:9001`.

### Removed

- **`gcs` package**: Import path `github.com/turahe/pkg/gcs` is gone. Use `github.com/turahe/pkg/storage`.
- **`GCS_ENABLED` / `GCS_BUCKET_NAME`**: Replaced by `STORAGE_DRIVER` / `STORAGE_BUCKET` (no silent aliases). GCS credentials path remains `GCS_CREDENTIALS_FILE`.
- **`GetClient` / `GetBucket`**: No longer exposed; call package operations only.

## [0.5.6] - 2026-09-14

### Added

- **otelx OTLP/gRPC**: Export traces over gRPC via `OTEL_TRACES_EXPORTER=otlp_grpc` or `OTEL_EXPORTER_OTLP_PROTOCOL=grpc` (default remains OTLP HTTP).
- **otelx gRPC instrumentation**: `GRPCServerOption` / `GRPCDialOption` (and handler helpers) wrapping otelgrpc stats handlers.
- **sentryx gRPC**: `ServerOptions` / `DialOptions` and unary/stream interceptors via `sentry-go/grpc`.
- **JWT `actor_type`**: `ResolveActorType` maps table/actor names (`admins` → `admin`, `users`/`User` → `user`, empty → `service`, `system` → `system`). `GenerateToken` / `GenerateRefreshToken` accept optional actor/table; impersonation tokens set `actor_type=user`. Auth middleware sets `actor_type` in Gin context.
- **Husky**: `.husky/hooks/pre-commit` runs `go fmt ./...` and `go test -v ./...` before commit.

### Fixed

- **Redis / rate-limiter tests**: Integration and rate-limiter setups skip when Redis requires auth (`NOAUTH`) or Setup fails, instead of failing the suite; honor `REDIS_PASSWORD` when set.

### Changed

- **Config / docs**: `OpenTelemetryConfiguration.Protocol`, `.env.example`, README, and package docs updated for OTLP gRPC and JWT actor types.

## [0.5.5] - 2026-08-11

### Fixed

- **Lint**: Clear `golangci-lint` findings — `mtls` sets `ReadHeaderTimeout`, Redis `Available` uses `DialContext`, GCS uses `WithAuthCredentialsFile`, error-string / staticcheck / gofmt cleanup; test exclusions for `noctx`/`ineffassign`.

## [0.5.4] - 2026-08-09

### Removed

- **JWT HS256**: Symmetric JWT signing is no longer supported. Use `RS256` (default) or `ES256` with `JWT_PRIVATE_KEY` / `JWT_PUBLIC_KEY` (or embedded PEM). Setting `JWT_SIGNING_ALGORITHM=HS256` now returns an error.
- **`SERVER_SECRET` / `config.Server.Secret`**: Removed. JWT no longer uses a shared secret; configure asymmetric keys only.

## [0.3.7] - 2026-02-28

### Added

- **JWT keys from config bytes**: `config.Server.JWTPrivateKeyPEM` and `JWTPublicKeyPEM` (optional `[]byte`) for loading keys from embedded or in-memory PEM (e.g. `//go:embed`). When set, used instead of `JWT_PRIVATE_KEY` / `JWT_PUBLIC_KEY`.

### Changed

- **JWT env vars**: Renamed `JWT_PRIVATE_KEY_PATH` and `JWT_PUBLIC_KEY_PATH` to **`JWT_PRIVATE_KEY`** and **`JWT_PUBLIC_KEY`**. Config fields renamed to `JWTPrivateKey` and `JWTPublicKey`.
- **JWT key value format**: `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY` accept either a **file path** or **inline PEM** (string containing `-----BEGIN`); the package detects format automatically.
- **Docs**: `.env.example`, README, and `jwt/doc.go` updated for new env names and embed usage.

## [0.3.6] - 2026-02-28

### Removed

- **Google Secret Manager** support from `config`, `jwt`, and `database` packages. JWT and database credentials are now loaded only from environment variables or file paths. Removed env vars: `JWT_SECRET_MANAGER_*`, `DATABASE_SECRET_MANAGER_PROJECT_ID`, `DATABASE_PASSWORD_SECRET_NAME` (and `_SITE` variants). Dependency `cloud.google.com/go/secretmanager` removed from `go.mod`.

### Changed

- **Rate limiter** (`middlewares`): switched from fixed-window to **sliding-window** algorithm using Redis ZSET + Lua (single round-trip). Config and Redis client are read once at middleware build time; added constants for default window and key prefix; added `toInt64` for script result handling.
- **Rate limiter tests**: Redis-dependent tests now **skip** when Redis is unreachable (e.g. `go test ./...` without Redis) instead of failing. Run `make test-docker` or start Redis to execute them.
- **`.env.example`**: Removed JWT and database Secret Manager variables; updated JWT section to document env/file only and optional `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_KEY_ID`.
- **Documentation**: `middlewares/doc.go` updated to describe sliding-window rate limiter.

### Fixed

- **crypto.ComparePassword**: When the stored hash is empty or shorter than 60 characters (invalid bcrypt), the function now returns `false` without calling bcrypt or logging an error, avoiding `crypto/bcrypt: hashedSecret too short` errors and log noise.

[Unreleased]: https://github.com/turahe/pkg/compare/v0.5.6...HEAD
[0.5.6]: https://github.com/turahe/pkg/compare/v0.5.5...v0.5.6
[0.5.5]: https://github.com/turahe/pkg/compare/v0.5.4...v0.5.5
[0.5.4]: https://github.com/turahe/pkg/compare/v0.5.3...v0.5.4
[0.3.7]: https://github.com/turahe/pkg/releases/tag/v0.3.7
[0.3.6]: https://github.com/turahe/pkg/releases/tag/v0.3.6
