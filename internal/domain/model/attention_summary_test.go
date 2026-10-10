package model

import "testing"

func TestNewAttentionSummary_RequiresConsultationID(t *testing.T) {
	_, err := NewAttentionSummary(0, "", "rest", nil, nil, nil, nil)
	if err == nil {
		t.Fatalf("expected error for zero consultationID")
	}
}

func TestNewAttentionSummary_RequiresRecommendations(t *testing.T) {
	_, err := NewAttentionSummary(1, "", "   ", nil, nil, nil, nil)
	if err == nil {
		t.Fatalf("expected error for empty recommendations")
	}
}

func TestNewAttentionSummary_NormalizesNilMedications(t *testing.T) {
	s, err := NewAttentionSummary(1, "", "rest", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Medications == nil {
		t.Fatalf("Medications should be an empty slice, not nil")
	}
	if len(s.Medications) != 0 {
		t.Fatalf("Medications should be empty")
	}
}

func TestMedicationInstruction_RequiresName(t *testing.T) {
	_, err := NewMedicationInstruction("   ", nil, nil, nil, nil)
	if err == nil {
		t.Fatalf("expected error for blank name")
	}
}

func TestMedicationInstruction_CleansOptionalFields(t *testing.T) {
	empty := "   "
	m, err := NewMedicationInstruction("Paracetamol", &empty, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Dosage != nil {
		t.Fatalf("blank dosage should be normalized to nil")
	}
}
