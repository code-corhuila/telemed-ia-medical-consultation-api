package model

import (
	"testing"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
)

const (
	uuidA = "00000000-0000-0000-0000-000000000001"
	uuidB = "00000000-0000-0000-0000-000000000002"
	uuidC = "00000000-0000-0000-0000-000000000003"
)

func TestNewConsultation_RejectsInvalidIDs(t *testing.T) {
	cases := []struct {
		name           string
		appointmentID  string
		patientID      string
		professionalID string
	}{
		{"empty appointment", "", uuidB, uuidC},
		{"empty patient", uuidA, "", uuidC},
		{"empty professional", uuidA, uuidB, ""},
		{"not a uuid appointment", "abc", uuidB, uuidC},
		{"not a uuid patient", uuidA, "abc", uuidC},
		{"not a uuid professional", uuidA, uuidB, "abc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewConsultation(tc.appointmentID, tc.patientID, tc.professionalID)
			if err != exception.ErrInvalidConsultation {
				t.Fatalf("expected ErrInvalidConsultation, got %v", err)
			}
		})
	}
}

func TestNewConsultation_StartsInProgress(t *testing.T) {
	c, err := NewConsultation(uuidA, uuidB, uuidC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != StatusInProgress {
		t.Fatalf("expected IN_PROGRESS, got %s", c.Status)
	}
	if c.CompletedAt != nil {
		t.Fatalf("CompletedAt should be nil")
	}
	if c.AppointmentID != uuidA || c.PatientID != uuidB || c.ProfessionalID != uuidC {
		t.Fatalf("UUIDs not stored correctly: %+v", c)
	}
}

func TestComplete_SetsStatusAndTimestamp(t *testing.T) {
	c, _ := NewConsultation(uuidA, uuidB, uuidC)
	if err := c.Complete(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != StatusCompleted {
		t.Fatalf("expected COMPLETED, got %s", c.Status)
	}
	if c.CompletedAt == nil {
		t.Fatalf("CompletedAt should be set")
	}
}

func TestComplete_IsNotIdempotent(t *testing.T) {
	c, _ := NewConsultation(uuidA, uuidB, uuidC)
	_ = c.Complete()
	if err := c.Complete(); err != exception.ErrConsultationAlreadyCompleted {
		t.Fatalf("expected ErrConsultationAlreadyCompleted, got %v", err)
	}
}

func TestEnsureInProgress_RejectsCompleted(t *testing.T) {
	c, _ := NewConsultation(uuidA, uuidB, uuidC)
	_ = c.Complete()
	if err := c.EnsureInProgress(); err == nil {
		t.Fatalf("expected error on completed consultation")
	}
}
