package dto

// RecordAttentionCommand is the input for the RecordAttention use case.
type RecordAttentionCommand struct {
	ConsultationID  int64
	Diagnosis       *string
	Recommendations string
	Medications     []MedicationDTO
	Observations    *string
	Referral        *string
	FollowUp        *string
}

// GeneratePostSummaryCommand is the input for the GeneratePostSummary use case.
// PreconsultationSummaryID is a UUID string (owned by agent-service).
type GeneratePostSummaryCommand struct {
	ConsultationID           int64
	PreconsultationSummaryID *string
}
