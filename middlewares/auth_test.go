package middlewares

import (
	"context"
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

func initTestJWT(t *testing.T) *jwt.Manager {
	t.Helper()
	config.Config = &config.Configuration{
		Server: config.ServerConfiguration{
			JWTSigningAlgorithm: "HS256",
			Secret:              "test-secret-key-for-auth-middleware-tests",
			AccessTokenExpiry:    1,
			RefreshTokenExpiry:   7,
		},
	}
	manager, err := jwt.NewManager(context.Background(), config.Config)
	require.NoError(t, err)
	return manager
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	manager := initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware(manager))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager.GenerateToken(userID)
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
	manager := initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware(manager))
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
	manager := initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware(manager))
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
	manager := initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware(manager))
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
	manager := initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware(manager))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager.GenerateTokenWithExpiry(userID, -time.Hour)
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
	manager := initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware(manager))
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
	manager := initTestJWT(t)

	router := setupRouter()
	router.Use(AuthMiddleware(manager))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found"})
			return
		}
		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id is not a string"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userIDStr})
	})

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager.GenerateToken(userID)
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

func TestAuthMiddleware_ImpersonationContext(t *testing.T) {
	manager := initTestJWT(t)

	adminID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	targetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	router := setupRouter()
	router.Use(AuthMiddleware(manager))
	router.GET("/impersonation", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		originalUserID, _ := c.Get("original_user_id")
		isImpersonating, _ := c.Get("is_impersonating")
		impersonatorID, _ := c.Get("impersonator_id")
		impersonatorRole, _ := c.Get("impersonator_role")

		c.JSON(http.StatusOK, gin.H{
			"user_id":          userID,
			"original_user_id": originalUserID,
			"is_impersonating": isImpersonating,
			"impersonator_id":  impersonatorID,
			"impersonator_role": impersonatorRole,
		})
	})

	token, err := manager.GenerateImpersonationToken(adminID, "admin", targetID, 15*time.Minute)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/impersonation", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, targetID.String(), resp["user_id"])
	assert.Equal(t, adminID.String(), resp["original_user_id"])
	assert.Equal(t, true, resp["is_impersonating"])
	assert.Equal(t, adminID.String(), resp["impersonator_id"])
	assert.Equal(t, "admin", resp["impersonator_role"])
}

