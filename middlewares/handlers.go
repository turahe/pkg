package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/turahe/pkg/logger"
	"github.com/turahe/pkg/response"

	"github.com/gin-gonic/gin"
)

// NoMethodHandler returns a handler that responds with 405 Method Not Allowed and a standardized JSON body.
func NoMethodHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		response.FailWithDetailed(ctx, http.StatusMethodNotAllowed, response.ServiceCodeCommon, response.CaseCodeOperationNotAllowed, nil, "Method Not Allowed")
	}
}

// NoRouteHandler returns a handler that responds with 404 and a standardized JSON body for unmatched routes.
func NoRouteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		response.NotFoundError(ctx, response.ServiceCodeCommon, response.CaseCodeRouteNotFound, "The processing function of the request route was not found")
	}
}

// RecoveryHandler is a Gin handler that recovers from panics, logs the stack trace, and responds with
// a 500 JSON body. It must be registered with gin.Use(RecoveryHandler) (no parentheses; it is a HandlerFunc, not a factory).
func RecoveryHandler(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("panic: %v\n%s", err, debug.Stack())
			response.FailWithDetailed(ctx, http.StatusInternalServerError, response.ServiceCodeCommon, response.CaseCodeInternalError, nil, errorToString(err))
			ctx.Abort()
		}
	}()
	ctx.Next()
}

func errorToString(r interface{}) string {
	switch v := r.(type) {
	case error:
		return v.Error()
	case string:
		return v
	default:
		return fmt.Sprint(r)
	}
}
