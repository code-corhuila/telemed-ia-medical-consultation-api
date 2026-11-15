package http

import (
	nethttp "net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/metrics"
)

// uuidPathPattern matches a UUID v1-v5 anywhere in the path. We replace it
// with "{id}" so Prometheus doesn't create a new series per consultation id
// (cardinality explosion).
var uuidPathPattern = regexp.MustCompile(
	`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`,
)

// statusRecorder captures the status code for metrics.
type statusRecorder struct {
	nethttp.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// sanitizePath replaces UUIDs with "{id}" to keep cardinality bounded.
func sanitizePath(path string) string {
	return uuidPathPattern.ReplaceAllString(path, "{id}")
}

// MetricsMiddleware intercepts requests to record latency and count (CP3).
// Labels: path (masked), method, status, outcome (success/error).
func MetricsMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}

		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		statusStr := strconv.Itoa(rec.status)
		path := sanitizePath(r.URL.Path)

		outcome := "success"
		if rec.status >= 400 {
			outcome = "error"
		}

		metrics.HttpDuration.WithLabelValues(path, r.Method, statusStr, outcome).Observe(duration)
		metrics.HttpRequestsTotal.WithLabelValues(path, r.Method, statusStr, outcome).Inc()
	})
}
