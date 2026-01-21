package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	router := setupRouter()
	router.Use(LoggerMiddleware())
	router.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("logs successful request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Middleware should not affect response
		assert.Contains(t, w.Body.String(), "success")
	})

	t.Run("logs request with query parameters", func(t *testing.T) {
		router := setupRouter()
		router.Use(LoggerMiddleware())
		router.GET("/search", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"query": ctx.Query("q")})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/search?q=test", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("logs error status codes", func(t *testing.T) {
		router := setupRouter()
		router.Use(LoggerMiddleware())
		router.GET("/notfound", func(ctx *gin.Context) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/notfound", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("logs server error status codes", func(t *testing.T) {
		router := setupRouter()
		router.Use(LoggerMiddleware())
		router.GET("/error", func(ctx *gin.Context) {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("logs different HTTP methods", func(t *testing.T) {
		router := setupRouter()
		router.Use(LoggerMiddleware())
		router.POST("/post", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"method": "POST"})
		})
		router.PUT("/put", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"method": "PUT"})
		})
		router.DELETE("/delete", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"method": "DELETE"})
		})

		tests := []struct {
			method string
			path   string
		}{
			{"POST", "/post"},
			{"PUT", "/put"},
			{"DELETE", "/delete"},
		}

		for _, tt := range tests {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("captures client IP", func(t *testing.T) {
		router := setupRouter()
		router.Use(LoggerMiddleware())
		router.GET("/ip", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"ip": ctx.ClientIP()})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ip", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
