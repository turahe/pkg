package crypto

import (
	"testing"
)

func BenchmarkHashAndSalt(b *testing.B) {
	password := []byte("benchmark-password-123")
	b.ReportAllocs()
	var sink string
	for b.Loop() {
		sink = HashAndSalt(password)
	}
	if sink == "" {
		b.Fatal("HashAndSalt returned empty")
	}
}

func BenchmarkHashArgon2id(b *testing.B) {
	password := []byte("benchmark-password-123")
	b.ReportAllocs()
	var sink string
	for b.Loop() {
		var err error
		if sink, err = HashArgon2id(password); err != nil {
			b.Fatal(err)
		}
	}
	if sink == "" {
		b.Fatal("HashArgon2id returned empty")
	}
}

func BenchmarkComparePassword_Argon2id(b *testing.B) {
	password := []byte("benchmark-password-123")
	hash, err := HashArgon2id(password)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	var sink bool
	for b.Loop() {
		sink = ComparePassword(hash, password)
	}
	if !sink {
		b.Fatal("ComparePassword returned false")
	}
}

func BenchmarkComparePassword(b *testing.B) {
	password := []byte("benchmark-password-123")
	hash := HashAndSalt(password)
	if hash == "" {
		b.Fatal("HashAndSalt returned empty")
	}

	b.ReportAllocs()
	var sink bool
	for b.Loop() {
		sink = ComparePassword(hash, password)
	}
	if !sink {
		b.Fatal("ComparePassword returned false")
	}
}
