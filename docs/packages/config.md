# config

Infrastructure package that loads application configuration from environment variables (and optional `.env` via [godotenv](https://github.com/joho/godotenv)).

**Import:** `github.com/turahe/pkg/config`

## Role

Single source of truth for server, JWT, CORS, database (primary + site), Redis, GCS, rate limiter, timezone, OpenTelemetry, Sentry, and mTLS settings.

## API

```go
config.Setup(configPath string) error  // "" loads .env from cwd / env only
config.GetConfig() *Configuration      // lazy-builds from env if needed
config.SetConfig(cfg *Configuration)   // override for tests / manual wiring
```

Main type: `Configuration` with nested `ServerConfiguration`, `CorsConfiguration`, `DatabaseConfiguration`, `RedisConfiguration`, `GCSConfiguration`, `RateLimiterConfiguration`, `TimezoneConfiguration`, `OpenTelemetryConfiguration`, `SentryConfiguration`, `MTLSConfiguration`.

## Usage

```go
if err := config.Setup(""); err != nil {
    log.Fatal(err)
}
cfg := config.GetConfig()
```

For JWT keys embedded at build time:

```go
cfg := config.GetConfig()
cfg.Server.JWTPrivateKeyPEM = privatePEM
cfg.Server.JWTPublicKeyPEM = publicPEM
config.SetConfig(cfg)
```

## Constraints

- No YAML/JSON config files (only `.env` for variable loading).
- No external secret managers — set env (or embed) before `Setup` / `GetConfig`.
- Database validation runs for primary and (if `Dbname` set) site DB only.
- Must not depend on database, Redis, or HTTP packages.

## See also

- [Environment variables](../environment.md)
- [`.env.example`](../../.env.example)
- GoDoc: [pkg.go.dev/.../config](https://pkg.go.dev/github.com/turahe/pkg/config)
