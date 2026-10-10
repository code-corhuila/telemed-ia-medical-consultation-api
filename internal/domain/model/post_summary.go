package model

import (
	"errors"
	"time"
)

// PostSummary is the structured summary generated after a consultation is
// completed. It references the pre-consultation summary produced by the
// agent-service and the attention summary recorded by the professional.
//
// Ownership: Medical Consultation. The PDF is not generated here
// (per ADR-010, that belongs to document-service).
type PostSummary struct {
	ID                       int64     `json:"id"`
	ConsultationID           int64     `json:"consultationId"`
	AppointmentID            int64     `json:"appointmentId"`
	PatientID                int64     `json:"patientId"`
	PreconsultationSummaryID *int64    `json:"preconsultationSummaryId,omitempty"`
	AttentionSummaryID       *int64    `json:"attentionSummaryId,omitempty"`
	GeneratedAt              time.Time `json:"generatedAt"`
}

// NewPostSummary builds a valid post-summary for an already-completed
// consultation. It requires:
//   - consultationID, appointmentID, patientID > 0
//   - the consultation to be COMPLETED (checked by the caller use case)
func NewPostSummary(
	consultationID, appointmentID, patientID int64,
	preconsultationSummaryID, attentionSummaryID *int64,
) (*PostSummary, error) {
	if consultationID <= 0 || appointmentID <= 0 || patientID <= 0 {
		return nil, errors.New("consultationID, appointmentID and patientID are required")
	}
	return &PostSummary{
		ConsultationID:           consultationID,
		AppointmentID:            appointmentID,
		PatientID:                patientID,
		PreconsultationSummaryID: preconsultationSummaryID,
		AttentionSummaryID:       attentionSummaryID,
		GeneratedAt:              time.Now().UTC(),
	}, nil
}
