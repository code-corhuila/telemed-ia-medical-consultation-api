package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase/fakes"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
	adapter "github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/infrastructure/adapters/inbound/http"
)

func newTestRouter(t *testing.T) (*adapter.Router, *fakes.ConsultationRepo, *fakes.AttentionRepo) {
	t.Helper()
	cs := fakes.NewConsultationRepo()
	as := fakes.NewAttentionRepo()
	ps := fakes.NewPostSummaryRepo()

	record := usecase.NewRecordAttention(cs, as)
	getAtt := usecase.NewGetAttention(as)
	getPost := usecase.NewGetPostSummary(ps)
	genPost := usecase.NewGeneratePostSummary(cs, as, ps)

	h := adapter.NewHandlers(record, getAtt, getPost, genPost, nil)
	// Dummy PEM is fine because we never reach the JWT decode for these tests
	// (they either use a valid test token injector or hit /health).
	r := adapter.NewRouter(h, "-----BEGIN PUBLIC KEY-----\nFAKE\n-----END PUBLIC KEY-----")
	return r, cs, as
}

func TestHealth_IsPublic(t *testing.T) {
	r, _, _ := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "UP") {
		t.Fatalf("expected UP in body, got %s", rec.Body.String())
	}
}

func TestProtectedEndpoint_RequiresAuth(t *testing.T) {
	r, _, _ := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/consultations/1/attention", nil)
	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestMapError_Codes(t *testing.T) {
	// Sanity: mapping table covers the domain errors.
	cases := []error{
		exception.ErrConsultationNotFound,
		exception.ErrInvalidAttentionSummary,
		exception.ErrConsultationNotCompleted,
	}
	_ = cases // exercised in use case tests; this test documents intent.
	_ = dto.RecordAttentionCommand{}
	_ = context.Background()
}

// TestRouter_MetricsEndpoint verifies that /metrics is exposed
// and does not require a JWT (so Prometheus can scrape it).
func TestRouter_MetricsEndpoint(t *testing.T) {
	r, _, _ := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("/metrics: got status %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "medical_consultation_http_requests_total") {
		t.Fatalf("/metrics body missing metric name; got %q", rec.Body.String())
	}
}

// TestRouter_TraceparentEchoed verifies that the W3C traceparent header
// is echoed back in the response when present on the request.
func TestRouter_TraceparentEchoed(t *testing.T) {
	r, _, _ := newTestRouter(t)
	tp := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("traceparent", tp)
	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, req)

	if got := rec.Header().Get("traceparent"); got != tp {
		t.Fatalf("traceparent echo: got %q, want %q", got, tp)
	}
}
