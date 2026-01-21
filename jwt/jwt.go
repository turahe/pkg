package jwt

import (
	"errors"
	"time"

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

// Claims represents JWT claims
type Claims struct {
	UUID string `json:"uuid"`
	jwt.RegisteredClaims
}

// Init initializes JWT with secret from config
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

// GenerateToken generates a JWT token for admin
func GenerateToken(id uuid.UUID) (string, error) {
	return GenerateTokenWithExpiry(id, accessTokenExpiry) // Default: 1 hour
}

// GenerateTokenWithExpiry generates a JWT token with custom expiration time
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

// GenerateRefreshToken generates a refresh token with longer expiration
func GenerateRefreshToken(id uuid.UUID) (string, error) {
	if refreshTokenExpiry == 0 {
		Init()
	}
	return GenerateTokenWithExpiry(id, refreshTokenExpiry)
}

// ValidateToken validates a JWT token and returns the claims
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

// ComparePassword compares a hashed password with a plain password
func ComparePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
