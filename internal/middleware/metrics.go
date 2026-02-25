package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsRecord holds per-request observability data for export to any
// metrics backend (Prometheus, Datadog, etc.). Attach a real collector
// via WithMetricsCollector before deploying to production.
type MetricsRecord struct {
	Method     string
	Path       string
	StatusCode int
	DurationMs float64
}

// MetricsCollector is implemented by any metrics backend.
type MetricsCollector interface {
	Observe(r MetricsRecord)
}

// noopCollector silently discards all metrics (default until wired up).
type noopCollector struct{}

func (n *noopCollector) Observe(_ MetricsRecord) {}

var activeCollector MetricsCollector = &noopCollector{}

// WithMetricsCollector sets the global collector used by the middleware.
func WithMetricsCollector(c MetricsCollector) {
	activeCollector = c
}

// Metrics records HTTP method, path, status, and latency for every request.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		activeCollector.Observe(MetricsRecord{
			Method:     c.Request.Method,
			Path:       c.FullPath(),
			StatusCode: c.Writer.Status(),
			DurationMs: float64(time.Since(start).Milliseconds()),
		})
	}
}

// StatusClass returns the HTTP status class string ("2xx", "4xx", etc.).
func StatusClass(code int) string {
	return strconv.Itoa(code/100) + "xx"
}
