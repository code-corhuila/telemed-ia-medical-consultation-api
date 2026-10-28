package http

import (
	"context"
	nethttp "net/http"
	"strings"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/correlation"
)

// Correlation reads X-Correlation-Id from the request (or generates one),
// parses the W3C traceparent header if present, stores both in the context,
// and echoes them in the response.
//
// traceparent format (W3C Trace Context, Level 1):
//
//	version "-" trace-id "-" parent-id "-" trace-flags
//	00-<32 hex>-<16 hex>-<2 hex>
//
// Example:
//
//	00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
func Correlation(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		id := r.Header.Get("X-Correlation-Id")
		if id == "" {
			id = correlation.New()
		}

		ctx := correlation.WithContext(r.Context(), id)

		if tp := strings.TrimSpace(r.Header.Get("traceparent")); tp != "" {
			if traceID, spanID, ok := correlation.ParseTraceparent(tp); ok {
				ctx = correlation.WithTraceparent(ctx, traceID, spanID)
				// Echo traceparent so the client can propagate it to the next hop.
				w.Header().Set("traceparent", tp)
			}
		}

		w.Header().Set("X-Correlation-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// traceFrom returns the correlation id stored in the context, or empty.
func traceFrom(ctx context.Context) string {
	return correlation.FromContext(ctx)
}
