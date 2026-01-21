package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/turahe/pkg/config"
)

func TestCORS_Global(t *testing.T) {
	// Save original config
	originalConfig := config.Config
	defer func() {
		config.Config = originalConfig
	}()

	// Set test config with global CORS enabled
	config.Config = &config.Configuration{
		Cors: config.CorsConfiguration{
			Global: true,
			Ips:    "http://example.com",
		},
	}

	router := setupRouter()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "POST, OPTIONS, GET, PUT, PATCH, DELETE", w.Header().Get("Access-Control-Allow-Methods"))
}

func TestCORS_SpecificIPs(t *testing.T) {
	// Save original config
	originalConfig := config.Config
	defer func() {
		config.Config = originalConfig
	}()

	// Set test config with specific IPs
	config.Config = &config.Configuration{
		Cors: config.CorsConfiguration{
			Global: false,
			Ips:    "http://example.com,http://test.com",
		},
	}

	router := setupRouter()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://example.com,http://test.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_OPTIONS(t *testing.T) {
	// Save original config
	originalConfig := config.Config
	defer func() {
		config.Config = originalConfig
	}()

	config.Config = &config.Configuration{
		Cors: config.CorsConfiguration{
			Global: true,
			Ips:    "",
		},
	}

	router := setupRouter()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test OPTIONS request (preflight)
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_Headers(t *testing.T) {
	// Save original config
	originalConfig := config.Config
	defer func() {
		config.Config = originalConfig
	}()

	config.Config = &config.Configuration{
		Cors: config.CorsConfiguration{
			Global: true,
			Ips:    "",
		},
	}

	router := setupRouter()
	router.Use(CORS())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "POST, OPTIONS, GET, PUT, PATCH, DELETE", w.Header().Get("Access-Control-Allow-Methods"))
}
