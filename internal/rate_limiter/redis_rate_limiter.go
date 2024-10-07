package rate_limiter

import (
	"context"
	"ip2country-service/config"
	"net"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisRateLimiter struct {
	client   *redis.Client
	limit    int
	duration time.Duration
}

func NewRedisRateLimiter(cfg *config.Config) *RedisRateLimiter {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &RedisRateLimiter{
		client:   client,
		limit:    cfg.RateLimit,
		duration: time.Second, // 1 second rate limiting window
	}
}

func (rl *RedisRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		key := "rate_limit:" + ip

		// Increment the request count for the current IP address
		count, err := rl.client.Incr(ctx, key).Result()
		if err != nil {
			http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Set the expiration time for the rate limit key if this is the first request
		if count == 1 {
			rl.client.Expire(ctx, key, rl.duration)
		}

		// Check if the request count exceeds the limit
		if count > int64(rl.limit) {
			http.Error(w, `{"error": "Rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (r *RedisRateLimiter) CheckConnection(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	return err
}
