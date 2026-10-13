package out

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// PostSummaryRepositoryPort is the persistence boundary for post-consultation summaries.
type PostSummaryRepositoryPort interface {
	Save(ctx context.Context, s *model.PostSummary) (*model.PostSummary, error)
	FindByID(ctx context.Context, id int64) (*model.PostSummary, error)
	FindByConsultationID(ctx context.Context, consultationID int64) (*model.PostSummary, error)
}
