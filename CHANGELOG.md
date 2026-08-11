# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[Unreleased]: https://github.com/turahe/pkg/compare/v0.5.4...HEAD
[0.5.4]: https://github.com/turahe/pkg/compare/v0.5.3...v0.5.4
[0.3.7]: https://github.com/turahe/pkg/releases/tag/v0.3.7
[0.3.6]: https://github.com/turahe/pkg/releases/tag/v0.3.6
