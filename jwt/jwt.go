package jwt

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turahe/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret          []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
)

// Claims is the JWT payload. It embeds jwt.RegisteredClaims (exp, iat, nbf, sub, jti)
// and adds:
//   - UUID: active user context (matches sub for new tokens)
//   - ImpersonatorID / ImpersonatorRole: admin identity when impersonating
//   - IsImpersonating: true when this token was issued for impersonation
//   - OriginalSub: original login identity (admin) when impersonating
//
// For non-impersonation tokens, only UUID and RegisteredClaims are populated;
// other fields use zero values for full backward compatibility.
type Claims struct {
	UUID string `json:"uuid"`

	jwt.RegisteredClaims

	ImpersonatorID   string `json:"impersonator_id,omitempty"`
	ImpersonatorRole string `json:"impersonator_role,omitempty"`
	IsImpersonating  bool   `json:"is_impersonating,omitempty"`
	OriginalSub      string `json:"original_sub,omitempty"`
}

// Init loads the JWT secret and token expiry from config.GetConfig() and panics if config is nil or Server.Secret is empty.
func Init() {
	conf := config.GetConfig()
	if conf != nil {
		secret := conf.Server.Secret
		if secret == "" {
			panic("JWT secret is not configured")
		}
		// Use the secret as-is (no trimming) to match exactly what's in config
		jwtSecret = []byte(secret)

		// Set default expiry if not configured
		if conf.Server.AccessTokenExpiry > 0 {
			accessTokenExpiry = time.Duration(conf.Server.AccessTokenExpiry) * time.Hour
		} else {
			accessTokenExpiry = time.Hour // Default: 1 hour
		}

		if conf.Server.RefreshTokenExpiry > 0 {
			refreshTokenExpiry = time.Duration(conf.Server.RefreshTokenExpiry) * 24 * time.Hour
		} else {
			refreshTokenExpiry = 7 * 24 * time.Hour // Default: 7 days
		}
	} else {
		panic("Config is not initialized")
	}
}

// GetSecret returns the JWT secret for debugging/verification purposes
// WARNING: Only use this for debugging. Never expose in production responses.
func GetSecret() string {
	if len(jwtSecret) == 0 {
		Init()
	}
	return string(jwtSecret)
}

// GetAccessTokenExpiry returns the access token expiry duration
func GetAccessTokenExpiry() time.Duration {
	if accessTokenExpiry == 0 {
		Init()
	}
	return accessTokenExpiry
}

// GenerateToken issues a signed JWT with the given UUID and default access token expiry.
func GenerateToken(id uuid.UUID) (string, error) {
	return GenerateTokenWithExpiry(id, accessTokenExpiry) // Default: 1 hour
}

// GenerateTokenWithExpiry issues a signed JWT with the given UUID and custom expiry duration.
func GenerateTokenWithExpiry(id uuid.UUID, expiry time.Duration) (string, error) {
	if len(jwtSecret) == 0 {
		Init()
	}

	now := time.Now()

	claims := Claims{
		UUID: id.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
			Subject:   id.String(),
			ID:        uuid.NewString(), // jti for revocation / audit
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateRefreshToken issues a signed JWT with the given UUID and refresh token expiry from config.
func GenerateRefreshToken(id uuid.UUID) (string, error) {
	if refreshTokenExpiry == 0 {
		Init()
	}
	return GenerateTokenWithExpiry(id, refreshTokenExpiry)
}

// GenerateImpersonationToken issues a signed JWT representing an administrator
// temporarily impersonating another user. The active user context (sub/UUID)
// is the impersonated user. The original admin identity is preserved in
// ImpersonatorID / ImpersonatorRole / OriginalSub and IsImpersonating is true.
//
// The requestedTTL is clamped to a safe maximum (30 minutes). If requestedTTL
// is zero or negative, the maximum is used.
func GenerateImpersonationToken(adminID uuid.UUID, adminRole string, targetUserID uuid.UUID, requestedTTL time.Duration) (string, error) {
	if len(jwtSecret) == 0 {
		Init()
	}

	maxTTL := 30 * time.Minute
	ttl := requestedTTL
	if ttl <= 0 || ttl > maxTTL {
		ttl = maxTTL
	}

	now := time.Now()

	claims := Claims{
		UUID: targetUserID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
			Subject:   targetUserID.String(),
			ID:        uuid.NewString(),
		},
		ImpersonatorID:   adminID.String(),
		ImpersonatorRole: adminRole,
		IsImpersonating:  true,
		OriginalSub:      adminID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken parses the token string, verifies signature and expiry, and returns Claims or an error.
func ValidateToken(tokenString string) (*Claims, error) {
	if len(jwtSecret) == 0 {
		Init()
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			// Return the secret key for verification
			return jwtSecret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		// jwt.WithSkipClaimsValidation(true),
		jwt.WithLeeway(30*time.Second),
	)

	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// ComparePassword returns true if plainPassword matches the bcrypt hash hashedPassword.
func ComparePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// GetCurrentUserUUID reads "user_id" from the Gin context (set by auth middleware). Supports string or uuid.UUID; returns (uuid.Nil, false) if missing or invalid.
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
