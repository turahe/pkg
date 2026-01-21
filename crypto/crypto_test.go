package crypto

import (
	"testing"
)

func TestHashAndSalt(t *testing.T) {
	plain := []byte("password123")
	hash := HashAndSalt(plain)
	if hash == "" {
		t.Error("HashAndSalt returned empty string")
	}
	if hash == string(plain) {
		t.Error("HashAndSalt returned plain text")
	}
	// same input should produce different hashes (bcrypt uses salt)
	hash2 := HashAndSalt(plain)
	if hash == hash2 {
		t.Error("HashAndSalt should produce different hashes for same input (salt)")
	}
}

func TestHashAndSalt_EmptyInput(t *testing.T) {
	hash := HashAndSalt([]byte{})
	if hash == "" {
		t.Error("HashAndSalt returned empty string for empty input")
	}
}

func TestComparePassword(t *testing.T) {
	plain := []byte("secret")
	hashed := HashAndSalt(plain)

	if !ComparePassword(hashed, plain) {
		t.Error("ComparePassword should return true for matching password")
	}
	if ComparePassword(hashed, []byte("wrong")) {
		t.Error("ComparePassword should return false for wrong password")
	}
}

func TestComparePassword_InvalidHash(t *testing.T) {
	if ComparePassword("not-a-valid-bcrypt-hash", []byte("x")) {
		t.Error("ComparePassword should return false for invalid hash")
	}
}
