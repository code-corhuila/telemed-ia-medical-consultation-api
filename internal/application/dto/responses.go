package dto

import "time"

// AttentionSummaryResponse is what the API returns for an attention summary.
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
type PostSummaryResponse struct {
	ID                       int64     `json:"id"`
	ConsultationID           int64     `json:"consultationId"`
	AppointmentID            int64     `json:"appointmentId"`
	PatientID                int64     `json:"patientId"`
	PreconsultationSummaryID *int64    `json:"preconsultationSummaryId,omitempty"`
	AttentionSummaryID       *int64    `json:"attentionSummaryId,omitempty"`
	GeneratedAt              time.Time `json:"generatedAt"`
}
