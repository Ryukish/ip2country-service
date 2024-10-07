package rate_limiter

import (
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

type client struct {
	tokens    float64
	lastCheck time.Time
}

type LocalRateLimiter struct {
	rate     float64
	capacity float64
	mu       sync.Mutex
	clients  map[string]*client
	jitter   time.Duration
}

func NewLocalRateLimiter(rate, capacity float64, jitter time.Duration) *LocalRateLimiter {
	return &LocalRateLimiter{
		rate:     rate,
		capacity: capacity,
		clients:  make(map[string]*client),
		jitter:   jitter,
	}
}

func (rl *LocalRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, `{"error": "Invalid client address"}`, http.StatusBadRequest)
			return
		}
		ip := host
		now := time.Now()

		rl.mu.Lock()
		c, exists := rl.clients[ip]
		if !exists {
			c = &client{tokens: rl.capacity, lastCheck: now}
			rl.clients[ip] = c
		}

		elapsed := now.Sub(c.lastCheck).Seconds()
		c.tokens = min(rl.capacity, c.tokens+elapsed*rl.rate)
		c.lastCheck = now

		if c.tokens >= 1 {
			c.tokens--
			rl.mu.Unlock() // Unlock before proceeding
			jitter := time.Duration(rand.Int63n(int64(rl.jitter)))
			time.Sleep(jitter)
			next.ServeHTTP(w, r)
		} else {
			rl.mu.Unlock() // Unlock before responding
			http.Error(w, `{"error": "Rate limit exceeded"}`, http.StatusTooManyRequests)
		}
	})
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
