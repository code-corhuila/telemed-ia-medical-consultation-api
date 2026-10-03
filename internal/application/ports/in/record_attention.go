package in

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
)

// RecordAttentionPort is the input boundary for recording an attention summary.
type RecordAttentionPort interface {
	Execute(ctx context.Context, cmd dto.RecordAttentionCommand) (*dto.AttentionSummaryResponse, error)
}
