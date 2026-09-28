package crypto

import (
	"github.com/turahe/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

// HashAndSalt hashes plainPassword with bcrypt (DefaultCost) and returns the hash string. On bcrypt error logs and returns "".
// Hashes created with a lower cost still verify with ComparePassword (the cost is stored in the hash).
// To honour HASH_DRIVER use Hash (after Setup) or a Hasher from NewHasher instead.
func HashAndSalt(plainPassword []byte) string {
	hash, err := bcrypt.GenerateFromPassword(plainPassword, bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("Failed to HashAndSalt: %v", err)
		return ""
	}
	return string(hash)
}

// ComparePassword returns true if plainPassword matches hashedPassword. Accepts bcrypt, argon2i and
// argon2id hashes (detected by prefix), so hashes from any driver can coexist while migrating.
// Malformed hashes are logged and return false; empty or unrecognised values return false silently.
func ComparePassword(hashedPassword string, plainPassword []byte) bool {
	return compare(hashedPassword, plainPassword)
}
