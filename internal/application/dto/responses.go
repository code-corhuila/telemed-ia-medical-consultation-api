package dto

import "time"

// AttentionSummaryResponse is what the API returns for an attention summary.
// No cross-domain UUID fields.
type AttentionSummaryResponse struct {
	ID              int64           `json:"id"`
	ConsultationID  int64           `json:"consultationId"`
	Diagnosis       *string         `json:"diagnosis,omitempty"`
	Recommendations string          `json:"recommendations"`
	Medications     []MedicationDTO `json:"medications"`
	Observations    *string         `json:"observations,omitempty"`
	Referral        *string         `json:"referral,omitempty"`
	FollowUp        *string         `json:"followUp,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// PostSummaryResponse is what the API returns for a post-summary.
// Cross-domain IDs are UUID strings.
type PostSummaryResponse struct {
	ID                       int64     `json:"id"`
	ConsultationID           int64     `json:"consultationId"`
	AppointmentID            string    `json:"appointmentId"`
	PatientID                string    `json:"patientId"`
	PreconsultationSummaryID *string   `json:"preconsultationSummaryId,omitempty"`
	AttentionSummaryID       *int64    `json:"attentionSummaryId,omitempty"`
	GeneratedAt              time.Time `json:"generatedAt"`
}
