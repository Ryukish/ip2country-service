package rate_limiter

import (
	"context"
	"fmt"
	"ip2country-service/config"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"ip2country-service/monitoring"

	"github.com/go-redis/redis/v8"
)

type RedisRateLimiter struct {
	client   *redis.Client
	rate     float64
	capacity float64
	jitter   time.Duration
}

func NewRedisRateLimiter(cfg *config.Config) *RedisRateLimiter {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &RedisRateLimiter{
		client:   client,
		rate:     cfg.RateLimit,
		capacity: cfg.RateCapacity,
		jitter:   cfg.RateJitter,
	}
}

func (rl *RedisRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, `{"error": "Invalid client address"}`, http.StatusBadRequest)
			return
		}
		ip := host
		key := "rate_limit:" + ip

		allowed, err := rl.allowRequest(ctx, key)
		if err != nil {
			// Log the error for debugging
			log.Printf("Rate limiter error for IP %s: %v", ip, err)
			http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		if allowed {
			jitter := time.Duration(rand.Int63n(int64(rl.jitter)))
			time.Sleep(jitter)
			next.ServeHTTP(w, r)
		} else {
			monitoring.RateLimitExceeded.WithLabelValues(r.URL.Path).Inc()
			http.Error(w, `{"error": "Rate limit exceeded"}`, http.StatusTooManyRequests)
		}
	})
}

func (rl *RedisRateLimiter) allowRequest(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMilli() // Use milliseconds
	script := `
        local tokens_key = KEYS[1]
        local timestamp_key = KEYS[2]

        local rate = tonumber(ARGV[1])
        local capacity = tonumber(ARGV[2])
        local now = tonumber(ARGV[3])
        local requested = tonumber(ARGV[4])

        local last_tokens = tonumber(redis.call("GET", tokens_key))
        if not last_tokens then
            last_tokens = capacity
        end

        local last_refreshed = tonumber(redis.call("GET", timestamp_key))
        if not last_refreshed then
            last_refreshed = now
        end

        local delta = now - last_refreshed
        if delta < 0 then
            delta = 0
        end

        local delta_seconds = delta / 1000.0 -- Convert milliseconds to seconds
        local filled_tokens = math.min(capacity, last_tokens + delta_seconds * rate)
        local allowed = filled_tokens >= requested
        local new_tokens = filled_tokens

        if allowed then
            new_tokens = filled_tokens - requested
        end

        redis.call("SETEX", tokens_key, 60, new_tokens)
        redis.call("SETEX", timestamp_key, 60, now)

        if allowed then
            return 1
        else
            return 0
        end
    `

	keys := []string{key + ":tokens", key + ":ts"}
	args := []interface{}{rl.rate, rl.capacity, now, 1}

	result, err := rl.client.Eval(ctx, script, keys, args...).Result()
	if err != nil {
		return false, err
	}

	intResult, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected script result type: %T", result)
	}

	return intResult == 1, nil
}
