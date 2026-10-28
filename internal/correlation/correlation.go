// Package correlation provides helpers to propagate a request-scoped
// correlation id (X-Correlation-Id) and the W3C trace context
// (traceparent) through context.
package correlation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
)

type ctxKey struct{}

type traceKey struct {
	traceID string
	spanID  string
}

// New returns a new random correlation id (32 hex chars).
func New() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// WithContext stores the correlation id in the context.
func WithContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext returns the correlation id stored in the context.
func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}

// WithTraceparent stores the W3C trace id and span id in the context.
func WithTraceparent(ctx context.Context, traceID, spanID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, traceKey{traceID: traceID, spanID: spanID})
}

// TraceIDFromContext returns the W3C trace id, or empty.
func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(traceKey); ok {
		return v.traceID
	}
	return ""
}

// SpanIDFromContext returns the W3C span id, or empty.
func SpanIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(traceKey); ok {
		return v.spanID
	}
	return ""
}

// ParseTraceparent validates a W3C traceparent header.
// Returns (traceID, spanID, true) if valid.
//
// Format: version "-" trace-id "-" parent-id "-" trace-flags
//   - version: 2 hex chars (we only accept "00")
//   - trace-id: 32 hex chars (not all zeros)
//   - parent-id: 16 hex chars (not all zeros)
//   - trace-flags: 2 hex chars
func ParseTraceparent(h string) (string, string, bool) {
	parts := strings.Split(h, "-")
	if len(parts) != 4 {
		return "", "", false
	}
	version, traceID, spanID, flags := parts[0], parts[1], parts[2], parts[3]
	if len(version) != 2 || len(traceID) != 32 || len(spanID) != 16 || len(flags) != 2 {
		return "", "", false
	}
	if version != "00" {
		return "", "", false
	}
	if isAllZeros(traceID) || isAllZeros(spanID) {
		return "", "", false
	}
	if !isHex(traceID) || !isHex(spanID) || !isHex(flags) {
		return "", "", false
	}
	return traceID, spanID, true
}

func isAllZeros(s string) bool {
	for _, c := range s {
		if c != '0' {
			return false
		}
	}
	return true
}

func isHex(s string) bool {
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}
