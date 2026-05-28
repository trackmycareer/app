package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/bhcloudlabs/trackmy-career/pkg/response"
)

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
