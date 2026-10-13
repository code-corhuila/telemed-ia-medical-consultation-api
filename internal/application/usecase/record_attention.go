package usecase

import (
	"context"
	"errors"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// RecordAttention records the clinical attention summary of a consultation.
type RecordAttention struct {
	consultations out.ConsultationRepositoryPort
	summaries     out.AttentionSummaryRepositoryPort
}

func NewRecordAttention(
	consultations out.ConsultationRepositoryPort,
	summaries out.AttentionSummaryRepositoryPort,
) *RecordAttention {
	return &RecordAttention{consultations: consultations, summaries: summaries}
}

func (uc *RecordAttention) Execute(ctx context.Context, cmd dto.RecordAttentionCommand) (*dto.AttentionSummaryResponse, error) {
	if cmd.ConsultationID <= 0 {
		return nil, exception.ErrInvalidAttentionSummary
	}
	c, err := uc.consultations.FindByID(ctx, cmd.ConsultationID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, exception.ErrConsultationNotFound
	}
	if err := c.EnsureInProgress(); err != nil {
		return nil, errors.New("consultation is not in progress")
	}

	meds := make([]model.MedicationInstruction, 0, len(cmd.Medications))
	for _, m := range cmd.Medications {
		med, err := model.NewMedicationInstruction(m.Name, m.Dosage, m.Frequency, m.Duration, m.Instructions)
		if err != nil {
			return nil, exception.ErrInvalidAttentionSummary
		}
		meds = append(meds, med)
	}

	var diag string
	if cmd.Diagnosis != nil {
		diag = *cmd.Diagnosis
	}
	s, err := model.NewAttentionSummary(
		cmd.ConsultationID, diag, cmd.Recommendations, meds,
		cmd.Observations, cmd.Referral, cmd.FollowUp,
	)
	if err != nil {
		return nil, exception.ErrInvalidAttentionSummary
	}
	saved, err := uc.summaries.Save(ctx, s)
	if err != nil {
		return nil, err
	}
	return toAttentionResponse(saved), nil
}

func toAttentionResponse(s *model.AttentionSummary) *dto.AttentionSummaryResponse {
	meds := make([]dto.MedicationDTO, 0, len(s.Medications))
	for _, m := range s.Medications {
		meds = append(meds, dto.MedicationDTO{
			Name: m.Name, Dosage: m.Dosage, Frequency: m.Frequency,
			Duration: m.Duration, Instructions: m.Instructions,
		})
	}
	return &dto.AttentionSummaryResponse{
		ID: s.ID, ConsultationID: s.ConsultationID,
		Diagnosis: s.Diagnosis, Recommendations: s.Recommendations,
		Medications: meds, Observations: s.Observations,
		Referral: s.Referral, FollowUp: s.FollowUp,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
}
