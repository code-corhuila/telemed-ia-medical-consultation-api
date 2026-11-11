package model

import (
	"errors"
	"time"
)

// PostSummaryGeneratedEventType is the event published when a post-consultation summary is created.
const PostSummaryGeneratedEventType = "PostConsultationSummaryGenerated"

// NewPostSummaryGeneratedEvent builds the outbox row of PostConsultationSummaryGenerated.
func NewPostSummaryGeneratedEvent(s *PostSummary, correlationID string, occurredAt time.Time) (OutboxEvent, error) {
	return OutboxEvent{}, errors.New("not implemented")
}
