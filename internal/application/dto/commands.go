package dto

import "encoding/json"

// RecordAttentionCommand is the input for the RecordAttention use case.
// ConsultationID is a UUID string (owned by this domain).
// Medications is always an array (may be empty).
// FollowUp is structured JSON when present (json.RawMessage so the
// application layer doesn't need to know the schema).
type RecordAttentionCommand struct {
	ConsultationID  string          `json:"-"`
	Diagnosis       *string         `json:"diagnosis,omitempty"`
	Recommendations string          `json:"recommendations"`
	Medications     []MedicationDTO `json:"medications"`
	Observations    *string         `json:"observations,omitempty"`
	Referral        *string         `json:"referral,omitempty"`
	FollowUp        json.RawMessage `json:"followUp,omitempty"`
}

// GeneratePostSummaryCommand is the input for the GeneratePostSummary use case.
// PreconsultationSummaryID is a UUID string (owned by agent-service).
// CorrelationID is the request-scoped X-Correlation-Id; it is persisted in the
// outbox row and in the event envelope so that the same trace ID flows from the
// HTTP request to the consumer.
type GeneratePostSummaryCommand struct {
	ConsultationID           string  `json:"-"`
	PreconsultationSummaryID *string `json:"preconsultationSummaryId,omitempty"`
	CorrelationID            string  `json:"-"`
}
