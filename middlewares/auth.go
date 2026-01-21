package middlewares

import (
	"strings"

	"github.com/turahe/pkg/jwt"
	"github.com/turahe/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token from Authorization header
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
