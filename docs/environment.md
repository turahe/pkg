# Environment Variables

Configuration is loaded by [`config`](packages/config.md) from the process environment, or from a `.env` file via `config.Setup("")` ([godotenv](https://github.com/joho/godotenv)). Real environment variables override `.env`.

Copy [`.env.example`](../.env.example) as a starting point. Tables below match the root [README](../README.md#environment-variables); this page is the docs-index entry point.

## Server & JWT

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Listen port |
| `SERVER_MODE` | `debug` | `debug` · `release` · `test` |
| `SERVER_SECRET` | — | HS256 signing secret |
| `SERVER_ACCESS_TOKEN_EXPIRY` | `1` | Access token hours |
| `SERVER_REFRESH_TOKEN_EXPIRY` | `7` | Refresh token days |
| `SERVER_SESSION_EXPIRY` | `24` | Session hours |
| `SERVER_SESSION_COOKIE_NAME` | `admin_session` | Cookie name |
| `SERVER_SESSION_SECURE` | `false` | Secure cookie flag |
| `SERVER_SESSION_HTTP_ONLY` | `true` | HttpOnly cookie |
| `SERVER_SESSION_SAME_SITE` | `lax` | `strict` · `lax` · `none` |
| `SERVER_TIMEZONE` | `UTC` | IANA timezone |
| `APP_ENV` | — | Fallback for Sentry / OTel environment |
| `JWT_SIGNING_ALGORITHM` | `RS256` | `HS256` · `RS256` · `ES256` |
| `JWT_PRIVATE_KEY` | — | Path or inline PEM (RS256/ES256) |
| `JWT_PUBLIC_KEY` | — | Path or inline PEM (RS256/ES256) |
| `JWT_ISSUER` | — | Optional `iss` |
| `JWT_AUDIENCE` | — | Optional `aud` (comma-separated) |
| `JWT_KEY_ID` | — | Optional `kid` |

Embed PEM at build time via `config.Server.JWTPrivateKeyPEM` / `JWTPublicKeyPEM` instead of env paths.

## CORS

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_GLOBAL` | `true` | Allow all origins when true |
| `CORS_FRONTEND` | — | Frontend origin when not global |
| `CORS_IPS` | — | Extra allowed origins (comma-separated) |

## Database

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_DRIVER` | `mysql` | `mysql` · `postgres` · `sqlite` · `sqlserver` · `cloudsql-mysql` · `cloudsql-postgres` |
| `DATABASE_HOST` | `127.0.0.1` | |
| `DATABASE_PORT` | `3306` | |
| `DATABASE_USERNAME` | — | Required |
| `DATABASE_PASSWORD` | — | Required |
| `DATABASE_DBNAME` | — | Required |
| `DATABASE_SSLMODE` | `false` | |
| `DATABASE_LOGMODE` | `false` | GORM query logging |
| `DATABASE_TIMEZONE` | — | Session TZ; empty → `SERVER_TIMEZONE` |
| `DATABASE_MAX_IDLE_CONNS` | `0` (→ 10) | Pool idle |
| `DATABASE_MAX_OPEN_CONNS` | `0` (→ 30) | Pool open |
| `DATABASE_CONN_MAX_LIFETIME` | `0` (→ 30) | Minutes |
| `DATABASE_CLOUD_SQL_INSTANCE` | — | `project:region:instance` |

Secondary DB: same keys with `_SITE` suffix. Leave `DATABASE_DBNAME_SITE` empty to disable.

Cloud SQL **IAM** and **Private IP** are set in code via `database.Options` (`WithIAM`, `WithPrivateIP`), not env.

## Redis

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_ENABLED` | `false` | |
| `REDIS_HOST` | `127.0.0.1` | |
| `REDIS_PORT` | `6379` | |
| `REDIS_PASSWORD` | — | |
| `REDIS_DB` | `0` | Ignored in cluster mode |
| `REDIS_CLUSTER_MODE` | `false` | |
| `REDIS_CLUSTER_NODES` | — | Comma-separated `host:port` |
| `REDIS_POOL_SIZE` | `0` | |
| `REDIS_MIN_IDLE_CONNS` | `0` | |
| `REDIS_READ_TIMEOUT_SEC` | `0` | |
| `REDIS_WRITE_TIMEOUT_SEC` | `0` | |

## Rate limiter

Requires Redis. See [middlewares](packages/middlewares.md).

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMITER_ENABLED` | `false` | |
| `RATE_LIMITER_REQUESTS` | `100` | Per window |
| `RATE_LIMITER_WINDOW` | `60` | Seconds |
| `RATE_LIMITER_KEY_BY` | `ip` | `ip` or `user` |
| `RATE_LIMITER_SKIP_PATHS` | — | e.g. `/health,/metrics` |

## GCS

| Variable | Default | Description |
|----------|---------|-------------|
| `GCS_ENABLED` | `false` | |
| `GCS_BUCKET_NAME` | — | |
| `GCS_CREDENTIALS_FILE` | — | SA JSON path; omit for ADC |

## OpenTelemetry

See [otelx](packages/otelx.md).

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_TRACES_EXPORTER` | `otlp` | `otlp` or `gcp` |
| `OTEL_SERVICE_NAME` | `app` | |
| `OTEL_ENVIRONMENT` | `APP_ENV` | |
| `OTEL_SERVICE_VERSION` | `SENTRY_RELEASE` | |
| `OTEL_TRACES_SAMPLER_ARG` | `1.0` | 0..1 |
| `OTEL_SHUTDOWN_TIMEOUT` | `5s` | |
| `OTEL_GORM_ENABLED` | `true` | Auto GORM plugin |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | — | Required for OTLP |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | — | Override |
| `OTEL_EXPORTER_OTLP_INSECURE` | `true` | |
| `OTEL_EXPORTER_OTLP_HEADERS` | — | `key=value,...` |
| `OTEL_GCP_PROJECT_ID` | `GOOGLE_CLOUD_PROJECT` | |
| `OTEL_GCP_PROPAGATOR` | `false` | Cloud Trace propagator |

## Sentry

See [sentryx](packages/sentryx.md). Empty `SENTRY_DSN` disables Sentry.

| Variable | Default | Description |
|----------|---------|-------------|
| `SENTRY_DSN` | — | |
| `SENTRY_ENVIRONMENT` | `APP_ENV` | |
| `SENTRY_RELEASE` | — | |
| `SENTRY_SERVER_NAME` | `app` | |
| `SENTRY_DEBUG` | `false` | |
| `SENTRY_ATTACH_STACKTRACE` | `true` | |
| `SENTRY_SAMPLE_RATE` | `1.0` | |
| `SENTRY_TRACES_SAMPLE_RATE` | `0.0` | |
| `SENTRY_FLUSH_TIMEOUT` | `2s` | |

## mTLS

See [mtls](packages/mtls.md).

| Variable | Default | Description |
|----------|---------|-------------|
| `MTLS_ENABLED` | `false` | |
| `MTLS_CA_CERT` | `/etc/mtls/ca.crt` | |
| `MTLS_SERVER_CERT` | `/etc/mtls/server.crt` | |
| `MTLS_SERVER_KEY` | `/etc/mtls/server.key` | |
| `MTLS_CLIENT_CERT` | `/etc/mtls/gateway.crt` | Outbound |
| `MTLS_CLIENT_KEY` | `/etc/mtls/gateway.key` | Outbound |
| `MTLS_SKIP_PATHS` | `/live,/ready,/metrics` | Bypass middleware |

## Other

| Variable | Description |
|----------|-------------|
| `GOOGLE_CLOUD_PROJECT` | Trace linking in logs; GCP OTel / ADC fallbacks |
