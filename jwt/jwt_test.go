package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/turahe/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func initTestJWT(t *testing.T) {
	config.Config = &config.Configuration{
		Server: config.ServerConfiguration{
			Secret:             "test-secret-key-for-jwt-tests",
			AccessTokenExpiry:  1,
			RefreshTokenExpiry: 7,
		},
	}
	Init()
}

func TestGenerateTokenWithExpiry_and_ValidateToken(t *testing.T) {
	initTestJWT(t)

	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	token, err := GenerateTokenWithExpiry(id, time.Hour)
	if err != nil {
		t.Fatalf("GenerateTokenWithExpiry: %v", err)
	}
	if token == "" {
		t.Fatal("token must not be empty")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UUID != id.String() {
		t.Errorf("UUID = %q, want %q", claims.UUID, id.String())
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	initTestJWT(t)

	_, err := ValidateToken("invalid-token")
	if err == nil {
		t.Error("ValidateToken should fail for invalid token")
	}
}

func TestValidateToken_Empty(t *testing.T) {
	initTestJWT(t)

	_, err := ValidateToken("")
	if err == nil {
		t.Error("ValidateToken should fail for empty token")
	}
}

func TestGetCurrentUserUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name      string
		setup     func(c *gin.Context)
		wantUUID  uuid.UUID
		wantFound bool
	}{
		{
			name:      "key absent",
			setup:     func(c *gin.Context) {},
			wantUUID:  uuid.Nil,
			wantFound: false,
		},
		{
			name: "valid UUID string",
			setup: func(c *gin.Context) {
				c.Set("user_id", validUUID.String())
			},
			wantUUID:  validUUID,
			wantFound: true,
		},
		{
			name: "uuid.UUID value",
			setup: func(c *gin.Context) {
				c.Set("user_id", validUUID)
			},
			wantUUID:  validUUID,
			wantFound: true,
		},
		{
			name: "malformed UUID string",
			setup: func(c *gin.Context) {
				c.Set("user_id", "not-a-uuid")
			},
			wantUUID:  uuid.Nil,
			wantFound: false,
		},
		{
			name: "empty string",
			setup: func(c *gin.Context) {
				c.Set("user_id", "")
			},
			wantUUID:  uuid.Nil,
			wantFound: false,
		},
		{
			name: "wrong type (int)",
			setup: func(c *gin.Context) {
				c.Set("user_id", 12345)
			},
			wantUUID:  uuid.Nil,
			wantFound: false,
		},
		{
			name: "wrong type (bool)",
			setup: func(c *gin.Context) {
				c.Set("user_id", true)
			},
			wantUUID:  uuid.Nil,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
			tt.setup(c)

			got, found := GetCurrentUserUUID(c)

			require.Equal(t, tt.wantFound, found)
			assert.Equal(t, tt.wantUUID, got)
		})
	}
}

func TestGetCurrentUserUUID_NilUUIDString(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Set("user_id", uuid.Nil.String()) // "00000000-0000-0000-0000-000000000000"

	got, found := GetCurrentUserUUID(c)

	// uuid.Nil is a valid UUID string — it parses successfully and returns true.
	assert.True(t, found)
	assert.Equal(t, uuid.Nil, got)
}

func TestJWT_ComparePassword(t *testing.T) {
	plain := "password123"
	hashed := "$2a$04$XQ6I5VvW8n/5QZ5Z5Z5Z5uK9qH8qH8qH8qH8qH8qH8qH8qH8qH8qH"
	if ComparePassword(hashed, plain) {
		t.Error("ComparePassword should return false for invalid hash")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}
	if !ComparePassword(string(hash), plain) {
		t.Error("ComparePassword should return true for valid hash")
	}
}
