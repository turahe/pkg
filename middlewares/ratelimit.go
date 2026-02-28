package middlewares

import (
	"fmt"
	"strings"
	"time"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/redis"
	"github.com/turahe/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	defaultWindowSec   = 60
	rateLimitKeyPrefix = "rate_limit:"
)

// Sliding-window rate limit: one Lua script, one round-trip. Uses a ZSET keyed by timestamp;
// removes entries older than (now - window), then counts and optionally adds current request.
// Returns: {current_count, ttl_sec} when allowed, or {-1, ttl_sec} when limit exceeded (request not added).
const slidingWindowScript = `
local key = KEYS[1]
local now_sec = tonumber(ARGV[1])
local window_sec = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local unique_id = tostring(ARGV[4])
local oldest = now_sec - window_sec
redis.call('ZREMRANGEBYSCORE', key, '-inf', oldest)
local count = redis.call('ZCARD', key)
if count >= limit then
  local ttl = redis.call('TTL', key)
  if ttl < 0 then ttl = 0 end
  return {-1, ttl}
end
redis.call('ZADD', key, now_sec, unique_id)
redis.call('EXPIRE', key, window_sec)
count = redis.call('ZCARD', key)
local ttl = redis.call('TTL', key)
if ttl < 0 then ttl = 0 end
return {count, ttl}
`

// RateLimiter returns a Gin middleware that enforces a sliding-window rate limit using Redis.
// Key is per IP or per user (from "admin_id" in context when KeyBy is "user").
// SkipPaths (comma-separated) are not counted. On exceed returns 429 with Retry-After and X-RateLimit-* headers.
// If config.RateLimiter.Enabled or config.Redis.Enabled is false, returns a no-op middleware. On Redis
// error allows the request (fail open).
func RateLimiter() gin.HandlerFunc {
	conf := config.GetConfig()

	if !conf.RateLimiter.Enabled {
		return func(ctx *gin.Context) { ctx.Next() }
	}
	if !conf.Redis.Enabled {
		return func(ctx *gin.Context) { ctx.Next() }
	}

	requests := conf.RateLimiter.Requests
	keyBy := strings.TrimSpace(conf.RateLimiter.KeyBy)
	if keyBy == "" {
		keyBy = "ip"
	}
	skipPaths := parseSkipPaths(conf.RateLimiter.SkipPaths)
	windowSec := conf.RateLimiter.Window
	if windowSec <= 0 {
		windowSec = defaultWindowSec
	}

	rdb := redis.GetUniversalClient()

	return func(ctx *gin.Context) {
		if shouldSkipPath(ctx.Request.URL.Path, skipPaths) {
			ctx.Next()
			return
		}

		key := getRateLimitKey(ctx, keyBy)
		if key == "" {
			ctx.Next()
			return
		}

		redisKey := rateLimitKeyPrefix + key
		now := time.Now()
		nowSec := now.Unix()
		reqCtx := ctx.Request.Context()

		result, err := rdb.Eval(reqCtx, slidingWindowScript, []string{redisKey}, nowSec, windowSec, requests, makeUniqueRequestID(now)).Result()
		if err != nil {
			ctx.Next()
			return
		}

		arr, _ := result.([]interface{})
		if len(arr) < 2 {
			ctx.Next()
			return
		}
		current, _ := toInt64(arr[0])
		ttlSec, _ := toInt64(arr[1])
		if ttlSec < 0 {
			ttlSec = 0
		}

		ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
		ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, requests-int(current))))
		ctx.Header("X-RateLimit-Reset", fmt.Sprintf("%d", now.Unix()+ttlSec))

		if current < 0 || current > int64(requests) {
			ctx.Header("Retry-After", fmt.Sprintf("%d", ttlSec))
			response.FailWithDetailed(
				ctx,
				429,
				response.ServiceCodeCommon,
				response.CaseCodeRateLimitExceeded,
				nil,
				fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %d seconds.", requests, windowSec),
			)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// makeUniqueRequestID returns a unique string for this request for use as ZSET member (avoids overwrites in same second).
func makeUniqueRequestID(now time.Time) string {
	return fmt.Sprintf("%d", now.UnixNano())
}

func toInt64(v interface{}) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case float64:
		return int64(x), true
	default:
		return 0, false
	}
}

// getRateLimitKey determines the key for rate limiting based on the strategy
func getRateLimitKey(ctx *gin.Context, keyBy string) string {
	switch keyBy {
	case "user":
		if adminID, exists := ctx.Get("admin_id"); exists {
			if id, ok := adminID.(string); ok && id != "" {
				return fmt.Sprintf("user:%s", id)
			}
		}
		return fmt.Sprintf("ip:%s", ctx.ClientIP())
	case "ip":
		fallthrough
	default:
		return fmt.Sprintf("ip:%s", ctx.ClientIP())
	}
}

// parseSkipPaths parses comma-separated paths into a slice (trimmed, non-empty only).
func parseSkipPaths(paths string) []string {
	if paths == "" {
		return nil
	}
	parts := strings.Split(paths, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// shouldSkipPath returns true if path has any of the skip paths as prefix.
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
