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
func ComparePassword(hashedPassword string, plainPassword []byte) bool {
	byteHash := []byte(hashedPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPassword)
	if err != nil {
		logger.Errorf("Failed to ComparePassword: %v", err)
	}
	return err == nil
}
