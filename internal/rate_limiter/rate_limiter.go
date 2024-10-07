package rate_limiter

import (
	"fmt"
	"ip2country-service/config"
	"net/http"
)

type RateLimiter interface {
	Limit(next http.Handler) http.Handler
}

func NewRateLimiter(cfg *config.Config) (RateLimiter, error) {
	switch cfg.RateLimiterType {
	case "local":
		return NewLocalRateLimiter(cfg.RateLimit, cfg.RateCapacity, cfg.RateJitter), nil
	case "redis":
		return NewRedisRateLimiter(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported rate limiter type: %s", cfg.RateLimiterType)
	}
}
