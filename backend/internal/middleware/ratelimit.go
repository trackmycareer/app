package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"

	"github.com/trackmycareer/app/pkg/response"
)

type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimit returns middleware that enforces per-IP rate limiting.
// Each client IP gets r requests per second with a burst of b.
// Useful for protecting public endpoints (login, register) against brute-force attacks.
func IPRateLimit(r rate.Limit, b int) gin.HandlerFunc {
	var mu sync.Mutex
	limiters := make(map[string]*ipLimiter)

	// Evict stale entries every 5 minutes.
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, il := range limiters {
				if time.Since(il.lastSeen) > 10*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		il, exists := limiters[ip]
		if !exists {
			il = &ipLimiter{limiter: rate.NewLimiter(r, b)}
			limiters[ip] = il
		}
		il.lastSeen = time.Now()
		mu.Unlock()

		if !il.limiter.Allow() {
			response.TooManyRequests(c, "too many requests, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}

// SearchRateLimit returns middleware that enforces per-user rate limiting.
// Each user gets r requests per second with a burst of b.
func SearchRateLimit(r rate.Limit, b int) gin.HandlerFunc {
	var mu sync.Mutex
	limiters := make(map[uuid.UUID]*userLimiter)

	// Evict stale entries every 5 minutes.
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for uid, ul := range limiters {
				if time.Since(ul.lastSeen) > 10*time.Minute {
					delete(limiters, uid)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			c.Next()
			return
		}
		uid := userID.(uuid.UUID)

		mu.Lock()
		ul, exists := limiters[uid]
		if !exists {
			ul = &userLimiter{limiter: rate.NewLimiter(r, b)}
			limiters[uid] = ul
		}
		ul.lastSeen = time.Now()
		mu.Unlock()

		if !ul.limiter.Allow() {
			response.TooManyRequests(c, "rate limit exceeded, please try again shortly")
			c.Abort()
			return
		}

		c.Next()
	}
}
