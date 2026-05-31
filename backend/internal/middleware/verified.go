package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func EmailVerifiedRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		verified, exists := c.Get("email_verified")
		if !exists || !verified.(bool) {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "email verification required",
				"code":    "EMAIL_NOT_VERIFIED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
