package usecase_test

import (
	"context"
	"testing"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase/fakes"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

const professionalID = "c3333333-3333-4333-8333-333333333333"

func seedCompletedConsultation(t *testing.T, repo *fakes.ConsultationRepo) *model.Consultation {
	t.Helper()
	c, err := model.NewConsultation(uuidFor(10), uuidFor(20), professionalID)
	if err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	if err := c.Complete(); err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	saved, err := repo.Save(context.Background(), c)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	return saved
}

func TestGeneratePostSummary_ReturnsTheExistingSummaryWhenItAlreadyExists(t *testing.T) {
	ctx := context.Background()
	cs, ps := fakes.NewConsultationRepo(), fakes.NewPostSummaryRepo()
	c := seedCompletedConsultation(t, cs)
	uc := usecase.NewGeneratePostSummary(cs, fakes.NewAttentionRepo(), ps)

	first, err := uc.Execute(ctx, dto.GeneratePostSummaryCommand{ConsultationID: c.ID})
	if err != nil || !first.Created {
		t.Fatalf("first call must create the summary: %+v, %v", first, err)
	}
	again, err := uc.Execute(ctx, dto.GeneratePostSummaryCommand{ConsultationID: c.ID})
	if err != nil {
		t.Fatalf("second call must not fail: %v", err)
	}
	if again.Created || again.ID != first.ID {
		t.Fatalf("second call must return the existing summary %q, got %+v", first.ID, again)
	}
}

func TestGetPostSummary_AddsTheProfessionalAndTheCompletionDate(t *testing.T) {
	ctx := context.Background()
	cs, ps := fakes.NewConsultationRepo(), fakes.NewPostSummaryRepo()
	c := seedCompletedConsultation(t, cs)
	s, _ := model.NewPostSummary(c.ID, c.AppointmentID, c.PatientID, nil, nil)
	saved, _ := ps.Save(ctx, s)

	got, err := usecase.NewGetPostSummary(ps, cs).Execute(ctx, saved.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ProfessionalID != professionalID {
		t.Fatalf("professionalId = %q, want %q", got.ProfessionalID, professionalID)
	}
	if got.CompletedAt == nil || !got.CompletedAt.Equal(*c.CompletedAt) {
		t.Fatalf("completedAt = %v, want %v", got.CompletedAt, c.CompletedAt)
	}
}
