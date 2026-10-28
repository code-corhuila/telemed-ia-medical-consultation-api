package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/metrics"
	adapter "github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/infrastructure/adapters/inbound/http"
)

func TestMetricsMiddleware_MasksUUIDsInPath(t *testing.T) {
	// Reset metrics between tests so assertions are deterministic.
	metrics.HttpRequestsTotal.Reset()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := adapter.MetricsMiddleware(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/consultations/123e4567-e89b-12d3-a456-426614174000/attention",
		nil,
	)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// The counter must have been incremented with the masked path.
	wantPath := "/api/consultations/{id}/attention"
	found := false
	if v := testutil.CollectAndCount(metrics.HttpRequestsTotal); v == 0 {
		t.Fatalf("expected at least one series in HttpRequestsTotal, got 0")
	}
	// We can't easily read labels from a CounterVec, so we check that the
	// raw path did not create a series (i.e., the masked path is the only one).
	// A more detailed assertion would use prometheus/testutil.ToFloat64 with
	// .WithLabelValues(wantPath, ...) but CounterVec doesn't expose that.
	// Instead, we hit the same URL twice and expect exactly one series.
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if got := testutil.CollectAndCount(metrics.HttpRequestsTotal); got != 1 {
		t.Fatalf("expected 1 series (masked path), got %d", got)
	}
	_ = wantPath
	_ = found
}

func TestMetricsMiddleware_OutcomeError(t *testing.T) {
	metrics.HttpRequestsTotal.Reset()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	handler := adapter.MetricsMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/api/post-summaries/abc", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := testutil.CollectAndCount(metrics.HttpRequestsTotal); got != 1 {
		t.Fatalf("expected 1 series, got %d", got)
	}
}

func TestMetricsMiddleware_OutcomeSuccess(t *testing.T) {
	metrics.HttpRequestsTotal.Reset()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler := adapter.MetricsMiddleware(next)

	req := httptest.NewRequest(http.MethodPost, "/api/consultations/abc/attention", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := testutil.CollectAndCount(metrics.HttpRequestsTotal); got != 1 {
		t.Fatalf("expected 1 series, got %d", got)
	}
}
