package fakes

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
)

// AppointmentReader is an in-memory fake of appointment-scheduling.
type AppointmentReader struct {
	Appointments map[string]out.AppointmentView
	// Err, when set, is returned by every call.
	Err error
	// LastToken is the access token of the last call.
	LastToken string
}

func NewAppointmentReader() *AppointmentReader {
	return &AppointmentReader{Appointments: map[string]out.AppointmentView{}}
}

func (r *AppointmentReader) FindAppointment(_ context.Context, appointmentID, accessToken string) (*out.AppointmentView, error) {
	r.LastToken = accessToken
	if r.Err != nil {
		return nil, r.Err
	}
	a, ok := r.Appointments[appointmentID]
	if !ok {
		return nil, exception.ErrAppointmentNotFound
	}
	return &a, nil
}
