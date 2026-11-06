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
//
// Cross-domain IDs are UUID strings; internal IDs stay int64.
type PostSummary struct {
	ID                       string    `json:"id"`
	ConsultationID           string    `json:"consultationId"`
	AppointmentID            string    `json:"appointmentId"`
	PatientID                string    `json:"patientId"`
	PreconsultationSummaryID *string   `json:"preconsultationSummaryId,omitempty"`
	AttentionSummaryID       *string   `json:"attentionSummaryId,omitempty"`
	GeneratedAt              time.Time `json:"generatedAt"`
}

// NewPostSummary builds a valid post-summary for an already-completed
// consultation. It requires:
//   - consultationID > 0
//   - appointmentID and patientID are UUID strings
//   - preconsultationSummaryID, when present, is a UUID string
//   - the consultation to be COMPLETED (checked by the caller use case)
func NewPostSummary(
	consultationID string,
	appointmentID, patientID string,
	preconsultationSummaryID *string,
	attentionSummaryID *string,
) (*PostSummary, error) {
	if !isUUID(consultationID) {
		return nil, errors.New("consultationID must be a UUID")
	}
	if !isUUID(appointmentID) || !isUUID(patientID) {
		return nil, errors.New("appointmentID and patientID must be UUIDs")
	}
	if preconsultationSummaryID != nil && !isUUID(*preconsultationSummaryID) {
		return nil, errors.New("preconsultationSummaryID must be a UUID when present")
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
