package out

import "context"

// DomainEvent is a business fact that already occurred.
//
// Per ADR-011 the broker (RabbitMQ) is still Proposed. Until it is accepted,
// the application depends only on this interface and the infrastructure
// provides a no-op implementation that logs the event.
type DomainEvent interface {
	EventName() string
}

// EventPublisherPort is the boundary the application uses to publish events.
type EventPublisherPort interface {
	Publish(ctx context.Context, event DomainEvent) error
}
