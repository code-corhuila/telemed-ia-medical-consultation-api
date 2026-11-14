package out

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// ConsultationRepositoryPort is the persistence boundary required by the
// application layer. The concrete implementation lives in infrastructure.
type ConsultationRepositoryPort interface {
	Save(ctx context.Context, c *model.Consultation) (*model.Consultation, error)
	FindByID(ctx context.Context, id string) (*model.Consultation, error)
	FindByAppointmentID(ctx context.Context, appointmentID string) (*model.Consultation, error)
	Complete(ctx context.Context, id string) (*model.Consultation, error)
}
