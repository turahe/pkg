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
	window := time.Duration(conf.RateLimiter.Window) * time.Second
	keyBy := conf.RateLimiter.KeyBy
	skipPaths := parseSkipPaths(conf.RateLimiter.SkipPaths)

	return func(ctx *gin.Context) {
		// Skip rate limiting for specified paths
		if shouldSkipPath(ctx.Request.URL.Path, skipPaths) {
			ctx.Next()
			return
		}

		// Determine the key for rate limiting
		key := getRateLimitKey(ctx, keyBy)
		if key == "" {
			// If we can't determine a key, allow the request
			ctx.Next()
			return
		}

		redisKey := fmt.Sprintf("rate_limit:%s", key)
		rdb := redis.GetUniversalClient()

		// Use Redis INCR with expiration for sliding window rate limiting
		current, err := rdb.Incr(ctx.Request.Context(), redisKey).Result()
		if err != nil {
			// If Redis fails, allow the request (fail open)
			ctx.Next()
			return
		}

		// Set expiration on first request
		if current == 1 {
			rdb.Expire(ctx.Request.Context(), redisKey, window)
		}

		// Set rate limit headers
		ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
		ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, requests-int(current))))

		// Get TTL to calculate reset time
		ttl, _ := rdb.TTL(ctx.Request.Context(), redisKey).Result()
		resetTime := time.Now().Add(ttl).Unix()
		ctx.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

		// Check if rate limit exceeded
		if current > int64(requests) {
			ctx.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
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
