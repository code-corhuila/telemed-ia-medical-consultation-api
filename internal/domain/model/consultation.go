package model

import (
	"errors"
	"time"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
)

// Status values of a consultation.
const (
	StatusInProgress = "IN_PROGRESS"
	StatusCompleted  = "COMPLETED"
)

// Consultation is the aggregate root of the Medical Consultation domain.
//
// A consultation is created when a healthcare professional starts recording
// medical attention for an existing appointment. It is completed when the
// professional finalizes it; a completed consultation cannot be modified
// as an active one.
type Consultation struct {
	ID             int64
	AppointmentID  int64
	PatientID      int64
	ProfessionalID int64
	Status         string
	CompletedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

// NewConsultation creates a new consultation in IN_PROGRESS state.
func NewConsultation(appointmentID, patientID, professionalID int64) (*Consultation, error) {
	if appointmentID <= 0 || patientID <= 0 || professionalID <= 0 {
		return nil, exception.ErrInvalidConsultation
	}
	now := time.Now().UTC()
	return &Consultation{
		AppointmentID:  appointmentID,
		PatientID:      patientID,
		ProfessionalID: professionalID,
		Status:         StatusInProgress,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Reconstitute rebuilds a consultation from persisted state.
// It does not apply validation rules.
func ReconstituteConsultation(
	id, appointmentID, patientID, professionalID int64,
	status string,
	completedAt *time.Time,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *Consultation {
	return &Consultation{
		ID:             id,
		AppointmentID:  appointmentID,
		PatientID:      patientID,
		ProfessionalID: professionalID,
		Status:         status,
		CompletedAt:    completedAt,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		DeletedAt:      deletedAt,
	}
}

// Complete transitions the consultation to COMPLETED.
// It is idempotent-safe at the aggregate level: completing an already
// completed consultation returns ErrConsultationAlreadyCompleted.
func (c *Consultation) Complete() error {
	if c.Status == StatusCompleted {
		return exception.ErrConsultationAlreadyCompleted
	}
	now := time.Now().UTC()
	c.Status = StatusCompleted
	c.CompletedAt = &now
	c.UpdatedAt = now
	return nil
}

// EnsureInProgress returns an error if the consultation is not active.
// Used by use cases that must reject modifications on completed consultations.
func (c *Consultation) EnsureInProgress() error {
	if c.Status != StatusInProgress {
		return errors.New("consultation is not in progress")
	}
	return nil
}
