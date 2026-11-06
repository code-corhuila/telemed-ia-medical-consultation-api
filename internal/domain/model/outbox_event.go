package model

import (
	"encoding/json"
	"time"
)

type OutboxEvent struct {
	ID            string
	EventType     string
	Payload       json.RawMessage
	CorrelationID string
	Status        string
	CreatedAt     time.Time
	PublishedAt   *time.Time
}
