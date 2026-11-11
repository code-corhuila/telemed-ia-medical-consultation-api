package usecase

import (
	"context"
	"errors"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
)

// StartConsultation creates the consultation of an appointment (DEC-006).
type StartConsultation struct {
	consultations out.ConsultationRepositoryPort
	appointments  out.AppointmentReaderPort
}

func NewStartConsultation(
	consultations out.ConsultationRepositoryPort,
	appointments out.AppointmentReaderPort,
) *StartConsultation {
	return &StartConsultation{consultations: consultations, appointments: appointments}
}

func (uc *StartConsultation) Execute(_ context.Context, _ dto.StartConsultationCommand) (*dto.ConsultationResponse, bool, error) {
	return nil, false, errors.New("not implemented")
}
