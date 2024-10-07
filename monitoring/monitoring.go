package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)

	RateLimitExceeded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_rate_limit_exceeded_total",
			Help: "Total number of HTTP requests that exceeded the rate limit",
		},
		[]string{"path"},
	)
)

func init() {
	prometheus.MustRegister(RequestsTotal, RequestDuration, RateLimitExceeded)
}
