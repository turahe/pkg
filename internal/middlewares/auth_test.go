package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/jwt"
	"github.com/turahe/pkg/response"
)

func initTestJWT(t *testing.T) {
	config.Config = &config.Configuration{
		Server: config.ServerConfiguration{
			Secret:             "test-secret-key-for-auth-middleware-tests",
			AccessTokenExpiry:  1,
			RefreshTokenExpiry: 7,
		},
	}
	jwt.Init()
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// Generate a valid token
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := jwt.GenerateToken(userID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), resp["user_id"])
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.CommonResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Authorization header is required", resp.Message)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test without "Bearer" prefix
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.CommonResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Invalid authorization header format", resp.Message)

	// Test with only "Bearer" (no token)
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test with multiple spaces
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token1 token2")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.CommonResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Message, "Invalid or expired token")
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Generate an expired token
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := jwt.GenerateTokenWithExpiry(userID, -time.Hour) // Expired 1 hour ago
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.CommonResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Message, "Invalid or expired token")
}

func TestAuthMiddleware_AbortsOnError(t *testing.T) {
	initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "should not reach here"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// Missing Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	// Should not contain the success message
	assert.NotContains(t, w.Body.String(), "should not reach here")
}

func TestAuthMiddleware_ContextValue(t *testing.T) {
	initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found"})
			return
		}
		// Verify it's a string
		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id is not a string"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userIDStr})
	})

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := jwt.GenerateToken(userID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), resp["user_id"])
}
