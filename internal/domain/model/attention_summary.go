package model

import (
	"errors"
	"time"
)

// AttentionSummary holds the clinical information recorded by the healthcare
// professional for one consultation.
//
// Ownership (per domain-map):
//   - This entity belongs to Medical Consultation.
//   - Only an authenticated PROFESSIONAL may create or modify it.
//   - The Intelligent Agent must never write to it.
type AttentionSummary struct {
	ID              int64                   `json:"id"`
	ConsultationID  int64                   `json:"consultationId"`
	Diagnosis       *string                 `json:"diagnosis,omitempty"`
	Recommendations string                  `json:"recommendations"`
	Medications     []MedicationInstruction `json:"medications"`
	Observations    *string                 `json:"observations,omitempty"`
	Referral        *string                 `json:"referral,omitempty"`
	FollowUp        *string                 `json:"followUp,omitempty"`
	CreatedAt       time.Time               `json:"createdAt"`
	UpdatedAt       time.Time               `json:"updatedAt"`
}

// NewAttentionSummary builds a valid attention summary for an existing
// consultation. It enforces the domain invariants of the contract:
//   - consultationID must be positive
//   - recommendations must not be empty
//   - medications must be non-nil (an empty slice is valid)
//   - every medication must pass its own validation
//
// The DB stores medications as JSONB; nil is not allowed, [] is.
func NewAttentionSummary(
	consultationID int64,
	diagnosis, recommendations string,
	medications []MedicationInstruction,
	observations, referral, followUp *string,
) (*AttentionSummary, error) {
	if consultationID <= 0 {
		return nil, errors.New("consultationID must be positive")
	}
	rec := trimOrEmpty(recommendations)
	if rec == "" {
		return nil, errors.New("recommendations are required")
	}
	if medications == nil {
		medications = []MedicationInstruction{}
	}
	now := time.Now().UTC()
	return &AttentionSummary{
		ConsultationID:  consultationID,
		Diagnosis:       clean(&diagnosis),
		Recommendations: rec,
		Medications:     medications,
		Observations:    clean(observations),
		Referral:        clean(referral),
		FollowUp:        clean(followUp),
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func trimOrEmpty(s string) string {
	t := s
	for len(t) > 0 && (t[0] == ' ' || t[0] == '\t' || t[0] == '\n' || t[0] == '\r') {
		t = t[1:]
	}
	for len(t) > 0 {
		last := t[len(t)-1]
		if last == ' ' || last == '\t' || last == '\n' || last == '\r' {
			t = t[:len(t)-1]
			continue
		}
		break
	}
	return t
}
