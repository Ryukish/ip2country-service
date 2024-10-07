package rate_limiter

import (
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type client struct {
	count     int
	timestamp time.Time
}

type LocalRateLimiter struct {
	limit    int
	mu       sync.Mutex
	requests map[string]*client
	duration time.Duration
	jitter   time.Duration
}

func NewLocalRateLimiter(limit int) *LocalRateLimiter {
	return &LocalRateLimiter{
		limit:    limit,
		requests: make(map[string]*client),
		duration: time.Second,            // 1 second rate limiting window
		jitter:   500 * time.Millisecond, // 500ms jitter
	}
}

func (rl *LocalRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		now := time.Now()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		// Check if the client exists
		c, exists := rl.requests[ip]

		if !exists || now.Sub(c.timestamp) > rl.duration+time.Duration(rand.Int63n(int64(rl.jitter))) {
			// Reset the request count if it's a new request or the time window has expired
			rl.requests[ip] = &client{count: 1, timestamp: now}
		} else {
			// Increment the request count
			c.count++
			if c.count > rl.limit {
				http.Error(w, `{"error": "Rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
