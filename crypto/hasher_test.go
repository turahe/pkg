package crypto

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/turahe/pkg/config"
)

// Generated with argon2-cffi (type=I) for refArgon2Secret.
const refArgon2i = "$argon2i$v=19$m=19456,t=2,p=1$wOVKo4pil1TXiaeUx+jRXA$IjDahCQS3T1xgXPjS7P/x5uWwx4wCVQvPv6fa7k4jos"

// fastArgon keeps argon tests quick; production defaults are exercised by argon2_test.go.
var fastArgon = HasherOptions{BcryptCost: bcrypt.MinCost, Argon: Argon2Params{Memory: 64, Iterations: 1}}

func mustHasher(t *testing.T, driver string, opts HasherOptions) Hasher {
	t.Helper()
	h, err := NewHasher(driver, opts)
	if err != nil {
		t.Fatalf("NewHasher(%q): %v", driver, err)
	}
	return h
}

func TestNewHasher_Drivers(t *testing.T) {
	cases := []struct {
		driver     string
		wantDriver string
		wantPrefix string
	}{
		{"bcrypt", DriverBcrypt, "$2a$"},
		{"", DriverBcrypt, "$2a$"},
		{"argon", DriverArgon, "$argon2i$v=19$m=64,t=1,p=1$"},
		{" ARGON2ID ", DriverArgon2id, "$argon2id$v=19$m=64,t=1,p=1$"},
	}
	for _, tc := range cases {
		t.Run(tc.driver, func(t *testing.T) {
			h := mustHasher(t, tc.driver, fastArgon)
			if h.Driver() != tc.wantDriver {
				t.Errorf("Driver() = %q, want %q", h.Driver(), tc.wantDriver)
			}
			hash, err := h.Hash([]byte("secret"))
			if err != nil {
				t.Fatalf("Hash: %v", err)
			}
			if !strings.HasPrefix(hash, tc.wantPrefix) {
				t.Errorf("hash = %q, want prefix %q", hash, tc.wantPrefix)
			}
			if !h.Check(hash, []byte("secret")) || h.Check(hash, []byte("wrong")) {
				t.Error("Check round-trip failed")
			}
			if h.NeedsRehash(hash) {
				t.Error("fresh hash must not need rehash")
			}
		})
	}
}

func TestNewHasher_Errors(t *testing.T) {
	if _, err := NewHasher("md5", HasherOptions{}); !errors.Is(err, ErrUnknownDriver) {
		t.Errorf("unknown driver err = %v, want ErrUnknownDriver", err)
	}
	if _, err := NewHasher(DriverBcrypt, HasherOptions{BcryptCost: 3}); err == nil {
		t.Error("bcrypt cost below MinCost must fail")
	}
	if _, err := NewHasher(DriverBcrypt, HasherOptions{BcryptCost: 32}); err == nil {
		t.Error("bcrypt cost above MaxCost must fail")
	}
	if _, err := NewHasher(DriverArgon2id, HasherOptions{Argon: Argon2Params{Memory: 8, Parallelism: 4}}); err == nil {
		t.Error("invalid argon params must fail")
	}
	if _, err := mustHasher(t, DriverBcrypt, fastArgon).Hash(make([]byte, 73)); err == nil {
		t.Error("bcrypt must reject passwords longer than 72 bytes")
	}
}

func TestNewHasher_CustomSaltAndKeyLength(t *testing.T) {
	h := mustHasher(t, DriverArgon2id, HasherOptions{Argon: Argon2Params{Memory: 64, Iterations: 1, SaltLength: 32, KeyLength: 64}})
	hash, err := h.Hash([]byte("x"))
	if err != nil {
		t.Fatal(err)
	}
	p, salt, key, err := argon2idVariant.decode(hash)
	if err != nil {
		t.Fatal(err)
	}
	if len(salt) != 32 || len(key) != 64 || p.Parallelism != DefaultArgon2Params.Parallelism {
		t.Errorf("salt=%d key=%d p=%d; want 32, 64, default parallelism", len(salt), len(key), p.Parallelism)
	}
}

func TestHasher_ChecksEveryFormat(t *testing.T) {
	bcryptHash := HashAndSalt([]byte(refArgon2Secret))
	for _, driver := range []string{DriverBcrypt, DriverArgon, DriverArgon2id} {
		h := mustHasher(t, driver, fastArgon)
		for name, hash := range map[string]string{"bcrypt": bcryptHash, "argon2i": refArgon2i, "argon2id": refArgon2OWASP} {
			if !h.Check(hash, []byte(refArgon2Secret)) {
				t.Errorf("%s driver: Check(%s hash) = false", driver, name)
			}
			if h.Check(hash, []byte("wrong")) {
				t.Errorf("%s driver: Check(%s hash, wrong) = true", driver, name)
			}
		}
		if h.Check("", []byte("x")) || h.Check("plaintext", []byte("plaintext")) {
			t.Errorf("%s driver: non-hash values must not verify", driver)
		}
	}
}

func TestHasher_NeedsRehash(t *testing.T) {
	bcrypt10 := mustHasher(t, DriverBcrypt, HasherOptions{})
	bcrypt4 := mustHasher(t, DriverBcrypt, HasherOptions{BcryptCost: bcrypt.MinCost})
	argonI := mustHasher(t, DriverArgon, HasherOptions{})
	argonID := mustHasher(t, DriverArgon2id, HasherOptions{})

	legacy, err := bcrypt4.Hash([]byte("x"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		hasher Hasher
		hash   string
		want   bool
	}{
		{"bcrypt: lower cost", bcrypt10, legacy, true},
		{"bcrypt: argon hash", bcrypt10, refArgon2OWASP, true},
		{"argon: matching argon2i", argonI, refArgon2i, false},
		{"argon: argon2id hash", argonI, refArgon2OWASP, true},
		{"argon2id: matching defaults", argonID, refArgon2OWASP, false},
		{"argon2id: different params (RFC 9106)", argonID, refArgon2RFC, true},
		{"argon2id: argon2i hash", argonID, refArgon2i, true},
		{"argon2id: bcrypt hash", argonID, legacy, true},
		{"argon2id: empty", argonID, "", true},
	}
	for _, tc := range cases {
		if got := tc.hasher.NeedsRehash(tc.hash); got != tc.want {
			t.Errorf("%s: NeedsRehash = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestNewHasherFromConfig(t *testing.T) {
	h, err := NewHasherFromConfig(config.HashingConfiguration{Driver: "argon2id", ArgonMemory: 64, ArgonTime: 1, ArgonThreads: 2})
	if err != nil {
		t.Fatalf("NewHasherFromConfig: %v", err)
	}
	hash, err := h.Hash([]byte("x"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=64,t=1,p=2$") {
		t.Errorf("hash = %q, want configured params", hash)
	}

	h, err = NewHasherFromConfig(config.HashingConfiguration{Driver: "bcrypt", BcryptRounds: 5})
	if err != nil {
		t.Fatal(err)
	}
	hash, _ = h.Hash([]byte("x"))
	if cost, _ := bcrypt.Cost([]byte(hash)); cost != 5 {
		t.Errorf("bcrypt cost = %d, want 5", cost)
	}

	for _, bad := range []config.HashingConfiguration{
		{Driver: "argon2id", ArgonMemory: -1},
		{Driver: "argon2id", ArgonThreads: 256},
		{Driver: "scrypt"},
	} {
		if _, err := NewHasherFromConfig(bad); err == nil {
			t.Errorf("NewHasherFromConfig(%+v) = nil error", bad)
		}
	}
}

func TestSetupAndDefault(t *testing.T) {
	origCfg := config.Config
	t.Cleanup(func() {
		config.Config = origCfg
		SetDefault(nil)
	})

	SetDefault(nil)
	if Default().Driver() != DriverBcrypt {
		t.Errorf("fallback driver = %q, want bcrypt", Default().Driver())
	}

	config.Config = &config.Configuration{Hashing: config.HashingConfiguration{Driver: "argon2id", ArgonMemory: 64, ArgonTime: 1}}
	if err := Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if Default().Driver() != DriverArgon2id {
		t.Errorf("Default().Driver() = %q, want argon2id", Default().Driver())
	}
	hash, err := Hash([]byte("secret"))
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !IsArgon2idHash(hash) || !ComparePassword(hash, []byte("secret")) {
		t.Errorf("Hash via default = %q, want verifiable argon2id", hash)
	}
	if NeedsRehash(hash) {
		t.Error("NeedsRehash(fresh default hash) = true")
	}
	if !NeedsRehash(HashAndSalt([]byte("secret"))) {
		t.Error("bcrypt hash must need rehash under argon2id default")
	}

	config.Config = &config.Configuration{Hashing: config.HashingConfiguration{Driver: "nope"}}
	if err := Setup(); !errors.Is(err, ErrUnknownDriver) {
		t.Errorf("Setup(unknown) err = %v, want ErrUnknownDriver", err)
	}
	if Default().Driver() != DriverArgon2id {
		t.Error("failed Setup must keep the previous default")
	}
}
