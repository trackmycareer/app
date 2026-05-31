package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets common security headers on all API responses.
// This protects self-hosted deployments that hit the Go server directly
// without nginx in front.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Next()
	}
}
