package middlewares

import (
	"strings"

	"github.com/turahe/pkg/jwt"
	"github.com/turahe/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware returns a Gin middleware that validates the Authorization: Bearer <token> header,
// calls jwt.ValidateToken, and sets "user_id" (claims.UUID) in the Gin context. On missing or invalid
// token it aborts with 401 and does not call Next.
//
// Requires jwt.Init() and config.Server.Secret to be set before use.
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			response.UnauthorizedError(ctx, "Authorization header is required")
			ctx.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.UnauthorizedError(ctx, "Invalid authorization header format")
			ctx.Abort()
			return
		}

		token := parts[1]
		claims, err := jwt.ValidateToken(token)
		if err != nil {
			response.UnauthorizedError(ctx, "Invalid or expired token")
			ctx.Abort()
			return
		}

		// Store claims in context for use in handlers
		ctx.Set("user_id", claims.UUID)

		ctx.Next()
	}
}
