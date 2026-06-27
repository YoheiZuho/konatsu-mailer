package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimiter is a simple in-memory fixed-window limiter keyed by client IP,
// used to throttle auth endpoints against brute-force / abuse.
type rateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	visitors map[string]*window
}

type window struct {
	count   int
	resetAt time.Time
}

func newRateLimiter(limit int, w time.Duration) *rateLimiter {
	rl := &rateLimiter{limit: limit, window: w, visitors: make(map[string]*window)}
	go rl.cleanupLoop()
	return rl
}

func (rl *rateLimiter) cleanupLoop() {
	t := time.NewTicker(10 * time.Minute)
	for range t.C {
		now := time.Now()
		rl.mu.Lock()
		for k, v := range rl.visitors {
			if now.After(v.resetAt) {
				delete(rl.visitors, k)
			}
		}
		rl.mu.Unlock()
	}
}

// allow reports whether a request from key is within the limit.
func (rl *rateLimiter) allow(key string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v := rl.visitors[key]
	if v == nil || now.After(v.resetAt) {
		rl.visitors[key] = &window{count: 1, resetAt: now.Add(rl.window)}
		return true
	}
	if v.count >= rl.limit {
		return false
	}
	v.count++
	return true
}

// middleware returns a Gin middleware enforcing the limit per client IP.
func (rl *rateLimiter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests,
				errorResponse("rate_limited", "リクエストが多すぎます。しばらくしてからお試しください。"))
			return
		}
		c.Next()
	}
}
