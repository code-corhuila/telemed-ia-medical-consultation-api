package usecase_test

import (
	"context"
	"testing"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase/fakes"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
)

func TestRecordAttention_RequiresConsultation(t *testing.T) {
	uc := usecase.NewRecordAttention(fakes.NewConsultationRepo(), fakes.NewAttentionRepo())
	_, err := uc.Execute(context.Background(), dto.RecordAttentionCommand{ConsultationID: 99})
	if err != exception.ErrConsultationNotFound {
		t.Fatalf("expected ErrConsultationNotFound, got %v", err)
	}
}

func TestRecordAttention_RequiresRecommendations(t *testing.T) {
	cs := fakes.NewConsultationRepo()
	uc := usecase.NewRecordAttention(cs, fakes.NewAttentionRepo())

	ctx := context.Background()
	if err := seedConsultation(ctx, cs, 1); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	_, err := uc.Execute(ctx, dto.RecordAttentionCommand{ConsultationID: 1, Recommendations: ""})
	if err != exception.ErrInvalidAttentionSummary {
		t.Fatalf("expected ErrInvalidAttentionSummary, got %v", err)
	}
}

func TestRecordAttention_HappyPath(t *testing.T) {
	cs := fakes.NewConsultationRepo()
	as := fakes.NewAttentionRepo()
	uc := usecase.NewRecordAttention(cs, as)

	ctx := context.Background()
	if err := seedConsultation(ctx, cs, 1); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	resp, err := uc.Execute(ctx, dto.RecordAttentionCommand{
		ConsultationID:  1,
		Recommendations: "Rest and hydration",
		Medications: []dto.MedicationDTO{
			{Name: "Paracetamol"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.ConsultationID != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
