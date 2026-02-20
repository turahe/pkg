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

// RateLimiter returns a rate limiting middleware
func RateLimiter() gin.HandlerFunc {
	conf := config.GetConfig()

	// If rate limiter is disabled, return a no-op middleware
	if !conf.RateLimiter.Enabled {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}

	// If Redis is not enabled, skip rate limiting
	if !conf.Redis.Enabled {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}

	requests := conf.RateLimiter.Requests
	keyBy := conf.RateLimiter.KeyBy
	skipPaths := parseSkipPaths(conf.RateLimiter.SkipPaths)
	windowSec := conf.RateLimiter.Window
	if windowSec <= 0 {
		windowSec = 60
	}

	// One Redis round-trip: INCR + conditional EXPIRE + TTL
	const rateLimitScript = `local c = redis.call('INCR', KEYS[1])
if c == 1 then redis.call('EXPIRE', KEYS[1], ARGV[1]) end
return {c, redis.call('TTL', KEYS[1])}`

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

		redisKey := fmt.Sprintf("rate_limit:%s", key)
		rdb := redis.GetUniversalClient()
		reqCtx := ctx.Request.Context()

		result, err := rdb.Eval(reqCtx, rateLimitScript, []string{redisKey}, windowSec).Result()
		if err != nil {
			ctx.Next()
			return
		}

		arr, _ := result.([]interface{})
		if len(arr) < 2 {
			ctx.Next()
			return
		}
		current, _ := arr[0].(int64)
		ttlSec := int64(0)
		if v, ok := arr[1].(int64); ok {
			ttlSec = v
		}
		if ttlSec < 0 {
			ttlSec = 0
		}
		ttl := time.Duration(ttlSec) * time.Second
		resetTime := time.Now().Add(ttl).Unix()

		ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
		ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, requests-int(current))))
		ctx.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

		if current > int64(requests) {
			ctx.Header("Retry-After", fmt.Sprintf("%d", ttlSec))
			response.FailWithDetailed(
				ctx,
				429,
				response.ServiceCodeCommon,
				response.CaseCodeRateLimitExceeded,
				nil,
				fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %d seconds.", requests, conf.RateLimiter.Window),
			)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// getRateLimitKey determines the key for rate limiting based on the strategy
func getRateLimitKey(ctx *gin.Context, keyBy string) string {
	switch keyBy {
	case "user":
		// Try to get user ID from context (set by auth middleware)
		if adminID, exists := ctx.Get("admin_id"); exists {
			if id, ok := adminID.(string); ok && id != "" {
				return fmt.Sprintf("user:%s", id)
			}
		}
		// Fallback to IP if user ID not available
		return fmt.Sprintf("ip:%s", ctx.ClientIP())
	case "ip":
		fallthrough
	default:
		return fmt.Sprintf("ip:%s", ctx.ClientIP())
	}
}

// parseSkipPaths parses comma-separated paths into a slice
func parseSkipPaths(paths string) []string {
	if paths == "" {
		return []string{}
	}

	parts := strings.Split(paths, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// shouldSkipPath checks if the given path should skip rate limiting
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
