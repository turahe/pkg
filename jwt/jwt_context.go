package jwt

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

// GetActorType reads "actor_type" from the Gin context (set by auth middleware from JWT claims).
// Returns the actor type string (user, admin, service, system) and true when present and non-empty.
func GetActorType(ctx *gin.Context) (string, bool) {
	v, exists := ctx.Get("actor_type")
	if !exists {
		return "", false
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", false
	}
	return s, true
}
