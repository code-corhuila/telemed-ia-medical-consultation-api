package usecase_test

import (
	"context"
	"testing"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase/fakes"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

func TestRecordAttention_RequiresConsultation(t *testing.T) {
	uc := usecase.NewRecordAttention(fakes.NewConsultationRepo(), fakes.NewAttentionRepo())
	_, err := uc.Execute(context.Background(), dto.RecordAttentionCommand{ConsultationID: uuidFor(99)})
	if err != exception.ErrConsultationNotFound {
		t.Fatalf("expected ErrConsultationNotFound, got %v", err)
	}
}

func TestRecordAttention_RequiresRecommendations(t *testing.T) {
	cs := fakes.NewConsultationRepo()
	uc := usecase.NewRecordAttention(cs, fakes.NewAttentionRepo())

	ctx := context.Background()
	if err := seedConsultation(ctx, cs, uuidFor(1)); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	_, err := uc.Execute(ctx, dto.RecordAttentionCommand{ConsultationID: uuidFor(1), Recommendations: ""})
	if err != exception.ErrInvalidAttentionSummary {
		t.Fatalf("expected ErrInvalidAttentionSummary, got %v", err)
	}
}

func TestRecordAttention_HappyPath(t *testing.T) {
	cs := fakes.NewConsultationRepo()
	as := fakes.NewAttentionRepo()
	uc := usecase.NewRecordAttention(cs, as)

	ctx := context.Background()
	if err := seedConsultation(ctx, cs, uuidFor(1)); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	resp, err := uc.Execute(ctx, dto.RecordAttentionCommand{
		ConsultationID:  uuidFor(1),
		Recommendations: "Rest and hydration",
		Medications: []dto.MedicationDTO{
			{Name: "Paracetamol"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.ConsultationID != uuidFor(1) {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestRecordAttention_RejectsCompletedConsultation(t *testing.T) {
	cs := fakes.NewConsultationRepo()
	uc := usecase.NewRecordAttention(cs, fakes.NewAttentionRepo())
	ctx := context.Background()

	c, _ := model.NewConsultation(uuidFor(1), uuidFor(2), uuidFor(3))
	c.ID = uuidFor(1)
	_ = c.Complete() // La consulta ya está completada
	_, _ = cs.Save(ctx, c)

	_, err := uc.Execute(ctx, dto.RecordAttentionCommand{
		ConsultationID:  uuidFor(1),
		Recommendations: "Rest",
	})

	if err != exception.ErrConsultationAlreadyCompleted {
		t.Fatalf("expected ErrConsultationAlreadyCompleted, got %v", err)
	}
}
