package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PublicCORS returns a Gin middleware that sets permissive CORS headers for
// fully public endpoints. No credentials are allowed. This is intended for
// endpoints that are consumed from custom domains where the origin is not
// known in advance.
func PublicCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
