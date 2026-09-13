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

// gRPC (after Init)
srv := grpc.NewServer(sentryx.ServerOptions()...)
conn, err := grpc.NewClient(target, append(sentryx.DialOptions(), /* credentials... */)...)
```

- `Init` returns `bool` (enabled), not `error` — invalid DSN is logged and treated as off so the service keeps running.
- Default server name: `app` when `SENTRY_SERVER_NAME` is empty.
- Sample rates clamped to `[0, 1]`.

## gRPC

`ServerOptions` / `DialOptions` wrap [sentry-go/grpc](https://pkg.go.dev/github.com/getsentry/sentry-go/grpc) unary and stream interceptors. Server defaults set `Repanic: true`.

## Constraints

- Must not block startup on Sentry failure.
- Prefer `defer Flush` immediately after Init so SIGTERM does not drop the last events.

## See also

- [Environment — Sentry](../environment.md#sentry)
- [Getting Started](../getting-started.md)
- [otelx](otelx.md) (gRPC OTel stats handlers)
