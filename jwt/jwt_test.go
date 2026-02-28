package jwt

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/turahe/pkg/config"
	"golang.org/x/crypto/bcrypt"
)

func testManagerHS256(t *testing.T) *Manager {
	t.Helper()
	cfg := &config.Configuration{
		Server: config.ServerConfiguration{
			JWTSigningAlgorithm: "HS256",
			Secret:              "test-secret-key-for-jwt-tests",
			AccessTokenExpiry:    1,
			RefreshTokenExpiry:   7,
		},
	}
	m, err := NewManager(context.Background(), cfg)
	require.NoError(t, err)
	return m
}

func TestGenerateTokenWithExpiry_and_ValidateToken(t *testing.T) {
	m := testManagerHS256(t)

	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	token, err := m.GenerateTokenWithExpiry(id, time.Hour)
	if err != nil {
		t.Fatalf("GenerateTokenWithExpiry: %v", err)
	}
	if token == "" {
		t.Fatal("token must not be empty")
	}

	claims, err := m.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UUID != id.String() {
		t.Errorf("UUID = %q, want %q", claims.UUID, id.String())
	}
	assert.Equal(t, TokenTypeAccess, claims.TokenType)
}

func TestGenerateImpersonationToken_Claims(t *testing.T) {
	m := testManagerHS256(t)

	adminID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	targetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	token, err := m.GenerateImpersonationToken(adminID, "admin", targetID, 15*time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := m.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, targetID.String(), claims.UUID)
	assert.Equal(t, targetID.String(), claims.Subject)
	assert.True(t, claims.IsImpersonating)
	assert.Equal(t, adminID.String(), claims.ImpersonatorID)
	assert.Equal(t, "admin", claims.ImpersonatorRole)
	assert.Equal(t, adminID.String(), claims.OriginalSub)
	assert.Equal(t, TokenTypeImpersonation, claims.TokenType)

	require.NotNil(t, claims.ExpiresAt)
	require.NotNil(t, claims.IssuedAt)
	ttl := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	assert.LessOrEqual(t, ttl, 30*time.Minute+5*time.Second)
}

func TestValidateToken_Invalid(t *testing.T) {
	m := testManagerHS256(t)

	_, err := m.ValidateToken("invalid-token")
	if err == nil {
		t.Error("ValidateToken should fail for invalid token")
	}
}

func TestValidateToken_Empty(t *testing.T) {
	m := testManagerHS256(t)

	_, err := m.ValidateToken("")
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

// setupRS256Config creates temp RSA key files and returns a Manager for RS256.
func setupRS256Config(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	privPath := filepath.Join(dir, "priv.pem")
	pubPath := filepath.Join(dir, "pub.pem")

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pub := &priv.PublicKey

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	require.NoError(t, os.WriteFile(privPath, privPEM, 0600))

	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	require.NoError(t, os.WriteFile(pubPath, pubPEM, 0644))

	cfg := &config.Configuration{
		Server: config.ServerConfiguration{
			JWTSigningAlgorithm: "RS256",
			JWTPrivateKeyPath:   privPath,
			JWTPublicKeyPath:    pubPath,
			AccessTokenExpiry:   1,
			RefreshTokenExpiry:  7,
		},
	}
	m, err := NewManager(context.Background(), cfg)
	require.NoError(t, err)
	return m
}

func TestRS256_GenerateAndValidate(t *testing.T) {
	m := setupRS256Config(t)

	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := m.GenerateTokenWithExpiry(id, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := m.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, id.String(), claims.UUID)
}

// setupES256Config creates temp ECDSA P-256 key files and returns a Manager for ES256.
func setupES256Config(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	privPath := filepath.Join(dir, "ec-priv.pem")
	pubPath := filepath.Join(dir, "ec-pub.pem")

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	privDER, err := x509.MarshalECPrivateKey(priv)
	require.NoError(t, err)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})
	require.NoError(t, os.WriteFile(privPath, privPEM, 0600))

	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	require.NoError(t, os.WriteFile(pubPath, pubPEM, 0644))

	cfg := &config.Configuration{
		Server: config.ServerConfiguration{
			JWTSigningAlgorithm: "ES256",
			JWTPrivateKeyPath:   privPath,
			JWTPublicKeyPath:    pubPath,
			AccessTokenExpiry:   1,
			RefreshTokenExpiry:  7,
		},
	}
	m, err := NewManager(context.Background(), cfg)
	require.NoError(t, err)
	return m
}

func TestES256_GenerateAndValidate(t *testing.T) {
	m := setupES256Config(t)

	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	token, err := m.GenerateTokenWithExpiry(id, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := m.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, id.String(), claims.UUID)
}

func TestNewManager_NilConfig(t *testing.T) {
	_, err := NewManager(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config is required")
}

func TestNewManager_InvalidAlgorithm(t *testing.T) {
	cfg := &config.Configuration{
		Server: config.ServerConfiguration{
			JWTSigningAlgorithm: "HS512",
			Secret:              "secret",
		},
	}
	_, err := NewManager(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HS256, RS256, or ES256")
}

func TestNewManager_HS256_NoSecret(t *testing.T) {
	cfg := &config.Configuration{
		Server: config.ServerConfiguration{
			JWTSigningAlgorithm: "HS256",
			Secret:              "",
		},
	}
	_, err := NewManager(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secret is not configured")
}

func TestManager_IssuerAudienceAndKid(t *testing.T) {
	cfg := &config.Configuration{
		Server: config.ServerConfiguration{
			JWTSigningAlgorithm: "HS256",
			Secret:              "test-secret",
			AccessTokenExpiry:    1,
			RefreshTokenExpiry:   7,
			JWTIssuer:            "https://api.example.com",
			JWTAudience:          "api.example.com,web.example.com",
			JWTKeyID:             "key-2024",
		},
	}
	m, err := NewManager(context.Background(), cfg)
	require.NoError(t, err)

	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := m.GenerateToken(id)
	require.NoError(t, err)

	claims, err := m.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", claims.Issuer)
	assert.ElementsMatch(t, []string{"api.example.com", "web.example.com"}, claims.Audience)
	assert.Equal(t, TokenTypeAccess, claims.TokenType)
	// kid is in header; validation passes so kid matched
}

func TestNewSigner_NewVerifier_Split(t *testing.T) {
	cfg := &config.Configuration{
		Server: config.ServerConfiguration{
			JWTSigningAlgorithm: "HS256",
			Secret:              "split-test-secret",
			AccessTokenExpiry:   1,
			RefreshTokenExpiry:  7,
		},
	}
	signer, err := NewSigner(context.Background(), cfg)
	require.NoError(t, err)
	verifier, err := NewVerifier(context.Background(), cfg)
	require.NoError(t, err)

	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := signer.GenerateToken(id)
	require.NoError(t, err)

	claims, err := verifier.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, id.String(), claims.UUID)
	assert.Equal(t, TokenTypeAccess, claims.TokenType)
}

func TestDefaultAlgorithm_RS256(t *testing.T) {
	// Empty algorithm defaults to RS256, so secret-only config fails (RS256 needs key paths).
	cfg := &config.Configuration{
		Server: config.ServerConfiguration{
			JWTSigningAlgorithm: "",
			Secret:              "some-secret",
		},
	}
	_, err := NewManager(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_PRIVATE_KEY_PATH")
}
