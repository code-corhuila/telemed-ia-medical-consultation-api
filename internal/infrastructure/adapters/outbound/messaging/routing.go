package messaging

// routingKeyFor returns the routing key of an event type (DEC-007): the event name in
// kebab-case. ok is false for a type this relay does not publish.
func routingKeyFor(eventType string) (routingKey string, ok bool) {
	return "", false
}
