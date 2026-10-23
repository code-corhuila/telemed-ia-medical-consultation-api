package dto

import "encoding/json"

// RecordAttentionCommand is the input for the RecordAttention use case.
type RecordAttentionCommand struct {
	ConsultationID  string
	Diagnosis       *string
	Recommendations string
	Medications     []MedicationDTO
	Observations    *string
	Referral        *string
	FollowUp        json.RawMessage
}

// GeneratePostSummaryCommand is the input for the GeneratePostSummary use case.
// PreconsultationSummaryID is a UUID string (owned by agent-service).
type GeneratePostSummaryCommand struct {
	ConsultationID           string
	PreconsultationSummaryID *string
}
