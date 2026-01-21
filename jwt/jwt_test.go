package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/turahe/pkg/config"
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
