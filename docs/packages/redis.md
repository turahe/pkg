# redis

Wrapper around [go-redis](https://github.com/redis/go-redis) for standalone and cluster clients, plus helpers for strings, hashes, lists, sets, locks, pub/sub, and pipelines.

**Import:** `github.com/turahe/pkg/redis`

## Role

Infrastructure cache / coordination client. No business logic.

## Lifecycle

```go
if err := redis.Setup(); err != nil {
    log.Fatal(err)
}
defer redis.Close()

if !redis.Available() { /* disabled or not configured */ }
ok := redis.IsAlive(ctx)
client := redis.GetUniversalClient() // works for both modes
```

- `GetRedis()` / `GetRedisCluster()` panic if the wrong mode is configured or Setup was not called.
Prefer `GetUniversalClient()` when mode-agnostic.

## Helpers (selection)

| Area | Functions |
|------|-----------|
| KV | `Get`, `Set`, `Delete`, `MGet`, `MSet` |
| Hash | `HGet`, `HGetAll`, `HSet`, `HSetMap` |
| List | `LPush`, `RPop`, `LRange` |
| Set | `SAdd`, `SMembers`, `SRem` |
| Lock | `AcquireLock`, `ExtendLock`, `ReleaseLock` |
| Pipeline | `Pipeline`, `PipelineSet` |
| Pub/Sub | `PublishMessage`, `SubscribeToChannel` |
| Misc | `ScanKeys`, `Save`, `BGSave` |

## Constraints

- Single client instance; no runtime mode switch after Setup.
- Rate limiting middleware requires Redis enabled.
- Must not contain use-case logic.

## See also

- [Environment — Redis](../environment.md#redis)
- [middlewares — Rate limiter](middlewares.md)
