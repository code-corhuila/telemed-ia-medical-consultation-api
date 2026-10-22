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
//
// Cross-domain IDs are UUID strings; the internal PK (ID) stays int64.
type Consultation struct {
	ID             int64
	AppointmentID  string
	PatientID      string
	ProfessionalID string
	Status         string
	CompletedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

// NewConsultation creates a new consultation in IN_PROGRESS state.
func NewConsultation(appointmentID, patientID, professionalID string) (*Consultation, error) {
	if !isUUID(appointmentID) || !isUUID(patientID) || !isUUID(professionalID) {
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
	id int64,
	appointmentID, patientID, professionalID string,
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
func (c *Consultation) EnsureInProgress() error {
	if c.Status != StatusInProgress {
		return errors.New("consultation is not in progress")
	}
	return nil
}

// isUUID validates the canonical 8-4-4-4-12 UUID format (any case).
// Shared by the model package; no external dependency needed.
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}
