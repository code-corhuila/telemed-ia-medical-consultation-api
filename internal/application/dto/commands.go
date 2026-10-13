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
type GeneratePostSummaryCommand struct {
	ConsultationID           int64
	PreconsultationSummaryID *int64
}
