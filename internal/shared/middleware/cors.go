package middleware

import (
	"github.com/gin-gonic/gin"
)

const (
	corsOrigin      = "Access-Control-Allow-Origin"
	corsCredentials = "Access-Control-Allow-Credentials"
	corsHeaders     = "Access-Control-Allow-Headers"
	corsMethods     = "Access-Control-Allow-Methods"
	allowedHeaders  = "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
	allowedMethods  = "POST, OPTIONS, GET, PUT, DELETE, PATCH"
)

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set(corsOrigin, "*")
		c.Writer.Header().Set(corsCredentials, "true")
		c.Writer.Header().Set(corsHeaders, allowedHeaders)
		c.Writer.Header().Set(corsMethods, allowedMethods)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
