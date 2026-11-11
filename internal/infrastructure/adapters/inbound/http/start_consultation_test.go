package http_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase/fakes"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
	adapter "github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/infrastructure/adapters/inbound/http"
)

const (
	httpAppointmentID = "a1111111-1111-4111-8111-111111111111"
	httpPatientID     = "b2222222-2222-4222-8222-222222222222"
	httpDoctorID      = "c3333333-3333-4333-8333-333333333333"
)

type startRig struct {
	handler      http.Handler
	appointments *fakes.AppointmentReader
	key          *rsa.PrivateKey
}

func newStartRig(t *testing.T) *startRig {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	publicPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))

	cs := fakes.NewConsultationRepo()
	as := fakes.NewAttentionRepo()
	ps := fakes.NewPostSummaryRepo()
	ar := fakes.NewAppointmentReader()
	ar.Appointments[httpAppointmentID] = out.AppointmentView{
		ID: httpAppointmentID, PatientID: httpPatientID, ProfessionalID: httpDoctorID, Status: "CONFIRMED",
	}

	h := adapter.NewHandlers(
		usecase.NewRecordAttention(cs, as), usecase.NewGetAttention(as), usecase.NewGetPostSummary(ps, cs),
		usecase.NewGeneratePostSummary(cs, as, ps), cs, usecase.NewStartConsultation(cs, ar),
	)
	return &startRig{handler: adapter.NewRouter(h, publicPEM).Handler(), appointments: ar, key: key}
}

func (r *startRig) token(t *testing.T, sub, role string) string {
	t.Helper()
	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": sub, "role": role, "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString(r.key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func (r *startRig) post(t *testing.T, token, body, idempotencyKey string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/consultations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	rec := httptest.NewRecorder()
	r.handler.ServeHTTP(rec, req)
	return rec
}

type consultationEnvelope struct {
	Data struct {
		ID             string  `json:"id"`
		AppointmentID  string  `json:"appointmentId"`
		PatientID      string  `json:"patientId"`
		ProfessionalID string  `json:"professionalId"`
		Status         string  `json:"status"`
		CompletedAt    *string `json:"completedAt"`
		CreatedAt      string  `json:"createdAt"`
	} `json:"data"`
	Error   string `json:"error"`
	TraceID string `json:"traceId"`
}

func decodeConsultation(t *testing.T, rec *httptest.ResponseRecorder) consultationEnvelope {
	t.Helper()
	var env consultationEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return env
}

func TestStartConsultation_CreatesThenReturnsTheSameConsultation(t *testing.T) {
	rig := newStartRig(t)
	doctor := rig.token(t, httpDoctorID, "PROFESSIONAL")
	body := `{"appointmentId":"` + httpAppointmentID + `"}`

	first := rig.post(t, doctor, body, "key-1")
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", first.Code, first.Body.String())
	}
	created := decodeConsultation(t, first)
	if created.Data.ID == "" || created.Data.AppointmentID != httpAppointmentID || created.Data.PatientID != httpPatientID ||
		created.Data.ProfessionalID != httpDoctorID || created.Data.Status != "IN_PROGRESS" || created.Data.CompletedAt != nil {
		t.Fatalf("unexpected consultation %s", first.Body.String())
	}
	if rig.appointments.LastToken != doctor {
		t.Fatal("the caller's bearer token was not relayed to appointment-scheduling")
	}

	second := rig.post(t, doctor, body, "key-2")
	if second.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", second.Code, second.Body.String())
	}
	if decodeConsultation(t, second).Data.ID != created.Data.ID {
		t.Fatal("expected the existing consultation")
	}
}

func TestStartConsultation_RequiresAnIdempotencyKeyAndAValidBody(t *testing.T) {
	rig := newStartRig(t)
	doctor := rig.token(t, httpDoctorID, "PROFESSIONAL")

	if rec := rig.post(t, doctor, `{"appointmentId":"`+httpAppointmentID+`"}`, ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("missing key: expected 400, got %d", rec.Code)
	}
	for _, body := range []string{`{`, `{"appointmentId":"nope"}`, `{"appointmentId":"` + httpAppointmentID + `","x":1}`} {
		rec := rig.post(t, doctor, body, "key")
		if rec.Code != http.StatusBadRequest || decodeConsultation(t, rec).Error != "INVALID_PAYLOAD" {
			t.Fatalf("%s: expected 400 INVALID_PAYLOAD, got %d (%s)", body, rec.Code, rec.Body.String())
		}
	}
}

func TestStartConsultation_OnlyForProfessionals(t *testing.T) {
	rig := newStartRig(t)

	rec := rig.post(t, rig.token(t, httpPatientID, "PATIENT"), `{"appointmentId":"`+httpAppointmentID+`"}`, "key")

	if rec.Code != http.StatusForbidden || decodeConsultation(t, rec).Error != "FORBIDDEN" {
		t.Fatalf("expected 403 FORBIDDEN, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestStartConsultation_MapsTheAppointmentErrors(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{exception.ErrAppointmentNotFound, http.StatusNotFound, "APPOINTMENT_NOT_FOUND"},
		{exception.ErrForbidden, http.StatusForbidden, "FORBIDDEN"},
		{exception.ErrAppointmentServiceUnavailable, http.StatusServiceUnavailable, "APPOINTMENT_SERVICE_UNAVAILABLE"},
		{exception.ErrAppointmentNotAttendable, http.StatusConflict, "APPOINTMENT_NOT_ATTENDABLE"},
	}
	for _, tc := range cases {
		rig := newStartRig(t)
		rig.appointments.Err = tc.err

		rec := rig.post(t, rig.token(t, httpDoctorID, "PROFESSIONAL"), `{"appointmentId":"`+httpAppointmentID+`"}`, "key")

		env := decodeConsultation(t, rec)
		if rec.Code != tc.status || env.Error != tc.code || env.TraceID == "" {
			t.Fatalf("%v: expected %d %s with a traceId, got %d (%s)", tc.err, tc.status, tc.code, rec.Code, rec.Body.String())
		}
	}
}
