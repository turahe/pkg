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

// Claims is the JWT payload: UUID (string) plus jwt.RegisteredClaims (exp, iat, nbf).
type Claims struct {
	UUID string `json:"uuid"`
	jwt.RegisteredClaims
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

	claims := Claims{
		UUID: id.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-30 * time.Second)),
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
