package jwt

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turahe/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// TokenType identifies the kind of JWT (access, refresh, impersonation).
const (
	TokenTypeAccess       = "access"
	TokenTypeRefresh      = "refresh"
	TokenTypeImpersonation = "impersonation"
)

// Manager holds JWT signing and verification configuration. Create with NewManager for all-in-one use.
// For split services use NewSigner (auth server) and NewVerifier (API servers).
type Manager struct {
	signingMethod   jwt.SigningMethod
	signKey         any
	verifyKey       any
	accessExpiry    time.Duration
	refreshExpiry   time.Duration
	issuer          string
	audience        []string
	kid             string
}

// Signer issues JWTs (private key or secret only). Use for auth/login services.
type Signer struct {
	signingMethod jwt.SigningMethod
	signKey       any
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
	audience      []string
	kid           string
}

// Verifier validates JWTs (public key or secret only). Use for API/gateway services that only verify.
type Verifier struct {
	signingMethod jwt.SigningMethod
	verifyKey     any
	kid           string
}

// TokenVerifier is implemented by *Manager and *Verifier. Use it in auth middleware so either can be passed.
type TokenVerifier interface {
	ValidateToken(tokenString string) (*Claims, error)
}

// Claims is the JWT payload. It embeds jwt.RegisteredClaims (exp, iat, nbf, sub, iss, aud, jti)
// and adds UUID, TokenType, and optional impersonation fields.
type Claims struct {
	UUID string `json:"uuid"`

	jwt.RegisteredClaims

	TokenType string `json:"token_type,omitempty"` // "access", "refresh", "impersonation"

	ImpersonatorID   string `json:"impersonator_id,omitempty"`
	ImpersonatorRole string `json:"impersonator_role,omitempty"`
	IsImpersonating  bool   `json:"is_impersonating,omitempty"`
	OriginalSub      string `json:"original_sub,omitempty"`
}

// NewManager builds a JWT Manager from config: loads secret or keys from env or file paths.
// Returns an error instead of panicking when config is invalid or keys cannot be loaded.
func NewManager(ctx context.Context, conf *config.Configuration) (*Manager, error) {
	if conf == nil {
		return nil, errors.New("config is required")
	}

	alg := strings.ToUpper(strings.TrimSpace(conf.Server.JWTSigningAlgorithm))
	if alg == "" {
		alg = "RS256"
	}
	if alg != "HS256" && alg != "RS256" && alg != "ES256" {
		return nil, fmt.Errorf("JWT signing algorithm must be HS256, RS256, or ES256; got %q", alg)
	}

	m := &Manager{
		accessExpiry:  time.Hour,
		refreshExpiry: 7 * 24 * time.Hour,
		issuer:        strings.TrimSpace(conf.Server.JWTIssuer),
		kid:           strings.TrimSpace(conf.Server.JWTKeyID),
	}
	if conf.Server.AccessTokenExpiry > 0 {
		m.accessExpiry = time.Duration(conf.Server.AccessTokenExpiry) * time.Hour
	}
	if conf.Server.RefreshTokenExpiry > 0 {
		m.refreshExpiry = time.Duration(conf.Server.RefreshTokenExpiry) * 24 * time.Hour
	}
	if a := strings.TrimSpace(conf.Server.JWTAudience); a != "" {
		for _, s := range strings.Split(a, ",") {
			if t := strings.TrimSpace(s); t != "" {
				m.audience = append(m.audience, t)
			}
		}
	}

	if err := m.loadFromEnvOrFiles(conf, alg); err != nil {
		return nil, err
	}

	return m, nil
}

// NewSigner builds a JWT Signer from config (loads only private key or secret). Use for auth services that issue tokens.
func NewSigner(ctx context.Context, conf *config.Configuration) (*Signer, error) {
	if conf == nil {
		return nil, errors.New("config is required")
	}
	alg := strings.ToUpper(strings.TrimSpace(conf.Server.JWTSigningAlgorithm))
	if alg == "" {
		alg = "RS256"
	}
	if alg != "HS256" && alg != "RS256" && alg != "ES256" {
		return nil, fmt.Errorf("JWT signing algorithm must be HS256, RS256, or ES256; got %q", alg)
	}
	s := &Signer{
		accessExpiry:  time.Hour,
		refreshExpiry: 7 * 24 * time.Hour,
		issuer:        strings.TrimSpace(conf.Server.JWTIssuer),
		kid:           strings.TrimSpace(conf.Server.JWTKeyID),
	}
	if conf.Server.AccessTokenExpiry > 0 {
		s.accessExpiry = time.Duration(conf.Server.AccessTokenExpiry) * time.Hour
	}
	if conf.Server.RefreshTokenExpiry > 0 {
		s.refreshExpiry = time.Duration(conf.Server.RefreshTokenExpiry) * 24 * time.Hour
	}
	if a := strings.TrimSpace(conf.Server.JWTAudience); a != "" {
		for _, part := range strings.Split(a, ",") {
			if t := strings.TrimSpace(part); t != "" {
				s.audience = append(s.audience, t)
			}
		}
	}
	if err := s.loadSignKey(ctx, conf, alg); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Signer) loadSignKey(ctx context.Context, conf *config.Configuration, alg string) error {
	return s.loadSignKeyFromEnvOrFiles(conf, alg)
}

func (s *Signer) loadSignKeyFromEnvOrFiles(conf *config.Configuration, alg string) error {
	switch alg {
	case "HS256":
		if conf.Server.Secret == "" {
			return errors.New("JWT secret is not configured (required for HS256). Set SERVER_SECRET or use JWT Secret Manager")
		}
		s.signingMethod = jwt.SigningMethodHS256
		s.signKey = []byte(conf.Server.Secret)
	case "RS256":
		if conf.Server.JWTPrivateKeyPath == "" {
			return errors.New("JWT_PRIVATE_KEY_PATH is required for RS256")
		}
		key, err := loadPrivateKey(conf.Server.JWTPrivateKeyPath)
		if err != nil {
			return fmt.Errorf("JWT RS256 private key: %w", err)
		}
		if _, ok := key.(*rsa.PrivateKey); !ok {
			return errors.New("JWT RS256 private key is not RSA")
		}
		s.signingMethod = jwt.SigningMethodRS256
		s.signKey = key
	case "ES256":
		if conf.Server.JWTPrivateKeyPath == "" {
			return errors.New("JWT_PRIVATE_KEY_PATH is required for ES256")
		}
		key, err := loadPrivateKey(conf.Server.JWTPrivateKeyPath)
		if err != nil {
			return fmt.Errorf("JWT ES256 private key: %w", err)
		}
		if _, ok := key.(*ecdsa.PrivateKey); !ok {
			return errors.New("JWT ES256 private key is not ECDSA")
		}
		s.signingMethod = jwt.SigningMethodES256
		s.signKey = key
	}
	return nil
}

// NewVerifier builds a JWT Verifier from config (loads only public key or secret). Use for API services that only validate tokens.
func NewVerifier(ctx context.Context, conf *config.Configuration) (*Verifier, error) {
	if conf == nil {
		return nil, errors.New("config is required")
	}
	alg := strings.ToUpper(strings.TrimSpace(conf.Server.JWTSigningAlgorithm))
	if alg == "" {
		alg = "RS256"
	}
	if alg != "HS256" && alg != "RS256" && alg != "ES256" {
		return nil, fmt.Errorf("JWT signing algorithm must be HS256, RS256, or ES256; got %q", alg)
	}
	v := &Verifier{kid: strings.TrimSpace(conf.Server.JWTKeyID)}
	if err := v.loadVerifyKey(ctx, conf, alg); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *Verifier) loadVerifyKey(ctx context.Context, conf *config.Configuration, alg string) error {
	return v.loadVerifyKeyFromEnvOrFiles(conf, alg)
}

func (v *Verifier) loadVerifyKeyFromEnvOrFiles(conf *config.Configuration, alg string) error {
	switch alg {
	case "HS256":
		if conf.Server.Secret == "" {
			return errors.New("JWT secret is not configured (required for HS256). Set SERVER_SECRET")
		}
		v.signingMethod = jwt.SigningMethodHS256
		v.verifyKey = []byte(conf.Server.Secret)
	case "RS256":
		if conf.Server.JWTPublicKeyPath == "" {
			return errors.New("JWT_PUBLIC_KEY_PATH is required for RS256")
		}
		key, err := loadPublicKey(conf.Server.JWTPublicKeyPath)
		if err != nil {
			return fmt.Errorf("JWT RS256 public key: %w", err)
		}
		if _, ok := key.(*rsa.PublicKey); !ok {
			return errors.New("JWT RS256 public key is not RSA")
		}
		v.signingMethod = jwt.SigningMethodRS256
		v.verifyKey = key
	case "ES256":
		if conf.Server.JWTPublicKeyPath == "" {
			return errors.New("JWT_PUBLIC_KEY_PATH is required for ES256")
		}
		key, err := loadPublicKey(conf.Server.JWTPublicKeyPath)
		if err != nil {
			return fmt.Errorf("JWT ES256 public key: %w", err)
		}
		if _, ok := key.(*ecdsa.PublicKey); !ok {
			return errors.New("JWT ES256 public key is not ECDSA")
		}
		v.signingMethod = jwt.SigningMethodES256
		v.verifyKey = key
	}
	return nil
}

// ValidateToken implements TokenVerifier.
func (v *Verifier) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != v.signingMethod.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			if v.kid != "" && token.Header["kid"] != nil {
				if k, _ := token.Header["kid"].(string); k != v.kid {
					return nil, errors.New("key id mismatch")
				}
			}
			return v.verifyKey, nil
		},
		jwt.WithValidMethods([]string{v.signingMethod.Alg()}),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// buildRegisteredClaims (Signer) sets exp, iat, nbf, sub, jti, iss, aud.
func (s *Signer) buildRegisteredClaims(sub string, expiry time.Duration) jwt.RegisteredClaims {
	now := time.Now()
	rc := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
		Subject:   sub,
		ID:        uuid.NewString(),
	}
	if s.issuer != "" {
		rc.Issuer = s.issuer
	}
	if len(s.audience) > 0 {
		rc.Audience = s.audience
	}
	return rc
}

func (s *Signer) signToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(s.signingMethod, claims)
	if s.kid != "" {
		token.Header["kid"] = s.kid
	}
	return token.SignedString(s.signKey)
}

// GenerateToken issues a signed JWT (token_type: access).
func (s *Signer) GenerateToken(id uuid.UUID) (string, error) {
	return s.GenerateTokenWithExpiry(id, s.accessExpiry)
}

// GenerateTokenWithExpiry issues a signed JWT (token_type: access) with custom expiry.
func (s *Signer) GenerateTokenWithExpiry(id uuid.UUID, expiry time.Duration) (string, error) {
	claims := Claims{
		UUID:            id.String(),
		RegisteredClaims: s.buildRegisteredClaims(id.String(), expiry),
		TokenType:       TokenTypeAccess,
	}
	return s.signToken(claims)
}

// GenerateRefreshToken issues a signed JWT (token_type: refresh).
func (s *Signer) GenerateRefreshToken(id uuid.UUID) (string, error) {
	claims := Claims{
		UUID:            id.String(),
		RegisteredClaims: s.buildRegisteredClaims(id.String(), s.refreshExpiry),
		TokenType:       TokenTypeRefresh,
	}
	return s.signToken(claims)
}

// GenerateImpersonationToken issues a short-lived JWT (token_type: impersonation). TTL clamped to max 30 minutes.
func (s *Signer) GenerateImpersonationToken(adminID uuid.UUID, adminRole string, targetUserID uuid.UUID, requestedTTL time.Duration) (string, error) {
	maxTTL := 30 * time.Minute
	ttl := requestedTTL
	if ttl <= 0 || ttl > maxTTL {
		ttl = maxTTL
	}
	claims := Claims{
		UUID: targetUserID.String(),
		RegisteredClaims: s.buildRegisteredClaims(targetUserID.String(), ttl),
		TokenType:        TokenTypeImpersonation,
		ImpersonatorID:   adminID.String(),
		ImpersonatorRole: adminRole,
		IsImpersonating:  true,
		OriginalSub:      adminID.String(),
	}
	return s.signToken(claims)
}

func (m *Manager) loadFromEnvOrFiles(conf *config.Configuration, alg string) error {
	switch alg {
	case "HS256":
		secret := conf.Server.Secret
		if secret == "" {
			return errors.New("JWT secret is not configured (required for HS256). Set SERVER_SECRET")
		}
		m.signingMethod = jwt.SigningMethodHS256
		m.signKey = []byte(secret)
		m.verifyKey = []byte(secret)
	case "RS256":
		if conf.Server.JWTPrivateKeyPath == "" || conf.Server.JWTPublicKeyPath == "" {
			return errors.New("JWT_PRIVATE_KEY_PATH and JWT_PUBLIC_KEY_PATH are required for RS256")
		}
		privateKey, err := loadPrivateKey(conf.Server.JWTPrivateKeyPath)
		if err != nil {
			return fmt.Errorf("JWT RS256 private key: %w", err)
		}
		rsaPrivate, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return errors.New("JWT RS256 private key is not RSA")
		}
		publicKey, err := loadPublicKey(conf.Server.JWTPublicKeyPath)
		if err != nil {
			return fmt.Errorf("JWT RS256 public key: %w", err)
		}
		rsaPublic, ok := publicKey.(*rsa.PublicKey)
		if !ok {
			return errors.New("JWT RS256 public key is not RSA")
		}
		m.signingMethod = jwt.SigningMethodRS256
		m.signKey = rsaPrivate
		m.verifyKey = rsaPublic
	case "ES256":
		if conf.Server.JWTPrivateKeyPath == "" || conf.Server.JWTPublicKeyPath == "" {
			return errors.New("JWT_PRIVATE_KEY_PATH and JWT_PUBLIC_KEY_PATH are required for ES256")
		}
		privateKey, err := loadPrivateKey(conf.Server.JWTPrivateKeyPath)
		if err != nil {
			return fmt.Errorf("JWT ES256 private key: %w", err)
		}
		ecPrivate, ok := privateKey.(*ecdsa.PrivateKey)
		if !ok {
			return errors.New("JWT ES256 private key is not ECDSA")
		}
		publicKey, err := loadPublicKey(conf.Server.JWTPublicKeyPath)
		if err != nil {
			return fmt.Errorf("JWT ES256 public key: %w", err)
		}
		ecPublic, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return errors.New("JWT ES256 public key is not ECDSA")
		}
		m.signingMethod = jwt.SigningMethodES256
		m.signKey = ecPrivate
		m.verifyKey = ecPublic
	}
	return nil
}

func loadPrivateKey(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parsePrivateKeyPEM(data)
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

// buildRegisteredClaims sets exp, iat, nbf, sub, jti, iss, aud from Manager.
func (m *Manager) buildRegisteredClaims(sub string, expiry time.Duration) jwt.RegisteredClaims {
	now := time.Now()
	rc := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
		Subject:   sub,
		ID:        uuid.NewString(),
	}
	if m.issuer != "" {
		rc.Issuer = m.issuer
	}
	if len(m.audience) > 0 {
		rc.Audience = m.audience
	}
	return rc
}

// signToken creates a JWT, sets optional kid header, and signs. Claims must already include TokenType.
func (m *Manager) signToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(m.signingMethod, claims)
	if m.kid != "" {
		token.Header["kid"] = m.kid
	}
	return token.SignedString(m.signKey)
}

// GetAccessTokenExpiry returns the access token expiry duration.
func (m *Manager) GetAccessTokenExpiry() time.Duration {
	return m.accessExpiry
}

// GenerateToken issues a signed JWT with the given UUID and default access token expiry (token_type: "access").
func (m *Manager) GenerateToken(id uuid.UUID) (string, error) {
	return m.GenerateTokenWithExpiry(id, m.accessExpiry)
}

// GenerateTokenWithExpiry issues a signed JWT with the given UUID and custom expiry (token_type: "access").
func (m *Manager) GenerateTokenWithExpiry(id uuid.UUID, expiry time.Duration) (string, error) {
	claims := Claims{
		UUID:            id.String(),
		RegisteredClaims: m.buildRegisteredClaims(id.String(), expiry),
		TokenType:       TokenTypeAccess,
	}
	return m.signToken(claims)
}

// GenerateRefreshToken issues a signed JWT with refresh token expiry (token_type: "refresh").
func (m *Manager) GenerateRefreshToken(id uuid.UUID) (string, error) {
	claims := Claims{
		UUID:            id.String(),
		RegisteredClaims: m.buildRegisteredClaims(id.String(), m.refreshExpiry),
		TokenType:       TokenTypeRefresh,
	}
	return m.signToken(claims)
}

// GenerateImpersonationToken issues a short-lived JWT for admin-as-user (token_type: "impersonation").
// TTL is clamped to max 30 minutes.
func (m *Manager) GenerateImpersonationToken(adminID uuid.UUID, adminRole string, targetUserID uuid.UUID, requestedTTL time.Duration) (string, error) {
	maxTTL := 30 * time.Minute
	ttl := requestedTTL
	if ttl <= 0 || ttl > maxTTL {
		ttl = maxTTL
	}
	claims := Claims{
		UUID: targetUserID.String(),
		RegisteredClaims: m.buildRegisteredClaims(targetUserID.String(), ttl),
		TokenType:        TokenTypeImpersonation,
		ImpersonatorID:   adminID.String(),
		ImpersonatorRole: adminRole,
		IsImpersonating:  true,
		OriginalSub:      adminID.String(),
	}
	return m.signToken(claims)
}

// ValidateToken parses the token string, verifies signature and expiry, and returns Claims or an error.
// Validates alg and optional kid against the Manager.
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != m.signingMethod.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			if m.kid != "" && token.Header["kid"] != nil {
				if k, _ := token.Header["kid"].(string); k != m.kid {
					return nil, errors.New("key id mismatch")
				}
			}
			return m.verifyKey, nil
		},
		jwt.WithValidMethods([]string{m.signingMethod.Alg()}),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// ComparePassword returns true if plainPassword matches the bcrypt hash hashedPassword.
func ComparePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// GetCurrentUserUUID reads "user_id" from the Gin context (set by auth middleware).
func GetCurrentUserUUID(ctx *gin.Context) (uuid.UUID, bool) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	switch v := userID.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil, false
		}
		return parsed, true
	case uuid.UUID:
		return v, true
	default:
		return uuid.Nil, false
	}
}
