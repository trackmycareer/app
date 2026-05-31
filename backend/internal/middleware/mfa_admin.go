package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MFARequiredForAdmin returns middleware that blocks admin users who have not
// enabled multi-factor authentication from accessing protected admin routes.
// Non-admin users pass through unaffected.
func MFARequiredForAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.Next()
			return
		}

		mfaEnabled, exists := c.Get("mfa_enabled")
		if !exists || !mfaEnabled.(bool) {
			c.JSON(http.StatusForbidden, gin.H{
				"message":            "Administrators must enable multi-factor authentication",
				"mfa_setup_required": true,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
