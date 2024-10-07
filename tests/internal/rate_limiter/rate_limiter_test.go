package rate_limiter_test

import (
	"ip2country-service/config"
	"ip2country-service/internal/rate_limiter"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name            string
		rateLimiterType string
		rateLimit       float64
		rateCapacity    float64
		rateJitter      time.Duration
		expectError     bool
	}{
		{"Local Rate Limiter", "local", 1, 5, 100 * time.Millisecond, false},
		{"Redis Rate Limiter", "redis", 1, 5, 100 * time.Millisecond, false},
		{"Unsupported Rate Limiter", "unsupported", 1, 5, 100 * time.Millisecond, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				RateLimiterType: tt.rateLimiterType,
				RateLimit:       tt.rateLimit,
				RateCapacity:    tt.rateCapacity,
				RateJitter:      tt.rateJitter,
			}
			_, err := rate_limiter.NewRateLimiter(cfg)
			if (err != nil) != tt.expectError {
				t.Errorf("NewRateLimiter() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
