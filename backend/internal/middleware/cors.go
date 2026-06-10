package middleware

import (
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// DomainChecker returns true if the given hostname is a valid custom domain.
type DomainChecker func(hostname string) bool

func CORS(allowedOrigins string, domainChecker DomainChecker) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")
	originSet := make(map[string]bool, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o == "*" {
			log.Fatal("ALLOWED_ORIGINS must not contain '*' when credentials are enabled; specify exact origins")
		}
		originSet[o] = true
	}

	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if originSet[origin] {
				return true
			}
			if domainChecker != nil {
				parsed, err := url.Parse(origin)
				if err == nil && parsed.Hostname() != "" {
					return domainChecker(parsed.Hostname())
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
