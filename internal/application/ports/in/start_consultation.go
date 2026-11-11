package in

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
)

// StartConsultationPort creates the consultation of an appointment, or returns the existing
// one (DEC-006). created is false when the consultation already existed.
type StartConsultationPort interface {
	Execute(ctx context.Context, cmd dto.StartConsultationCommand) (resp *dto.ConsultationResponse, created bool, err error)
}
