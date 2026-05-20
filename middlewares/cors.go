package middlewares

import (
	"strings"

	"github.com/turahe/pkg/config"

	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that sets Access-Control-* headers from config. If Cors.Global is true
// it sets Allow-Origin to "*"; otherwise it reflects the request Origin when it matches Cors.Frontend
// or an entry in Cors.Ips. Responds to OPTIONS with 204.
func CORS() gin.HandlerFunc {
	conf := config.GetConfig()

	return func(ctx *gin.Context) {
		if conf.Cors.Global {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin := corsAllowOrigin(conf.Cors, ctx.Request.Header.Get("Origin")); origin != "" {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-2FA-Code")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()
	}
}

func corsAllowOrigin(cors config.CorsConfiguration, requestOrigin string) string {
	if requestOrigin == "" {
		return ""
	}
	for _, allowed := range corsAllowedOrigins(cors) {
		if allowed == requestOrigin {
			return requestOrigin
		}
	}
	return ""
}

func corsAllowedOrigins(cors config.CorsConfiguration) []string {
	var origins []string
	if cors.Frontend != "" {
		origins = append(origins, strings.TrimSpace(cors.Frontend))
	}
	if cors.Ips == "" {
		return origins
	}
	for _, entry := range strings.Split(cors.Ips, ",") {
		if o := strings.TrimSpace(entry); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}
