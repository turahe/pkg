package middlewares

import (
	"github.com/gin-gonic/gin"
)

// setupRouter creates a new Gin router in test mode for testing middleware
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}
