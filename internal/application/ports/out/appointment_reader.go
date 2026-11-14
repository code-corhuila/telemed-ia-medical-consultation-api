package out

import "context"

// AppointmentView is the part of an appointment the consultation domain needs.
type AppointmentView struct {
	ID             string
	PatientID      string
	ProfessionalID string
	Status         string
}

// AppointmentReaderPort reads one appointment from appointment-scheduling, relaying the
// caller's access token so that service applies its own ownership rules (DEC-006).
//
// Errors: exception.ErrAppointmentNotFound, exception.ErrForbidden,
// exception.ErrAppointmentServiceUnavailable.
type AppointmentReaderPort interface {
	FindAppointment(ctx context.Context, appointmentID, accessToken string) (*AppointmentView, error)
}
