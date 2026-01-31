package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/redis"
	"github.com/turahe/pkg/response"
)

// setupTestConfig sets up test configuration for rate limiter
func setupTestConfig(t *testing.T, rateLimiterEnabled, redisEnabled bool) {
	config.Config = &config.Configuration{
		RateLimiter: config.RateLimiterConfiguration{
			Enabled:   rateLimiterEnabled,
			Requests:  5,
			Window:    60,
			KeyBy:     "ip",
			SkipPaths: "/health,/metrics",
		},
		Redis: config.RedisConfiguration{
			Enabled: redisEnabled,
			Host:    "127.0.0.1",
			Port:    "6379",
			DB:      0,
		},
	}

	if redisEnabled {
		err := redis.Setup()
		if err != nil {
			t.Skipf("Redis not available for rate limiter tests: %v", err)
		}
	}
}

// cleanupTestRedis cleans up test keys from Redis (works with standard and cluster mode)
func cleanupTestRedis(t *testing.T) {
	rdb := redis.GetUniversalClient()
	if rdb == nil {
		return
	}
	ctx := context.Background()
	keys, _ := rdb.Keys(ctx, "rate_limit:*").Result()
	for _, key := range keys {
		_ = rdb.Del(ctx, key).Err()
	}
}

func TestRateLimiter_Disabled(t *testing.T) {
	setupTestConfig(t, false, false)

	router := setupRouter()
	router.Use(RateLimiter())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_RedisDisabled(t *testing.T) {
	setupTestConfig(t, true, false)

	router := setupRouter()
	router.Use(RateLimiter())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should allow request when Redis is disabled
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_SkipPaths(t *testing.T) {
	setupTestConfig(t, true, true)
	defer cleanupTestRedis(t)

	router := setupRouter()
	router.Use(RateLimiter())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"metrics": "ok"})
	})
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test /health should skip rate limiting
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test /metrics should skip rate limiting
	req = httptest.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test /api/test should NOT skip (different path)
	req = httptest.NewRequest("GET", "/api/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_WithinLimit(t *testing.T) {
	setupTestConfig(t, true, true)
	defer cleanupTestRedis(t)

	router := setupRouter()
	router.Use(RateLimiter())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make requests within limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
		assert.Contains(t, w.Header().Get("X-RateLimit-Limit"), "5")
		remaining := w.Header().Get("X-RateLimit-Remaining")
		assert.NotEmpty(t, remaining)
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	}
}

func TestRateLimiter_ExceedsLimit(t *testing.T) {
	setupTestConfig(t, true, true)
	defer cleanupTestRedis(t)

	router := setupRouter()
	router.Use(RateLimiter())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make requests up to limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.2:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Header().Get("Retry-After"), "")
	assert.Contains(t, w.Header().Get("X-RateLimit-Limit"), "5")

	var resp response.CommonResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Message, "Rate limit exceeded")
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	setupTestConfig(t, true, true)
	defer cleanupTestRedis(t)

	router := setupRouter()
	router.Use(RateLimiter())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make 5 requests from IP 1
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.10:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// IP 1 should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.10:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// IP 2 should still be able to make requests
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.20:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_KeyByUser(t *testing.T) {
	setupTestConfig(t, true, true)
	defer cleanupTestRedis(t)

	config.Config.RateLimiter.KeyBy = "user"

	router := setupRouter()
	router.Use(RateLimiter())
	// Set admin_id in context (simulating auth middleware)
	router.GET("/test", func(c *gin.Context) {
		c.Set("admin_id", "user123")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make requests with user ID
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimiter_KeyByUser_FallbackToIP(t *testing.T) {
	setupTestConfig(t, true, true)
	defer cleanupTestRedis(t)

	config.Config.RateLimiter.KeyBy = "user"

	router := setupRouter()
	router.Use(RateLimiter())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request without admin_id should fallback to IP
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.30:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should use IP-based key when user ID is not available
}

func TestGetRateLimitKey_IP(t *testing.T) {
	router := setupRouter()
	router.GET("/test", func(c *gin.Context) {
		key := getRateLimitKey(c, "ip")
		c.JSON(http.StatusOK, gin.H{"key": key})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ip:192.168.1.100", resp["key"])
}

func TestGetRateLimitKey_User(t *testing.T) {
	router := setupRouter()
	router.GET("/test", func(c *gin.Context) {
		c.Set("admin_id", "user456")
		key := getRateLimitKey(c, "user")
		c.JSON(http.StatusOK, gin.H{"key": key})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "user:user456", resp["key"])
}

func TestGetRateLimitKey_User_FallbackToIP(t *testing.T) {
	router := setupRouter()
	router.GET("/test", func(c *gin.Context) {
		// No admin_id set
		key := getRateLimitKey(c, "user")
		c.JSON(http.StatusOK, gin.H{"key": key})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.200:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ip:192.168.1.200", resp["key"])
}

func TestGetRateLimitKey_Default(t *testing.T) {
	router := setupRouter()
	router.GET("/test", func(c *gin.Context) {
		key := getRateLimitKey(c, "unknown")
		c.JSON(http.StatusOK, gin.H{"key": key})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ip:10.0.0.1", resp["key"])
}

func TestParseSkipPaths_Empty(t *testing.T) {
	result := parseSkipPaths("")
	assert.Empty(t, result)
}

func TestParseSkipPaths_SinglePath(t *testing.T) {
	result := parseSkipPaths("/health")
	assert.Equal(t, []string{"/health"}, result)
}

func TestParseSkipPaths_MultiplePaths(t *testing.T) {
	result := parseSkipPaths("/health,/metrics,/status")
	assert.Equal(t, []string{"/health", "/metrics", "/status"}, result)
}

func TestParseSkipPaths_WithSpaces(t *testing.T) {
	result := parseSkipPaths("/health, /metrics , /status")
	assert.Equal(t, []string{"/health", "/metrics", "/status"}, result)
}

func TestParseSkipPaths_EmptyPaths(t *testing.T) {
	result := parseSkipPaths("/health,,/metrics")
	assert.Equal(t, []string{"/health", "/metrics"}, result)
}

func TestShouldSkipPath_ExactMatch(t *testing.T) {
	skipPaths := []string{"/health", "/metrics"}
	assert.True(t, shouldSkipPath("/health", skipPaths))
	assert.True(t, shouldSkipPath("/metrics", skipPaths))
	assert.False(t, shouldSkipPath("/api/test", skipPaths))
}

func TestShouldSkipPath_PrefixMatch(t *testing.T) {
	skipPaths := []string{"/health", "/metrics"}
	assert.True(t, shouldSkipPath("/health/check", skipPaths))
	assert.True(t, shouldSkipPath("/metrics/prometheus", skipPaths))
	assert.False(t, shouldSkipPath("/api/health", skipPaths))
}

func TestShouldSkipPath_EmptySkipPaths(t *testing.T) {
	skipPaths := []string{}
	assert.False(t, shouldSkipPath("/health", skipPaths))
	assert.False(t, shouldSkipPath("/api/test", skipPaths))
}

func TestMax(t *testing.T) {
	assert.Equal(t, 5, max(5, 3))
	assert.Equal(t, 5, max(3, 5))
	assert.Equal(t, 0, max(0, 0))
	assert.Equal(t, -1, max(-1, -5))
	assert.Equal(t, 100, max(100, 50))
}

func TestRateLimiter_Headers(t *testing.T) {
	setupTestConfig(t, true, true)
	defer cleanupTestRedis(t)

	router := setupRouter()
	router.Use(RateLimiter())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.40:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimiter_WindowExpiration(t *testing.T) {
	setupTestConfig(t, true, true)
	defer cleanupTestRedis(t)

	// Set a very short window for testing
	config.Config.RateLimiter.Window = 2 // 2 seconds
	config.Config.RateLimiter.Requests = 2

	router := setupRouter()
	router.Use(RateLimiter())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make 2 requests (limit)
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.50:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.50:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Third request should be rate limited
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.50:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Wait for window to expire
	time.Sleep(3 * time.Second)

	// Should be able to make requests again
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.50:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
