package model

import (
	"errors"
	"time"
)

// PostSummaryGeneratedEventType is the event published when a post-consultation summary is created.
const PostSummaryGeneratedEventType = "PostConsultationSummaryGenerated"

// NewPostSummaryGeneratedEvent builds the outbox row of PostConsultationSummaryGenerated.
//
// The payload carries identifiers only. The event id is derived from the summary id, so a
// consumer can deduplicate a summary that is published again.
func NewPostSummaryGeneratedEvent(s *PostSummary, correlationID string, occurredAt time.Time) (OutboxEvent, error) {
	if s == nil || s.ID == "" {
		return OutboxEvent{}, errors.New("PostConsultationSummaryGenerated needs a saved summary")
	}
	return newOutboxEvent(s.ID, PostSummaryGeneratedEventType, map[string]string{
		"summaryId":      s.ID,
		"consultationId": s.ConsultationID,
		"appointmentId":  s.AppointmentID,
		"patientId":      s.PatientID,
		"generatedAt":    s.GeneratedAt.UTC().Format(time.RFC3339),
	}, correlationID, occurredAt)
}
