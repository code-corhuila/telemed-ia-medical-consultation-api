package messaging

// routingKeys maps each event type this relay publishes to its routing key: the event name
// in kebab-case (communication-patterns.md, DEC-007).
var routingKeys = map[string]string{
	"ConsultationCompleted":            "consultation-completed",
	"PostConsultationSummaryGenerated": "post-consultation-summary-generated",
}

// routingKeyFor returns the routing key of an event type. ok is false for a type this relay
// does not publish.
func routingKeyFor(eventType string) (routingKey string, ok bool) {
	routingKey, ok = routingKeys[eventType]
	return routingKey, ok
}
