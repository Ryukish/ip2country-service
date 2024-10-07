package rate_limiter

import (
	"ip2country-service/internal/rate_limiter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocalRateLimiter(t *testing.T) {
	limiter := rate_limiter.NewLocalRateLimiter(1)

	handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()

	// First request should pass
	handler.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Result().StatusCode)
	}

	// Second request should be rate limited
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status Too Many Requests, got %v", w.Result().StatusCode)
	}
}
