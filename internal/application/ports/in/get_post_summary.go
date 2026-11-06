package in

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
)

// GetPostSummaryPort is the input boundary for reading a post-summary.
type GetPostSummaryPort interface {
	Execute(ctx context.Context, id string) (*dto.PostSummaryResponse, error)
}
