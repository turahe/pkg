package middlewares

import (
	"github.com/turahe/pkg/config"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	conf := config.GetConfig()

	return func(ctx *gin.Context) {
		if conf.Cors.Global {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", conf.Cors.Ips)
		}
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()
	}
}
