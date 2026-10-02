package model

import (
	"testing"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
)

func TestNewConsultation_RejectsNonPositiveIDs(t *testing.T) {
	cases := []struct {
		name          string
		appointmentID int64
		patientID     int64
		professionalID int64
	}{
		{"zero appointment", 0, 1, 1},
		{"negative patient", 1, -1, 1},
		{"negative professional", 1, 1, -1},
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
	c, err := NewConsultation(1, 2, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != StatusInProgress {
		t.Fatalf("expected IN_PROGRESS, got %s", c.Status)
	}
	if c.CompletedAt != nil {
		t.Fatalf("CompletedAt should be nil")
	}
}

func TestComplete_SetsStatusAndTimestamp(t *testing.T) {
	c, _ := NewConsultation(1, 2, 3)
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
	c, _ := NewConsultation(1, 2, 3)
	_ = c.Complete()
	if err := c.Complete(); err != exception.ErrConsultationAlreadyCompleted {
		t.Fatalf("expected ErrConsultationAlreadyCompleted, got %v", err)
	}
}

func TestEnsureInProgress_RejectsCompleted(t *testing.T) {
	c, _ := NewConsultation(1, 2, 3)
	_ = c.Complete()
	if err := c.EnsureInProgress(); err == nil {
		t.Fatalf("expected error on completed consultation")
	}
}
