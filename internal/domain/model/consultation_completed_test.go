package model

import (
	"encoding/json"
	"testing"
	"time"
)

func completedConsultation(id string) *Consultation {
	completedAt := time.Date(2026, 11, 11, 15, 4, 5, 0, time.UTC)
	return ReconstituteConsultation(id,
		"a1111111-1111-4111-8111-111111111111",
		"b2222222-2222-4222-8222-222222222222",
		"c3333333-3333-4333-8333-333333333333",
		StatusCompleted, &completedAt, completedAt, completedAt, nil)
}

func TestNewConsultationCompletedEvent_BuildsTheEnvelopeWithIdentifiersOnly(t *testing.T) {
	occurredAt := time.Date(2026, 11, 11, 15, 4, 6, 0, time.UTC)

	ev, err := NewConsultationCompletedEvent(completedConsultation(uuidA), "", "corr-1", occurredAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ev.EventType != "ConsultationCompleted" || ev.CorrelationID != "corr-1" || ev.ID == "" {
		t.Fatalf("unexpected event %+v", ev)
	}
	var envelope map[string]any
	if err := json.Unmarshal(ev.Payload, &envelope); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if envelope["eventId"] != ev.ID || envelope["eventType"] != "ConsultationCompleted" ||
		envelope["correlationId"] != "corr-1" || envelope["occurredAt"] != "2026-11-11T15:04:06Z" ||
		envelope["version"] != float64(1) {
		t.Fatalf("unexpected envelope %v", envelope)
	}
	payload, _ := envelope["payload"].(map[string]any)
	want := map[string]any{
		"consultationId": uuidA,
		"appointmentId":  "a1111111-1111-4111-8111-111111111111",
		"patientId":      "b2222222-2222-4222-8222-222222222222",
		"professionalId": "c3333333-3333-4333-8333-333333333333",
		"completedAt":    "2026-11-11T15:04:05Z",
	}
	if len(payload) != len(want) {
		t.Fatalf("payload must carry identifiers only, got %v", payload)
	}
	for key, value := range want {
		if payload[key] != value {
			t.Fatalf("payload[%s] = %v, want %v", key, payload[key], value)
		}
	}
}

func TestNewConsultationCompletedEvent_DerivesTheEventIdFromTheConsultation(t *testing.T) {
	first, _ := NewConsultationCompletedEvent(completedConsultation(uuidA), "", "corr-1", time.Now())
	again, _ := NewConsultationCompletedEvent(completedConsultation(uuidA), "", "corr-2", time.Now())
	other, _ := NewConsultationCompletedEvent(completedConsultation(uuidB), "", "corr-1", time.Now())

	if first.ID == "" || first.ID != again.ID {
		t.Fatalf("the same consultation must give the same event id: %q vs %q", first.ID, again.ID)
	}
	if first.ID == other.ID {
		t.Fatal("different consultations must give different event ids")
	}
}

func TestNewConsultationCompletedEvent_RequiresACompletedConsultation(t *testing.T) {
	inProgress, _ := NewConsultation(uuidA, uuidB, uuidC)
	inProgress.ID = uuidA

	if _, err := NewConsultationCompletedEvent(inProgress, "", "corr-1", time.Now()); err == nil {
		t.Fatal("expected an error for a consultation that is not completed")
	}
}

func TestNewConsultationCompletedEvent_CarriesThePostSummaryIdWhenPresent(t *testing.T) {
	postSummaryID := "d4444444-4444-4444-8444-444444444444"

	ev, err := NewConsultationCompletedEvent(completedConsultation(uuidA), postSummaryID, "corr-1", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var envelope map[string]any
	_ = json.Unmarshal(ev.Payload, &envelope)
	payload, _ := envelope["payload"].(map[string]any)
	if payload["postSummaryId"] != postSummaryID || len(payload) != 6 {
		t.Fatalf("payload must add postSummaryId, got %v", payload)
	}
}
