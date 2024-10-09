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
		[]string{"path", "status"},
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

	IPLookupDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ip_lookup_duration_seconds",
			Help:    "Duration of IP lookups in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{},
	)

	DatabaseQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{},
	)

	AllowedFieldsUsage = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "allowed_fields_usage_total",
			Help: "Total usage count of allowed fields",
		},
		[]string{"field"},
	)
)

func init() {
	prometheus.MustRegister(RequestsTotal, RequestDuration, RateLimitExceeded, IPLookupDuration, DatabaseQueryDuration, AllowedFieldsUsage)
}
