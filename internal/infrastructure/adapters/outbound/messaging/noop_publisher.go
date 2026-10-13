// Package messaging contains outbound messaging adapters.
package messaging

import (
	"context"
	"log/slog"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
)

// NoopPublisher logs domain events but does not deliver them anywhere.
//
// It exists because ADR-011 (RabbitMQ broker) is still Proposed. Until the
// broker is accepted, the application only depends on the EventPublisherPort
// interface and this adapter satisfies it.
type NoopPublisher struct{}

// NewNoopPublisher builds the no-op publisher.
func NewNoopPublisher() *NoopPublisher {
	return &NoopPublisher{}
}

// Publish logs the event name and returns nil.
func (p *NoopPublisher) Publish(_ context.Context, event out.DomainEvent) error {
	slog.Info("domain event (noop)", "event", event.EventName())
	return nil
}
