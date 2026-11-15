package messaging

import "testing"

func TestRoutingKeyFor_EachPublishedEventHasItsOwnKey(t *testing.T) {
	cases := map[string]string{
		"ConsultationCompleted":            "consultation-completed",
		"PostConsultationSummaryGenerated": "post-consultation-summary-generated",
	}
	for eventType, want := range cases {
		got, ok := routingKeyFor(eventType)
		if !ok || got != want {
			t.Fatalf("%s: got (%q, %v), want %q", eventType, got, ok, want)
		}
	}
}

func TestRoutingKeyFor_UnknownTypesAreNotPublished(t *testing.T) {
	for _, eventType := range []string{"", "ReferralCreated", "consultationcompleted"} {
		if key, ok := routingKeyFor(eventType); ok {
			t.Fatalf("%q: expected no routing key, got %q", eventType, key)
		}
	}
}
