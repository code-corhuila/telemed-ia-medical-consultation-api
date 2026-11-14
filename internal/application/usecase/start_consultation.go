package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

const roleProfessional = "PROFESSIONAL"

// attendableStatuses are the appointment statuses a consultation can start from (DEC-006).
var attendableStatuses = map[string]bool{"CONFIRMED": true, "RESCHEDULED": true}

// StartConsultation creates the consultation of an appointment, or returns the existing one
// (DEC-006). One consultation per appointment.
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

func (uc *StartConsultation) Execute(ctx context.Context, cmd dto.StartConsultationCommand) (*dto.ConsultationResponse, bool, error) {
	if cmd.CallerRole != roleProfessional {
		return nil, false, exception.ErrForbidden
	}
	if _, err := uuid.Parse(cmd.AppointmentID); err != nil || len(cmd.AppointmentID) != 36 {
		return nil, false, exception.ErrInvalidConsultation
	}

	appointment, err := uc.appointments.FindAppointment(ctx, cmd.AppointmentID, cmd.AccessToken)
	if err != nil {
		return nil, false, err
	}
	if appointment.ProfessionalID != cmd.CallerID {
		return nil, false, exception.ErrAppointmentNotAttendable
	}

	existing, err := uc.consultations.FindByAppointmentID(ctx, cmd.AppointmentID)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return toConsultationResponse(existing), false, nil
	}
	if !attendableStatuses[appointment.Status] {
		return nil, false, exception.ErrAppointmentNotAttendable
	}

	consultation, err := model.NewConsultation(cmd.AppointmentID, appointment.PatientID, appointment.ProfessionalID)
	if err != nil {
		return nil, false, err
	}
	saved, err := uc.consultations.Save(ctx, consultation)
	if err != nil {
		// A concurrent call may have created it first (uq_consultations_appointment).
		if winner, findErr := uc.consultations.FindByAppointmentID(ctx, cmd.AppointmentID); findErr == nil && winner != nil {
			return toConsultationResponse(winner), false, nil
		}
		return nil, false, err
	}
	return toConsultationResponse(saved), true, nil
}

func toConsultationResponse(c *model.Consultation) *dto.ConsultationResponse {
	return &dto.ConsultationResponse{
		ID:             c.ID,
		AppointmentID:  c.AppointmentID,
		PatientID:      c.PatientID,
		ProfessionalID: c.ProfessionalID,
		Status:         c.Status,
		CompletedAt:    c.CompletedAt,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
