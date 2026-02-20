package middlewares

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current number of HTTP requests being processed.",
	})
)

// Metrics returns a Gin middleware that exposes Prometheus HTTP metrics:
//   - http_requests_total (counter, labels: method, path, status)
//   - http_request_duration_seconds (histogram, labels: method, path, status)
//   - http_requests_in_flight (gauge)
//
// Register the /metrics endpoint separately using promhttp.Handler().
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip the metrics endpoint itself to avoid self-instrumentation noise.
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		httpRequestsInFlight.Inc()

		c.Next()

		httpRequestsInFlight.Dec()

		status := strconv.Itoa(c.Writer.Status())
		// Use the matched route pattern (e.g. "/items/:id") not the raw path,
		// so high-cardinality IDs don't create unbounded label values.
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		elapsed := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, route, status).Observe(elapsed)
	}
}
