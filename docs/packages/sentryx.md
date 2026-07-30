# sentryx

Thin wrapper around [sentry-go](https://github.com/getsentry/sentry-go) configured from the `config` package.

**Import:** `github.com/turahe/pkg/sentryx`

## Role

Infrastructure error reporting. Empty `SENTRY_DSN` is the explicit off switch.

## API

```go
cfg := sentryx.LoadConfig()
enabled := sentryx.Init(cfg)
defer sentryx.Flush(cfg.FlushTimeout) // safe if never initialized
```

- `Init` returns `bool` (enabled), not `error` — invalid DSN is logged and treated as off so the service keeps running.
- Default server name: `app` when `SENTRY_SERVER_NAME` is empty.
- Sample rates clamped to `[0, 1]`.

## Constraints

- Must not block startup on Sentry failure.
- Prefer `defer Flush` immediately after Init so SIGTERM does not drop the last events.

## See also

- [Environment — Sentry](../environment.md#sentry)
- [Getting Started](../getting-started.md)
