package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/turahe/pkg/logger"
	"github.com/turahe/pkg/response"
)

// NoMethodHandler returns a handler that responds with 405 Method Not Allowed and a standardized JSON body.
func NoMethodHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		response.FailWithDetailed(
			ctx,
			http.StatusMethodNotAllowed,
			response.ServiceCodeCommon,
			response.CaseCodeOperationNotAllowed,
			nil,
			"Method Not Allowed",
		)
	}
}

// NoRouteHandler returns a handler that responds with 404 and a standardized JSON body for unmatched routes.
func NoRouteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		response.NotFoundError(ctx, response.ServiceCodeCommon, response.CaseCodeRouteNotFound, "The processing function of the request route was not found")
	}
}

// RecoveryHandler recovers from panics, logs structured CRITICAL with trace/correlation preserved, and responds with 500.
// The panic value is logged only; clients get a generic message. http.ErrAbortHandler is re-panicked so net/http
// aborts the connection as intended.
// Register with gin.Use(RecoveryHandler). Do not use for fatal errors in request context; use logger.Errorf + abort.
func RecoveryHandler(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			if err == http.ErrAbortHandler {
				panic(err)
			}
			reqCtx := ctx.Request.Context()
			logger.GetLogger().LogAttrs(reqCtx, logger.LevelCritical, "panic recovered",
				slog.String("panic", errorToString(err)),
				slog.String("stacktrace", string(debug.Stack())))
			response.FailWithDetailed(
				ctx,
				http.StatusInternalServerError,
				response.ServiceCodeCommon,
				response.CaseCodeInternalError,
				nil,
				"Internal server error",
			)
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
