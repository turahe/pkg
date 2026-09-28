package crypto

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/logger"
)

// Hash driver names accepted by NewHasher and HASH_DRIVER (Laravel-compatible naming).
const (
	DriverBcrypt   = "bcrypt"
	DriverArgon    = "argon"    // Argon2i
	DriverArgon2id = "argon2id" // Argon2id (recommended)
)

var (
	// ErrUnknownDriver is returned by NewHasher for an unsupported driver name.
	ErrUnknownDriver = errors.New("unknown hash driver")
	// errUnsupportedHash marks values that are not bcrypt/argon2 hashes (e.g. empty); not logged.
	errUnsupportedHash = errors.New("unsupported hash format")
)

// Hasher hashes passwords with one driver.
//
// Check accepts bcrypt, argon2i and argon2id hashes regardless of the hasher's driver, so switching
// HASH_DRIVER never locks users out; NeedsRehash reports hashes that are in another format or were
// made with different parameters, so they can be upgraded after a successful login.
type Hasher interface {
	Driver() string
	Hash(plainPassword []byte) (string, error)
	Check(hashedPassword string, plainPassword []byte) bool
	NeedsRehash(hashedPassword string) bool
}

// HasherOptions tunes NewHasher. Zero values use defaults: bcrypt.DefaultCost and DefaultArgon2Params
// (per field, so setting only Argon.Memory keeps the default iterations, parallelism, salt and key length).
type HasherOptions struct {
	BcryptCost int
	Argon      Argon2Params
}

// NewHasher returns a Hasher for driver "bcrypt", "argon" (Argon2i) or "argon2id". Empty means bcrypt.
func NewHasher(driver string, opts HasherOptions) (Hasher, error) {
	switch d := strings.ToLower(strings.TrimSpace(driver)); d {
	case "", DriverBcrypt:
		cost := opts.BcryptCost
		if cost == 0 {
			cost = bcrypt.DefaultCost
		}
		if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
			return nil, fmt.Errorf("bcrypt: cost must be %d..%d", bcrypt.MinCost, bcrypt.MaxCost)
		}
		return &bcryptHasher{cost: cost}, nil
	case DriverArgon, DriverArgon2id:
		p := opts.Argon.withDefaults()
		variant := argon2idVariant
		if d == DriverArgon {
			variant = argon2iVariant
		}
		if err := p.validate(); err != nil {
			return nil, fmt.Errorf("%s: %w", variant.name, err)
		}
		return &argon2Hasher{driver: d, variant: variant, params: p}, nil
	default:
		return nil, fmt.Errorf("%w %q (want %s, %s or %s)", ErrUnknownDriver, driver, DriverBcrypt, DriverArgon, DriverArgon2id)
	}
}

// NewHasherFromConfig builds a Hasher from HASH_* settings (see config.HashingConfiguration).
func NewHasherFromConfig(cfg config.HashingConfiguration) (Hasher, error) {
	if cfg.BcryptRounds < 0 || cfg.ArgonMemory < 0 || cfg.ArgonTime < 0 || cfg.ArgonThreads < 0 || cfg.ArgonThreads > 255 {
		return nil, errors.New("hashing: HASH_* values must be non-negative (HASH_ARGON_THREADS <= 255)")
	}
	return NewHasher(cfg.Driver, HasherOptions{
		BcryptCost: cfg.BcryptRounds,
		Argon: Argon2Params{
			Memory:      uint32(cfg.ArgonMemory),
			Iterations:  uint32(cfg.ArgonTime),
			Parallelism: uint8(cfg.ArgonThreads),
		},
	})
}

var (
	defaultMu     sync.RWMutex
	defaultHasher Hasher
)

// Setup configures the package default Hasher from config.GetConfig().Hashing (HASH_DRIVER etc.).
func Setup() error {
	h, err := NewHasherFromConfig(config.GetConfig().Hashing)
	if err != nil {
		return err
	}
	SetDefault(h)
	logger.Infof("crypto: hash driver=%s", h.Driver())
	return nil
}

// SetDefault replaces the package default Hasher; nil restores the bcrypt fallback.
func SetDefault(h Hasher) {
	defaultMu.Lock()
	defaultHasher = h
	defaultMu.Unlock()
}

// Default returns the Hasher set by Setup / SetDefault, or bcrypt with DefaultCost if none was set.
func Default() Hasher {
	defaultMu.RLock()
	h := defaultHasher
	defaultMu.RUnlock()
	if h == nil {
		return &bcryptHasher{cost: bcrypt.DefaultCost}
	}
	return h
}

// Hash hashes plainPassword with the default Hasher (see Setup).
func Hash(plainPassword []byte) (string, error) {
	return Default().Hash(plainPassword)
}

// NeedsRehash reports whether hashedPassword should be re-hashed with the default Hasher.
func NeedsRehash(hashedPassword string) bool {
	return Default().NeedsRehash(hashedPassword)
}

type bcryptHasher struct {
	cost int
}

func (h *bcryptHasher) Driver() string { return DriverBcrypt }

func (h *bcryptHasher) Hash(plainPassword []byte) (string, error) {
	b, err := bcrypt.GenerateFromPassword(plainPassword, h.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt: %w", err)
	}
	return string(b), nil
}

func (h *bcryptHasher) Check(hashedPassword string, plainPassword []byte) bool {
	return compare(hashedPassword, plainPassword)
}

func (h *bcryptHasher) NeedsRehash(hashedPassword string) bool {
	if !isBcryptHash(hashedPassword) {
		return true
	}
	cost, err := bcrypt.Cost([]byte(hashedPassword))
	return err != nil || cost != h.cost
}

type argon2Hasher struct {
	driver  string
	variant argon2Variant
	params  Argon2Params
}

func (h *argon2Hasher) Driver() string { return h.driver }

func (h *argon2Hasher) Hash(plainPassword []byte) (string, error) {
	return h.variant.hash(plainPassword, h.params)
}

func (h *argon2Hasher) Check(hashedPassword string, plainPassword []byte) bool {
	return compare(hashedPassword, plainPassword)
}

func (h *argon2Hasher) NeedsRehash(hashedPassword string) bool {
	got, _, _, err := h.variant.decode(hashedPassword)
	if err != nil {
		return true
	}
	return got.Memory != h.params.Memory || got.Iterations != h.params.Iterations ||
		got.Parallelism != h.params.Parallelism || got.KeyLength != h.params.KeyLength
}

func (p Argon2Params) withDefaults() Argon2Params {
	d := DefaultArgon2Params
	if p.Memory != 0 {
		d.Memory = p.Memory
	}
	if p.Iterations != 0 {
		d.Iterations = p.Iterations
	}
	if p.Parallelism != 0 {
		d.Parallelism = p.Parallelism
	}
	if p.SaltLength != 0 {
		d.SaltLength = p.SaltLength
	}
	if p.KeyLength != 0 {
		d.KeyLength = p.KeyLength
	}
	return d
}

// compare verifies any supported format and logs malformed hashes; a wrong password is not logged.
func compare(hashedPassword string, plainPassword []byte) bool {
	ok, err := verifyAny(hashedPassword, plainPassword)
	if err != nil && !errors.Is(err, errUnsupportedHash) {
		logger.Errorf("Failed to ComparePassword: %v", err)
	}
	return ok
}

// verifyAny dispatches on the hash prefix: $argon2id$, $argon2i$, or bcrypt ($2a$/$2b$/$2y$).
func verifyAny(hashedPassword string, plainPassword []byte) (bool, error) {
	switch {
	case strings.HasPrefix(hashedPassword, argon2idVariant.prefix()):
		return argon2idVariant.verify(hashedPassword, plainPassword)
	case strings.HasPrefix(hashedPassword, argon2iVariant.prefix()):
		return argon2iVariant.verify(hashedPassword, plainPassword)
	case isBcryptHash(hashedPassword):
		err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), plainPassword)
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return err == nil, err
	default:
		return false, errUnsupportedHash
	}
}

// isBcryptHash reports whether s has the shape of a bcrypt hash ($2?$cost$ + 53 chars = 60 bytes).
func isBcryptHash(s string) bool {
	return len(s) == 60 && strings.HasPrefix(s, "$2")
}
