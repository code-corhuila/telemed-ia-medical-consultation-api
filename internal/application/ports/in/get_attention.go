package in

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
)

// GetAttentionPort is the input boundary for reading an attention summary.
type GetAttentionPort interface {
	Execute(ctx context.Context, id int64) (*dto.AttentionSummaryResponse, error)
}
