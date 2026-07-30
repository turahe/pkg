# Architecture

This module follows **clean architecture** boundaries: domain and use-cases stay free of infrastructure; HTTP and persistence are adapters.

```
┌─────────────────────────────────────────────────────────┐
│  HTTP (Gin)                                             │
│  middlewares → handler → response                       │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│  Application (usecase)                                  │
│  depends on domain/port interfaces only                 │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│  Domain                                                 │
│  domain (errors) · domain/port (interfaces) · types     │
└──────────────────────────┬──────────────────────────────┘
                           │ implemented by
┌──────────────────────────▼──────────────────────────────┐
│  Infrastructure / Adapters                              │
│  config · database · redis · jwt · crypto · gcs         │
│  logger · otelx · sentryx · mtls · repositories         │
└─────────────────────────────────────────────────────────┘
```

## Layers

| Layer | Packages | Rule |
|-------|----------|------|
| **Domain** | `domain`, `domain/port` | No imports of database, Redis, config, or HTTP |
| **Application** | `usecase` | Depends on ports only; no Gin / GORM / Redis |
| **Adapters** | `handler`, `middlewares`, `response`, `repositories` | Translate HTTP/DB into domain/use-case calls |
| **Infrastructure** | `config`, `database`, `redis`, `jwt`, `crypto`, `gcs`, `logger`, `otelx`, `sentryx`, `mtls` | Tech-specific wiring |
| **Shared** | `types`, `util` | No business rules; usable by any layer |

## Dependency rules

1. **Dependencies point inward** — outer layers may import inner ones; never the reverse.
2. **Ports over concretions** — use cases call `domain/port` interfaces; repositories implement them in your app.
3. **Config is infrastructure** — load once at startup via `config.Setup` / `GetConfig`; do not import config from domain.
4. **No business logic in middleware/handlers** — validate, authorize, map errors; orchestrate in use cases.
5. **Prefer DI over globals** — `database.New(...)` and `jwt.NewManager(...)` over legacy `database.Setup()` / package singletons (compat APIs remain for migration).

## Typical request path

1. Middleware: recovery → trace IDs → log → metrics → timeout → CORS → optional mTLS / auth / rate limit
2. Handler: bind + validate → call use case → map domain errors via `BaseHandler.HandleServiceError`
3. Use case: call port interfaces with `context.Context`
4. Repository: GORM query with context; map `ErrRecordNotFound` to `(notFound, nil)` or domain `ErrNotFound`
5. Response: `response.Ok` / `Fail` / pagination helpers with composite codes

## Observability stack

| Concern | Package |
|---------|---------|
| Logs | `logger` (+ `LoggerMiddleware`) |
| Traces | `otelx` (+ `TraceMiddleware` / `CloudTraceMiddleware`) |
| Errors | `sentryx` |
| Metrics | `middlewares.Metrics` (Prometheus) |
| Health | `database.Health`, Redis `IsAlive`, `/live` + `/ready` |

## What packages must not do

Each package's `doc.go` lists explicit **must not** rules. In short:

- Domain packages must not import infrastructure
- Use cases must not import handlers or drivers
- Middleware / response / crypto / util must not contain business rules
- Infrastructure packages must not define domain entities or use-case flows
