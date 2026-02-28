package crypto

import (
	"github.com/turahe/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

// HashAndSalt hashes plainPassword with bcrypt (MinCost) and returns the hash string. On bcrypt error logs and returns "".
func HashAndSalt(plainPassword []byte) string {
	hash, err := bcrypt.GenerateFromPassword(plainPassword, bcrypt.MinCost)
	if err != nil {
		logger.Errorf("Failed to HashAndSalt: %v", err)
	}
	return string(hash)
}

// ComparePassword returns true if plainPassword matches hashedPassword; logs and returns false on error.
// Returns false without calling bcrypt if hashedPassword is empty or too short to be a valid bcrypt hash,
// avoiding error logs for missing or invalid stored hashes.
func ComparePassword(hashedPassword string, plainPassword []byte) bool {
	// Bcrypt hashes are 60 bytes (e.g. $2a$10$...); reject empty or too-short to avoid bcrypt error
	if len(hashedPassword) < 60 {
		return false
	}
	byteHash := []byte(hashedPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPassword)
	if err != nil {
		logger.Errorf("Failed to ComparePassword: %v", err)
	}
	return err == nil
}
