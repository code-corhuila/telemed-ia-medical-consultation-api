package http

import (
	"context"
	nethttp "net/http"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/correlation"
)

// Correlation reads X-Correlation-Id from the request (or generates one),
// stores it in the context and echoes it in the response.
func Correlation(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		id := r.Header.Get("X-Correlation-Id")
		if id == "" {
			id = correlation.New()
		}
		w.Header().Set("X-Correlation-Id", id)
		ctx := correlation.WithContext(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// traceFrom returns the correlation id stored in the context, or empty.
func traceFrom(ctx context.Context) string {
	return correlation.FromContext(ctx)
}
