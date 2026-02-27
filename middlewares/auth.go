package middlewares

import (
	"strings"

	"github.com/turahe/pkg/jwt"
	"github.com/turahe/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware returns a Gin middleware that validates the Authorization: Bearer <token> header,
// calls jwt.ValidateToken, and sets identity information in the Gin context:
//   - "user_id":        active user context (claims.UUID / sub)
//   - "original_user_id": original login identity (admin when impersonating)
//   - "is_impersonating": bool flag indicating impersonation
//   - "impersonator_id":  admin ID when impersonating (optional)
//   - "impersonator_role": admin role when impersonating (optional)
//
// Existing tokens without impersonation fields remain fully supported: "user_id"
// is still set and "original_user_id" mirrors "user_id" with is_impersonating=false.
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

		// Store identity information in context for use in handlers and logging.
		// Active user context is always claims.UUID.
		ctx.Set("user_id", claims.UUID)

		originalID := claims.UUID
		if claims.IsImpersonating && claims.OriginalSub != "" {
			originalID = claims.OriginalSub
			ctx.Set("is_impersonating", true)
			ctx.Set("impersonator_id", claims.ImpersonatorID)
			if claims.ImpersonatorRole != "" {
				ctx.Set("impersonator_role", claims.ImpersonatorRole)
			}
		} else {
			ctx.Set("is_impersonating", false)
		}
		ctx.Set("original_user_id", originalID)

		ctx.Next()
	}
}
