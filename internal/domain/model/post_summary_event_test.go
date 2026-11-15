package model

import (
	"encoding/json"
	"testing"
	"time"
)

func savedPostSummary(id string) *PostSummary {
	attentionID := "e5555555-5555-4555-8555-555555555555"
	return &PostSummary{
		ID:                 id,
		ConsultationID:     uuidA,
		AppointmentID:      "a1111111-1111-4111-8111-111111111111",
		PatientID:          "b2222222-2222-4222-8222-222222222222",
		AttentionSummaryID: &attentionID,
		GeneratedAt:        time.Date(2026, 11, 11, 15, 4, 5, 0, time.UTC),
	}
}

func TestNewPostSummaryGeneratedEvent_BuildsTheEnvelope(t *testing.T) {
	occurredAt := time.Date(2026, 11, 11, 15, 4, 6, 0, time.UTC)

	ev, err := NewPostSummaryGeneratedEvent(savedPostSummary(uuidB), "corr-1", occurredAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ev.EventType != "PostConsultationSummaryGenerated" || ev.CorrelationID != "corr-1" || ev.Status != "PENDING" {
		t.Fatalf("unexpected event %+v", ev)
	}
	var envelope map[string]any
	if err := json.Unmarshal(ev.Payload, &envelope); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if envelope["eventId"] != ev.ID || envelope["eventType"] != "PostConsultationSummaryGenerated" ||
		envelope["occurredAt"] != "2026-11-11T15:04:06Z" || envelope["version"] != float64(1) {
		t.Fatalf("unexpected envelope %v", envelope)
	}
	payload, _ := envelope["payload"].(map[string]any)
	want := map[string]any{
		"summaryId":      uuidB,
		"consultationId": uuidA,
		"appointmentId":  "a1111111-1111-4111-8111-111111111111",
		"patientId":      "b2222222-2222-4222-8222-222222222222",
		"generatedAt":    "2026-11-11T15:04:05Z",
	}
	if len(payload) != len(want) {
		t.Fatalf("unexpected payload %v", payload)
	}
	for key, value := range want {
		if payload[key] != value {
			t.Fatalf("payload[%s] = %v, want %v", key, payload[key], value)
		}
	}
}

func TestNewPostSummaryGeneratedEvent_DerivesTheEventIdFromTheSummary(t *testing.T) {
	first, _ := NewPostSummaryGeneratedEvent(savedPostSummary(uuidB), "corr-1", time.Now())
	again, _ := NewPostSummaryGeneratedEvent(savedPostSummary(uuidB), "corr-2", time.Now())
	other, _ := NewPostSummaryGeneratedEvent(savedPostSummary(uuidC), "corr-1", time.Now())

	if first.ID == "" || first.ID != again.ID || first.ID == other.ID {
		t.Fatalf("event ids must be derived from the summary id: %q %q %q", first.ID, again.ID, other.ID)
	}
}

func TestNewPostSummaryGeneratedEvent_RequiresASavedSummary(t *testing.T) {
	if _, err := NewPostSummaryGeneratedEvent(savedPostSummary(""), "corr-1", time.Now()); err == nil {
		t.Fatal("expected an error for a summary without id")
	}
}
