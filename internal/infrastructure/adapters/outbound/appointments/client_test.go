package appointments_test

import (
	"context"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/correlation"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/infrastructure/adapters/outbound/appointments"
)

const appointmentID = "a1111111-1111-4111-8111-111111111111"

func TestClient_ReadsTheAppointmentRelayingTheCallersToken(t *testing.T) {
	var gotPath, gotAuth, gotCorrelation string
	server := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		gotPath, gotAuth, gotCorrelation = r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("X-Correlation-Id")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"` + appointmentID + `","patientId":"b2222222-2222-4222-8222-222222222222",` +
			`"professionalId":"c3333333-3333-4333-8333-333333333333","status":"CONFIRMED","start":"2026-11-12T14:00:00Z"}`))
	}))
	defer server.Close()
	ctx := correlation.WithContext(context.Background(), "corr-123")

	got, err := appointments.NewClient(server.URL, time.Second).FindAppointment(ctx, appointmentID, "caller-token")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/appointments/"+appointmentID || gotAuth != "Bearer caller-token" || gotCorrelation != "corr-123" {
		t.Fatalf("unexpected request: path=%q auth=%q correlation=%q", gotPath, gotAuth, gotCorrelation)
	}
	if got.ID != appointmentID || got.PatientID != "b2222222-2222-4222-8222-222222222222" ||
		got.ProfessionalID != "c3333333-3333-4333-8333-333333333333" || got.Status != "CONFIRMED" {
		t.Fatalf("unexpected appointment %+v", got)
	}
}

func TestClient_MapsTheAppointmentServiceAnswers(t *testing.T) {
	cases := map[int]error{
		nethttp.StatusNotFound:            exception.ErrAppointmentNotFound,
		nethttp.StatusForbidden:           exception.ErrForbidden,
		nethttp.StatusInternalServerError: exception.ErrAppointmentServiceUnavailable,
		nethttp.StatusBadGateway:          exception.ErrAppointmentServiceUnavailable,
	}
	for status, want := range cases {
		server := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
			w.WriteHeader(status)
		}))

		_, err := appointments.NewClient(server.URL, time.Second).FindAppointment(context.Background(), appointmentID, "t")
		server.Close()

		if !errors.Is(err, want) {
			t.Fatalf("%d: expected %v, got %v", status, want, err)
		}
	}
}

func TestClient_TreatsATimeoutOrAnUnreachableServiceAsUnavailable(t *testing.T) {
	slow := httptest.NewServer(nethttp.HandlerFunc(func(_ nethttp.ResponseWriter, r *nethttp.Request) {
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
		}
	}))
	defer slow.Close()

	_, err := appointments.NewClient(slow.URL, 50*time.Millisecond).FindAppointment(context.Background(), appointmentID, "t")
	if !errors.Is(err, exception.ErrAppointmentServiceUnavailable) {
		t.Fatalf("timeout: expected ErrAppointmentServiceUnavailable, got %v", err)
	}

	_, err = appointments.NewClient("http://127.0.0.1:1", time.Second).FindAppointment(context.Background(), appointmentID, "t")
	if !errors.Is(err, exception.ErrAppointmentServiceUnavailable) {
		t.Fatalf("unreachable: expected ErrAppointmentServiceUnavailable, got %v", err)
	}
}
