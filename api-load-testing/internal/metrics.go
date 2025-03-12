package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Define Prometheus metrics
	requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_client_requests_total",
		Help: "Total number of HTTP requests made",
	},
		[]string{"server_addr"},
	)

	requestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_client_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // From 1ms to ~1s
	})

	requestErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_client_request_errors_total",
		Help: "Total number of HTTP request errors",
	},
		[]string{"server_addr", "status_code"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestErrors)
}
