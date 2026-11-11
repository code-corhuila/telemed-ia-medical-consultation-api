package model

import (
	"errors"
	"time"
)

// ConsultationCompletedEventType is the event published when the attention is recorded (DEC-007).
const ConsultationCompletedEventType = "ConsultationCompleted"

// NewConsultationCompletedEvent builds the outbox row of ConsultationCompleted (DEC-003, DEC-007).
func NewConsultationCompletedEvent(c *Consultation, correlationID string, occurredAt time.Time) (OutboxEvent, error) {
	return OutboxEvent{}, errors.New("not implemented")
}
