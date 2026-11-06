package dto

import (
	"encoding/json"
	"time"
)

// AttentionSummaryResponse is what the API returns for an attention summary.
// No cross-domain UUID fields.
type AttentionSummaryResponse struct {
	ID              string          `json:"id"`
	ConsultationID  string          `json:"consultationId"`
	Diagnosis       *string         `json:"diagnosis,omitempty"`
	Recommendations string          `json:"recommendations"`
	Medications     []MedicationDTO `json:"medications"`
	Observations    *string         `json:"observations,omitempty"`
	Referral        *string         `json:"referral,omitempty"`
	FollowUp        json.RawMessage `json:"followUp,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// PostSummaryResponse is what the API returns for a post-summary.
// Cross-domain IDs are UUID strings.
type PostSummaryResponse struct {
	ID                       string    `json:"id"`
	ConsultationID           string    `json:"consultationId"`
	AppointmentID            string    `json:"appointmentId"`
	PatientID                string    `json:"patientId"`
	PreconsultationSummaryID *string   `json:"preconsultationSummaryId,omitempty"`
	AttentionSummaryID       *string   `json:"attentionSummaryId,omitempty"`
	GeneratedAt              time.Time `json:"generatedAt"`
}
