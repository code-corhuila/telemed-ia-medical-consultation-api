package usecase

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
)

// GetAttention reads an attention summary by id.
type GetAttention struct {
	summaries out.AttentionSummaryRepositoryPort
}

func NewGetAttention(summaries out.AttentionSummaryRepositoryPort) *GetAttention {
	return &GetAttention{summaries: summaries}
}

func (uc *GetAttention) Execute(ctx context.Context, id int64) (*dto.AttentionSummaryResponse, error) {
	if id <= 0 {
		return nil, exception.ErrAttentionSummaryNotFound
	}
	s, err := uc.summaries.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, exception.ErrAttentionSummaryNotFound
	}
	return toAttentionResponse(s), nil
}
