package jwt

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/turahe/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType identifies the kind of JWT (access, refresh, impersonation).
const (
	TokenTypeAccess        = "access"
	TokenTypeRefresh       = "refresh"
	TokenTypeImpersonation = "impersonation"
)

// ActorType identifies who the subject represents (user/admin from app tables, or machine identity).
const (
	ActorTypeUser    = "user"
	ActorTypeAdmin   = "admin"
	ActorTypeService = "service"
	ActorTypeSystem  = "system"
)

// ResolveActorType maps an actor type or table name to a canonical actor_type claim.
// Empty → service. Table "admins" → admin. Table "users"/"User" → user.
// Explicit "system" / "service" / "user" / "admin" are kept as-is (case-insensitive).
func ResolveActorType(actorOrTable string) string {
	s := strings.TrimSpace(strings.ToLower(actorOrTable))
	switch s {
	case "":
		return ActorTypeService
	case ActorTypeUser, "users":
		return ActorTypeUser
	case ActorTypeAdmin, "admins":
		return ActorTypeAdmin
	case ActorTypeService, "services":
		return ActorTypeService
	case ActorTypeSystem, "systems":
		return ActorTypeSystem
	default:
		return ActorTypeService
	}
}

// Manager holds JWT signing and verification configuration. Create with NewManager for all-in-one use.
// For split services use NewSigner (auth server) and NewVerifier (API servers).
// Supported algorithms: RS256 (default) and ES256.
type Manager struct {
	signingMethod jwt.SigningMethod
	signKey       any
	verifyKey     any
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
	audience      []string
	kid           string
}

// Signer issues JWTs (private key only). Use for auth/login services.
type Signer struct {
	signingMethod jwt.SigningMethod
	signKey       any
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
	audience      []string
	kid           string
}

// Verifier validates JWTs (public key only). Use for API/gateway services that only verify.
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
// and adds UUID, ActorType, TokenType, and optional impersonation fields.
type Claims struct {
	UUID string `json:"uuid"`

	jwt.RegisteredClaims

	// ActorType is "user" (users table), "admin" (admins table), "service", or "system".
	// Empty input at issue time resolves to "service" via ResolveActorType.
	ActorType string `json:"actor_type,omitempty"`
	TokenType string `json:"token_type,omitempty"` // "access", "refresh", "impersonation"

	ImpersonatorID   string `json:"impersonator_id,omitempty"`
	ImpersonatorRole string `json:"impersonator_role,omitempty"`
	IsImpersonating  bool   `json:"is_impersonating,omitempty"`
	OriginalSub      string `json:"original_sub,omitempty"`
}

// NewManager builds a JWT Manager from config: loads asymmetric keys from env, file paths, or embedded PEM.
// Returns an error instead of panicking when config is invalid or keys cannot be loaded.
func NewManager(ctx context.Context, conf *config.Configuration) (*Manager, error) {
	if conf == nil {
		return nil, errors.New("config is required")
	}

	alg, err := normalizeSigningAlgorithm(conf.Server.JWTSigningAlgorithm)
	if err != nil {
		return nil, err
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

// NewSigner builds a JWT Signer from config (loads only the private key). Use for auth services that issue tokens.
func NewSigner(ctx context.Context, conf *config.Configuration) (*Signer, error) {
	if conf == nil {
		return nil, errors.New("config is required")
	}
	alg, err := normalizeSigningAlgorithm(conf.Server.JWTSigningAlgorithm)
	if err != nil {
		return nil, err
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
	case "RS256":
		key, err := getPrivateKey(conf)
		if err != nil {
			return err
		}
		if _, ok := key.(*rsa.PrivateKey); !ok {
			return errors.New("JWT RS256 private key is not RSA")
		}
		s.signingMethod = jwt.SigningMethodRS256
		s.signKey = key
	case "ES256":
		key, err := getPrivateKey(conf)
		if err != nil {
			return err
		}
		if _, ok := key.(*ecdsa.PrivateKey); !ok {
			return errors.New("JWT ES256 private key is not ECDSA")
		}
		s.signingMethod = jwt.SigningMethodES256
		s.signKey = key
	default:
		return fmt.Errorf("JWT signing algorithm must be RS256 or ES256; got %q", alg)
	}
	return nil
}

// NewVerifier builds a JWT Verifier from config (loads only the public key). Use for API services that only validate tokens.
func NewVerifier(ctx context.Context, conf *config.Configuration) (*Verifier, error) {
	if conf == nil {
		return nil, errors.New("config is required")
	}
	alg, err := normalizeSigningAlgorithm(conf.Server.JWTSigningAlgorithm)
	if err != nil {
		return nil, err
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
	case "RS256":
		key, err := getPublicKey(conf)
		if err != nil {
			return err
		}
		if _, ok := key.(*rsa.PublicKey); !ok {
			return errors.New("JWT RS256 public key is not RSA")
		}
		v.signingMethod = jwt.SigningMethodRS256
		v.verifyKey = key
	case "ES256":
		key, err := getPublicKey(conf)
		if err != nil {
			return err
		}
		if _, ok := key.(*ecdsa.PublicKey); !ok {
			return errors.New("JWT ES256 public key is not ECDSA")
		}
		v.signingMethod = jwt.SigningMethodES256
		v.verifyKey = key
	default:
		return fmt.Errorf("JWT signing algorithm must be RS256 or ES256; got %q", alg)
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
			kid, _ := token.Header["kid"].(string)
			if err := matchConfiguredKid(v.kid, kid); err != nil {
				return nil, err
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
// actorOrTable is an actor type ("user","admin","service","system") or table name
// ("users","admins","User"); empty defaults to service. See ResolveActorType.
func (s *Signer) GenerateToken(id uuid.UUID, actorOrTable ...string) (string, error) {
	return s.GenerateTokenWithExpiry(id, s.accessExpiry, actorOrTable...)
}

// GenerateTokenWithExpiry issues a signed JWT (token_type: access) with custom expiry.
func (s *Signer) GenerateTokenWithExpiry(id uuid.UUID, expiry time.Duration, actorOrTable ...string) (string, error) {
	claims := Claims{
		UUID:             id.String(),
		RegisteredClaims: s.buildRegisteredClaims(id.String(), expiry),
		ActorType:        resolveActorArg(actorOrTable...),
		TokenType:        TokenTypeAccess,
	}
	return s.signToken(claims)
}

// GenerateRefreshToken issues a signed JWT (token_type: refresh).
func (s *Signer) GenerateRefreshToken(id uuid.UUID, actorOrTable ...string) (string, error) {
	claims := Claims{
		UUID:             id.String(),
		RegisteredClaims: s.buildRegisteredClaims(id.String(), s.refreshExpiry),
		ActorType:        resolveActorArg(actorOrTable...),
		TokenType:        TokenTypeRefresh,
	}
	return s.signToken(claims)
}

// ImpersonationParams groups inputs for impersonation token generation.
type ImpersonationParams struct {
	AdminID      uuid.UUID
	AdminRole    string
	TargetUserID uuid.UUID
	TTL          time.Duration
}

// GenerateImpersonationToken issues a short-lived JWT (token_type: impersonation). TTL clamped to max 30 minutes.
// ActorType is always "user" (target subject); ImpersonatorRole carries the admin role.
func (s *Signer) GenerateImpersonationToken(adminID uuid.UUID, adminRole string, targetUserID uuid.UUID, requestedTTL time.Duration) (string, error) {
	return s.GenerateImpersonation(ImpersonationParams{
		AdminID: adminID, AdminRole: adminRole, TargetUserID: targetUserID, TTL: requestedTTL,
	})
}

func (s *Signer) GenerateImpersonation(p ImpersonationParams) (string, error) {
	maxTTL := 30 * time.Minute
	ttl := p.TTL
	if ttl <= 0 || ttl > maxTTL {
		ttl = maxTTL
	}
	claims := Claims{
		UUID:             p.TargetUserID.String(),
		RegisteredClaims: s.buildRegisteredClaims(p.TargetUserID.String(), ttl),
		ActorType:        ActorTypeUser,
		TokenType:        TokenTypeImpersonation,
		ImpersonatorID:   p.AdminID.String(),
		ImpersonatorRole: p.AdminRole,
		IsImpersonating:  true,
		OriginalSub:      p.AdminID.String(),
	}
	return s.signToken(claims)
}

func resolveActorArg(actorOrTable ...string) string {
	if len(actorOrTable) == 0 {
		return ResolveActorType("")
	}
	return ResolveActorType(actorOrTable[0])
}

func matchConfiguredKid(configured, headerKid string) error {
	if configured == "" {
		return nil
	}
	if headerKid != "" && headerKid != configured {
		return errors.New("key id mismatch")
	}
	return nil
}

// normalizeSigningAlgorithm returns RS256 or ES256. Empty defaults to RS256. HS256 and other algs are rejected.
func normalizeSigningAlgorithm(raw string) (string, error) {
	alg := strings.ToUpper(strings.TrimSpace(raw))
	if alg == "" {
		return "RS256", nil
	}
	if alg == "HS256" {
		return "", errors.New("JWT HS256 is no longer supported; use RS256 or ES256")
	}
	if alg != "RS256" && alg != "ES256" {
		return "", fmt.Errorf("JWT signing algorithm must be RS256 or ES256; got %q", alg)
	}
	return alg, nil
}

func (m *Manager) loadFromEnvOrFiles(conf *config.Configuration, alg string) error {
	switch alg {
	case "RS256":
		privateKey, err := getPrivateKey(conf)
		if err != nil {
			return err
		}
		rsaPrivate, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return errors.New("JWT RS256 private key is not RSA")
		}
		publicKey, err := getPublicKey(conf)
		if err != nil {
			return err
		}
		rsaPublic, ok := publicKey.(*rsa.PublicKey)
		if !ok {
			return errors.New("JWT RS256 public key is not RSA")
		}
		m.signingMethod = jwt.SigningMethodRS256
		m.signKey = rsaPrivate
		m.verifyKey = rsaPublic
	case "ES256":
		privateKey, err := getPrivateKey(conf)
		if err != nil {
			return err
		}
		ecPrivate, ok := privateKey.(*ecdsa.PrivateKey)
		if !ok {
			return errors.New("JWT ES256 private key is not ECDSA")
		}
		publicKey, err := getPublicKey(conf)
		if err != nil {
			return err
		}
		ecPublic, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return errors.New("JWT ES256 public key is not ECDSA")
		}
		m.signingMethod = jwt.SigningMethodES256
		m.signKey = ecPrivate
		m.verifyKey = ecPublic
	default:
		return fmt.Errorf("JWT signing algorithm must be RS256 or ES256; got %q", alg)
	}
	return nil
}

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
// actorOrTable is an actor type or table name; empty defaults to service. See ResolveActorType.
func (m *Manager) GenerateToken(id uuid.UUID, actorOrTable ...string) (string, error) {
	return m.GenerateTokenWithExpiry(id, m.accessExpiry, actorOrTable...)
}

// GenerateTokenWithExpiry issues a signed JWT with the given UUID and custom expiry (token_type: "access").
func (m *Manager) GenerateTokenWithExpiry(id uuid.UUID, expiry time.Duration, actorOrTable ...string) (string, error) {
	claims := Claims{
		UUID:             id.String(),
		RegisteredClaims: m.buildRegisteredClaims(id.String(), expiry),
		ActorType:        resolveActorArg(actorOrTable...),
		TokenType:        TokenTypeAccess,
	}
	return m.signToken(claims)
}

// GenerateRefreshToken issues a signed JWT with refresh token expiry (token_type: "refresh").
func (m *Manager) GenerateRefreshToken(id uuid.UUID, actorOrTable ...string) (string, error) {
	claims := Claims{
		UUID:             id.String(),
		RegisteredClaims: m.buildRegisteredClaims(id.String(), m.refreshExpiry),
		ActorType:        resolveActorArg(actorOrTable...),
		TokenType:        TokenTypeRefresh,
	}
	return m.signToken(claims)
}

// GenerateImpersonationToken issues a short-lived JWT for admin-as-user (token_type: "impersonation").
// TTL is clamped to max 30 minutes. ActorType is always "user" (impersonated subject).
func (m *Manager) GenerateImpersonationToken(adminID uuid.UUID, adminRole string, targetUserID uuid.UUID, requestedTTL time.Duration) (string, error) {
	return m.GenerateImpersonation(ImpersonationParams{
		AdminID: adminID, AdminRole: adminRole, TargetUserID: targetUserID, TTL: requestedTTL,
	})
}

func (m *Manager) GenerateImpersonation(p ImpersonationParams) (string, error) {
	maxTTL := 30 * time.Minute
	ttl := p.TTL
	if ttl <= 0 || ttl > maxTTL {
		ttl = maxTTL
	}
	claims := Claims{
		UUID:             p.TargetUserID.String(),
		RegisteredClaims: m.buildRegisteredClaims(p.TargetUserID.String(), ttl),
		ActorType:        ActorTypeUser,
		TokenType:        TokenTypeImpersonation,
		ImpersonatorID:   p.AdminID.String(),
		ImpersonatorRole: p.AdminRole,
		IsImpersonating:  true,
		OriginalSub:      p.AdminID.String(),
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
			kid, _ := token.Header["kid"].(string)
			if err := matchConfiguredKid(m.kid, kid); err != nil {
				return nil, err
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
