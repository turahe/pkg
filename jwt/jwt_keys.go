package jwt

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"strings"

	"github.com/turahe/pkg/config"
)

func loadPrivateKey(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parsePrivateKeyPEM(data)
}

// getPrivateKey returns the private key from config: embedded PEM (JWTPrivateKeyPEM) if set, otherwise from JWTPrivateKey (path or inline PEM).
func getPrivateKey(conf *config.Configuration) (any, error) {
	if len(conf.Server.JWTPrivateKeyPEM) > 0 {
		return parsePrivateKeyPEM(conf.Server.JWTPrivateKeyPEM)
	}
	if conf.Server.JWTPrivateKey == "" {
		return nil, errors.New("JWT private key required for RS256/ES256: set JWT_PRIVATE_KEY or config.Server.JWTPrivateKeyPEM (e.g. from //go:embed)")
	}
	return loadPrivateKeyFromString(conf.Server.JWTPrivateKey)
}

// getPublicKey returns the public key from config: embedded PEM (JWTPublicKeyPEM) if set, otherwise from JWTPublicKey (path or inline PEM).
func getPublicKey(conf *config.Configuration) (any, error) {
	if len(conf.Server.JWTPublicKeyPEM) > 0 {
		return parsePublicKeyPEM(conf.Server.JWTPublicKeyPEM)
	}
	if conf.Server.JWTPublicKey == "" {
		return nil, errors.New("JWT public key required for RS256/ES256: set JWT_PUBLIC_KEY or config.Server.JWTPublicKeyPEM (e.g. from //go:embed)")
	}
	return loadPublicKeyFromString(conf.Server.JWTPublicKey)
}

// loadPrivateKeyFromString loads a private key from s: if s contains "-----BEGIN", parses as PEM; otherwise treats s as file path.
func loadPrivateKeyFromString(s string) (any, error) {
	if strings.Contains(s, "-----BEGIN") {
		return parsePrivateKeyPEM([]byte(s))
	}
	return loadPrivateKey(s)
}

// loadPublicKeyFromString loads a public key from s: if s contains "-----BEGIN", parses as PEM; otherwise treats s as file path.
func loadPublicKeyFromString(s string) (any, error) {
	if strings.Contains(s, "-----BEGIN") {
		return parsePublicKeyPEM([]byte(s))
	}
	return loadPublicKey(s)
}

func parsePrivateKeyPEM(data []byte) (any, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("no PEM block found")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, errors.New("unsupported private key format")
}

func loadPublicKey(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parsePublicKeyPEM(data)
}

func parsePublicKeyPEM(data []byte) (any, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("no PEM block found")
	}
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, errors.New("unsupported public key format")
}
