package rate_limiter_test

import (
	"ip2country-service/config"
	"ip2country-service/internal/rate_limiter"
	"testing"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name            string
		rateLimiterType string
		expectError     bool
	}{
		{"Local Rate Limiter", "local", false},
		{"Redis Rate Limiter", "redis", false},
		{"Unsupported Rate Limiter", "unsupported", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{RateLimiterType: tt.rateLimiterType}
			_, err := rate_limiter.NewRateLimiter(cfg)
			if (err != nil) != tt.expectError {
				t.Errorf("NewRateLimiter() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
