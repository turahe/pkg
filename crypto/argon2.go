package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Upper bounds applied when decoding a stored hash so a corrupted or crafted value cannot force
// huge allocations or CPU time during verification.
const (
	maxArgon2Memory        = 1024 * 1024 // KiB (1 GiB)
	maxArgon2Iterations    = 64
	maxArgon2SaltLength    = 64
	maxArgon2KeyLength     = 128
	minArgon2SaltLength    = 8
	minArgon2KeyLength     = 16
	maxArgon2EncodedLength = 512
)

// ErrInvalidArgon2Hash is returned when a stored value is not a well-formed argon2 PHC string
// of the expected variant or its parameters are out of the accepted range.
var ErrInvalidArgon2Hash = errors.New("invalid argon2 hash")

// Argon2Params configures argon2i / argon2id. Memory is in KiB.
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params follows the OWASP Password Storage Cheat Sheet baseline for argon2id
// (19 MiB, 2 iterations, parallelism 1), with a 16-byte salt and 32-byte key.
var DefaultArgon2Params = Argon2Params{
	Memory:      19 * 1024,
	Iterations:  2,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

// argon2Variant binds a PHC algorithm identifier to its key derivation function.
type argon2Variant struct {
	name string // PHC id: "argon2i" or "argon2id"
	key  func(password, salt []byte, time, memory uint32, threads uint8, keyLen uint32) []byte
}

var (
	argon2iVariant  = argon2Variant{name: "argon2i", key: argon2.Key}
	argon2idVariant = argon2Variant{name: "argon2id", key: argon2.IDKey}
)

func (v argon2Variant) prefix() string { return "$" + v.name + "$" }

// HashArgon2id hashes plainPassword with argon2id and DefaultArgon2Params.
// The result is a PHC string: $argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash> (raw std base64).
func HashArgon2id(plainPassword []byte) (string, error) {
	return HashArgon2idWithParams(plainPassword, DefaultArgon2Params)
}

// HashArgon2idWithParams hashes plainPassword with argon2id and the given parameters.
func HashArgon2idWithParams(plainPassword []byte, p Argon2Params) (string, error) {
	return argon2idVariant.hash(plainPassword, p)
}

// VerifyArgon2id reports whether plainPassword matches the argon2id PHC string encodedHash.
// Returns ErrInvalidArgon2Hash if encodedHash is malformed; a wrong password is (false, nil).
func VerifyArgon2id(encodedHash string, plainPassword []byte) (bool, error) {
	return argon2idVariant.verify(encodedHash, plainPassword)
}

// IsArgon2idHash reports whether encodedHash looks like an argon2id PHC string.
func IsArgon2idHash(encodedHash string) bool {
	return strings.HasPrefix(encodedHash, argon2idVariant.prefix())
}

func (v argon2Variant) hash(plainPassword []byte, p Argon2Params) (string, error) {
	if err := p.validate(); err != nil {
		return "", fmt.Errorf("%s: %w", v.name, err)
	}
	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("%s salt: %w", v.name, err)
	}
	key := v.key(plainPassword, salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)
	b64 := base64.RawStdEncoding
	return fmt.Sprintf("%sv=%d$m=%d,t=%d,p=%d$%s$%s",
		v.prefix(), argon2.Version, p.Memory, p.Iterations, p.Parallelism,
		b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

func (v argon2Variant) verify(encodedHash string, plainPassword []byte) (bool, error) {
	p, salt, key, err := v.decode(encodedHash)
	if err != nil {
		return false, err
	}
	other := v.key(plainPassword, salt, p.Iterations, p.Memory, p.Parallelism, uint32(len(key)))
	return subtle.ConstantTimeCompare(key, other) == 1, nil
}

func (p Argon2Params) validate() error {
	switch {
	case p.Memory == 0 || p.Memory > maxArgon2Memory:
		return fmt.Errorf("memory must be 1..%d KiB", maxArgon2Memory)
	case p.Memory < 8*uint32(p.Parallelism):
		return errors.New("memory must be at least 8 KiB per lane")
	case p.Iterations == 0 || p.Iterations > maxArgon2Iterations:
		return fmt.Errorf("iterations must be 1..%d", maxArgon2Iterations)
	case p.Parallelism == 0:
		return errors.New("parallelism must be >= 1")
	case p.SaltLength < minArgon2SaltLength || p.SaltLength > maxArgon2SaltLength:
		return fmt.Errorf("salt length must be %d..%d", minArgon2SaltLength, maxArgon2SaltLength)
	case p.KeyLength < minArgon2KeyLength || p.KeyLength > maxArgon2KeyLength:
		return fmt.Errorf("key length must be %d..%d", minArgon2KeyLength, maxArgon2KeyLength)
	}
	return nil
}

// decode parses $<variant>$v=19$m=<KiB>,t=<iter>,p=<lanes>$<salt>$<key>.
func (v argon2Variant) decode(encodedHash string) (Argon2Params, []byte, []byte, error) {
	var p Argon2Params
	if len(encodedHash) > maxArgon2EncodedLength {
		return p, nil, nil, ErrInvalidArgon2Hash
	}
	parts := strings.Split(encodedHash, "$")
	// "", variant, "v=19", "m=..,t=..,p=..", salt, key
	if len(parts) != 6 || parts[0] != "" || parts[1] != v.name {
		return p, nil, nil, ErrInvalidArgon2Hash
	}
	if parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return p, nil, nil, ErrInvalidArgon2Hash
	}
	var parallelism uint32
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &parallelism); err != nil ||
		parallelism == 0 || parallelism > 255 ||
		parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", p.Memory, p.Iterations, parallelism) {
		return p, nil, nil, ErrInvalidArgon2Hash
	}
	p.Parallelism = uint8(parallelism)
	b64 := base64.RawStdEncoding
	salt, err := b64.DecodeString(parts[4])
	if err != nil {
		return p, nil, nil, ErrInvalidArgon2Hash
	}
	key, err := b64.DecodeString(parts[5])
	if err != nil {
		return p, nil, nil, ErrInvalidArgon2Hash
	}
	p.SaltLength = uint32(len(salt))
	p.KeyLength = uint32(len(key))
	if p.validate() != nil {
		return p, nil, nil, ErrInvalidArgon2Hash
	}
	return p, salt, key, nil
}
