package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/bhcloudlabs/trackmy-career/internal/auth"
	"github.com/bhcloudlabs/trackmy-career/pkg/response"
)

func AuthRequired(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorised(c, "missing authorisation header")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Unauthorised(c, "invalid authorisation header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(parts[1])
		if err != nil {
			response.Unauthorised(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("is_admin", claims.IsAdmin)
		c.Next()
	}
}
