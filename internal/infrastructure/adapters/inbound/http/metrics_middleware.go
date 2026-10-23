package http

import (
	nethttp "net/http"
	"strconv"
	"time"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/metrics"
)

// statusRecorder captura el status code para las métricas
type statusRecorder struct {
	nethttp.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// MetricsMiddleware intercepta las peticiones para medir latencia y conteo (CP3)
func MetricsMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		
		next.ServeHTTP(rec, r)
		
		duration := time.Since(start).Seconds()
		statusStr := strconv.Itoa(rec.status)

		metrics.HttpDuration.WithLabelValues(r.URL.Path, r.Method, statusStr).Observe(duration)
		metrics.HttpRequestsTotal.WithLabelValues(r.URL.Path, r.Method, statusStr).Inc()
	})
}
