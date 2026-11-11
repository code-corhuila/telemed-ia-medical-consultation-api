package model

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ConsultationCompletedEventType is the event published when the attention is recorded (DEC-007).
const ConsultationCompletedEventType = "ConsultationCompleted"

// NewConsultationCompletedEvent builds the outbox row of ConsultationCompleted (DEC-003, DEC-007).
//
// The envelope follows ADR-011 and the payload carries identifiers only. The event id is a
// UUID v5 of the consultation id and the event type, so one consultation has one event id.
// postSummaryId is added when the post-summary was created with the attention (DEC-008).
func NewConsultationCompletedEvent(c *Consultation, postSummaryID, correlationID string, occurredAt time.Time) (OutboxEvent, error) {
	if c == nil || c.Status != StatusCompleted || c.CompletedAt == nil || c.ID == "" {
		return OutboxEvent{}, errors.New("ConsultationCompleted needs a completed consultation")
	}

	payload := map[string]string{
		"consultationId": c.ID,
		"appointmentId":  c.AppointmentID,
		"patientId":      c.PatientID,
		"professionalId": c.ProfessionalID,
		"completedAt":    c.CompletedAt.UTC().Format(time.RFC3339),
	}
	if postSummaryID != "" {
		payload["postSummaryId"] = postSummaryID
	}
	return newOutboxEvent(c.ID, ConsultationCompletedEventType, payload, correlationID, occurredAt)
}

// newOutboxEvent wraps a payload in the ADR-011 envelope. The event id is a UUID v5 of the
// aggregate id and the event type, so the same aggregate always gives the same event id.
func newOutboxEvent(aggregateID, eventType string, payload any, correlationID string, occurredAt time.Time) (OutboxEvent, error) {
	eventID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(aggregateID+":"+eventType)).String()
	envelope, err := json.Marshal(map[string]any{
		"eventId":       eventID,
		"eventType":     eventType,
		"occurredAt":    occurredAt.UTC().Format(time.RFC3339),
		"correlationId": correlationID,
		"version":       1,
		"payload":       payload,
	})
	if err != nil {
		return OutboxEvent{}, err
	}

	return OutboxEvent{
		ID:            eventID,
		EventType:     eventType,
		Payload:       envelope,
		CorrelationID: correlationID,
		Status:        "PENDING",
		CreatedAt:     occurredAt.UTC(),
	}, nil
}
