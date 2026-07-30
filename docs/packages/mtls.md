# mtls

Mutual TLS helpers for backend HTTPS servers and gateway outbound HTTP clients.

**Import:** `github.com/turahe/pkg/mtls`

## Role

Infrastructure TLS wiring from `config.MTLS`. When `MTLS_ENABLED=false`, configure helpers are no-ops / plain listen.

## Config

```go
cfg := mtls.LoadConfig() // alias of config.MTLSConfiguration
```

Env: `MTLS_ENABLED`, CA / server / client cert paths, `MTLS_SKIP_PATHS` (used by middleware).

## Server

```go
tlsCfg, err := mtls.ServerTLSConfig(cfg)
// or:
mtls.ConfigureServer(srv, cfg)           // attaches TLS when enabled
mtls.ListenAndServe(":8443", handler, cfg)
mtls.RunGin(engine, addr, cfg)           // Gin convenience
```

Pair with `middlewares.MTLSMiddleware()` so application-layer checks match TLS client auth (sets `mtls_client_cn` in context; skips configured paths).

## Client / transport

```go
tlsCfg, err := mtls.ClientTLSConfig(cfg)
tr, err := mtls.NewTransport(cfg) // *http.Transport with client cert
httpClient := &http.Client{Transport: tr}
```

`CloneTLSConfig` for safe copies.

## Constraints

- Enable server TLS **and** middleware together when requiring client certs end-to-end.
- Cert paths must be readable by the process (containers: mount secrets under `/etc/mtls`).

## See also

- [Environment — mTLS](../environment.md#mtls)
- [middlewares — MTLS](middlewares.md)
