package usecase

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// GetPostSummary reads a post-consultation summary by id.
type GetPostSummary struct {
	postSummaries out.PostSummaryRepositoryPort
}

func NewGetPostSummary(postSummaries out.PostSummaryRepositoryPort) *GetPostSummary {
	return &GetPostSummary{postSummaries: postSummaries}
}

func (uc *GetPostSummary) Execute(ctx context.Context, id int64) (*dto.PostSummaryResponse, error) {
	if id <= 0 {
		return nil, exception.ErrPostSummaryNotFound
	}
	s, err := uc.postSummaries.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, exception.ErrPostSummaryNotFound
	}
	return toPostSummaryResponse(s), nil
}

func toPostSummaryResponse(s *model.PostSummary) *dto.PostSummaryResponse {
	return &dto.PostSummaryResponse{
		ID:                       s.ID,
		ConsultationID:           s.ConsultationID,
		AppointmentID:            s.AppointmentID,
		PatientID:                s.PatientID,
		PreconsultationSummaryID: s.PreconsultationSummaryID,
		AttentionSummaryID:       s.AttentionSummaryID,
		GeneratedAt:              s.GeneratedAt,
	}
}
