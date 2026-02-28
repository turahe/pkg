package middlewares

import (
	"strings"

	"github.com/turahe/pkg/jwt"
	"github.com/turahe/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware returns a Gin middleware that validates the Authorization: Bearer <token> header
// using the given JWT verifier (Manager or Verifier), and sets identity information in the Gin context.
// Pass a *jwt.Manager (from jwt.NewManager) or *jwt.Verifier (from jwt.NewVerifier) for verification-only services.
// Verifier must not be nil.
func AuthMiddleware(verifier jwt.TokenVerifier) gin.HandlerFunc {
	if verifier == nil {
		panic("jwt.TokenVerifier is required for AuthMiddleware")
	}
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			response.UnauthorizedError(ctx, "Authorization header is required")
			ctx.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.UnauthorizedError(ctx, "Invalid authorization header format")
			ctx.Abort()
			return
		}

		claims, err := verifier.ValidateToken(parts[1])
		if err != nil {
			response.UnauthorizedError(ctx, "Invalid or expired token")
			ctx.Abort()
			return
		}

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
