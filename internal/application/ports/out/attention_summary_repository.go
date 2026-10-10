package out

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// AttentionSummaryRepositoryPort is the persistence boundary for attention summaries.
type AttentionSummaryRepositoryPort interface {
	Save(ctx context.Context, s *model.AttentionSummary) (*model.AttentionSummary, error)
	FindByID(ctx context.Context, id int64) (*model.AttentionSummary, error)
	FindByConsultationID(ctx context.Context, consultationID int64) (*model.AttentionSummary, error)
}
