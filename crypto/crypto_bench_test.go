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
