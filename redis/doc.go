/*
Package redis provides a Redis client wrapper (github.com/redis/go-redis/v9) for standalone and cluster modes, configured from config package.

Role in architecture:
  - Infrastructure: connects to Redis, exposes Cmdable; used by rate limiter and application cache/session code.

Responsibilities:
  - Setup: create standard or cluster client from config; no-op if Redis.Enabled is false.
  - Available: TCP reachability check for host:port without using the Redis client (no connection logs).
  - IsAlive: ping check. GetRedis/GetRedisCluster/GetUniversalClient: access the client.
  - Close: close the active client and release connections; safe to call when not enabled.

Constraints:
  - Single client per process; no provider switching or multi-instance.
  - Cluster mode is determined by config at Setup; no runtime switch.
  - common.go provides helpers (Get, Set, HGet, etc.) that call GetUniversalClient().

This package must NOT:
  - Contain business logic; only connection and command delegation.
*/
package redis
