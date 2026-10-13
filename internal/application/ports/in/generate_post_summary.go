package in

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
)

// GeneratePostSummaryPort is the input boundary for generating a post-summary.
type GeneratePostSummaryPort interface {
	Execute(ctx context.Context, cmd dto.GeneratePostSummaryCommand) (*dto.PostSummaryResponse, error)
}
