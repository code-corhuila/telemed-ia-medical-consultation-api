package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase/fakes"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

const (
	startAppointmentID = "a1111111-1111-4111-8111-111111111111"
	startPatientID     = "b2222222-2222-4222-8222-222222222222"
	startDoctorID      = "c3333333-3333-4333-8333-333333333333"
	startOtherDoctorID = "d4444444-4444-4444-8444-444444444444"
)

func startFixture(status string) (*usecase.StartConsultation, *fakes.ConsultationRepo, *fakes.AppointmentReader) {
	cs := fakes.NewConsultationRepo()
	ar := fakes.NewAppointmentReader()
	ar.Appointments[startAppointmentID] = out.AppointmentView{
		ID: startAppointmentID, PatientID: startPatientID, ProfessionalID: startDoctorID, Status: status,
	}
	return usecase.NewStartConsultation(cs, ar), cs, ar
}

func startCommand(callerID, role string) dto.StartConsultationCommand {
	return dto.StartConsultationCommand{
		AppointmentID: startAppointmentID, CallerID: callerID, CallerRole: role, AccessToken: "caller-token",
	}
}

func TestStartConsultation_CreatesAnInProgressConsultationFromTheAppointment(t *testing.T) {
	for _, status := range []string{"CONFIRMED", "RESCHEDULED"} {
		uc, cs, ar := startFixture(status)

		resp, created, err := uc.Execute(context.Background(), startCommand(startDoctorID, "PROFESSIONAL"))

		if err != nil || !created {
			t.Fatalf("%s: expected a new consultation, got created=%v err=%v", status, created, err)
		}
		if resp.ID == "" || resp.AppointmentID != startAppointmentID || resp.PatientID != startPatientID ||
			resp.ProfessionalID != startDoctorID || resp.Status != model.StatusInProgress || resp.CompletedAt != nil {
			t.Fatalf("%s: unexpected response %+v", status, resp)
		}
		if ar.LastToken != "caller-token" {
			t.Fatalf("%s: the caller's token was not relayed, got %q", status, ar.LastToken)
		}
		if stored, _ := cs.FindByAppointmentID(context.Background(), startAppointmentID); stored == nil {
			t.Fatalf("%s: consultation not stored", status)
		}
	}
}

func TestStartConsultation_ReturnsTheExistingConsultationWhateverItsStatus(t *testing.T) {
	uc, cs, ar := startFixture("CONFIRMED")
	first, _, err := uc.Execute(context.Background(), startCommand(startDoctorID, "PROFESSIONAL"))
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if _, err := cs.Complete(context.Background(), first.ID); err != nil {
		t.Fatalf("complete: %v", err)
	}
	// Recording the attention moves the appointment to COMPLETED.
	appointment := ar.Appointments[startAppointmentID]
	appointment.Status = "COMPLETED"
	ar.Appointments[startAppointmentID] = appointment

	again, created, err := uc.Execute(context.Background(), startCommand(startDoctorID, "PROFESSIONAL"))

	if err != nil || created {
		t.Fatalf("expected the existing consultation, got created=%v err=%v", created, err)
	}
	if again.ID != first.ID || again.Status != model.StatusCompleted {
		t.Fatalf("unexpected response %+v", again)
	}
}

func TestStartConsultation_OnlyProfessionals(t *testing.T) {
	for _, role := range []string{"PATIENT", "ADMIN", "SERVICE", ""} {
		uc, _, ar := startFixture("CONFIRMED")

		_, _, err := uc.Execute(context.Background(), startCommand(startDoctorID, role))

		if !errors.Is(err, exception.ErrForbidden) {
			t.Fatalf("%q: expected ErrForbidden, got %v", role, err)
		}
		if ar.LastToken != "" {
			t.Fatalf("%q: the appointment service must not be called", role)
		}
	}
}

func TestStartConsultation_RefusesTheAppointmentOfAnotherProfessional(t *testing.T) {
	uc, cs, _ := startFixture("CONFIRMED")

	_, _, err := uc.Execute(context.Background(), startCommand(startOtherDoctorID, "PROFESSIONAL"))

	if !errors.Is(err, exception.ErrAppointmentNotAttendable) {
		t.Fatalf("expected ErrAppointmentNotAttendable, got %v", err)
	}
	if stored, _ := cs.FindByAppointmentID(context.Background(), startAppointmentID); stored != nil {
		t.Fatal("no consultation must be stored")
	}
}

func TestStartConsultation_RefusesAnAppointmentThatCannotBeAttended(t *testing.T) {
	for _, status := range []string{"CANCELLED", "NO_SHOW", "COMPLETED"} {
		uc, _, _ := startFixture(status)

		_, _, err := uc.Execute(context.Background(), startCommand(startDoctorID, "PROFESSIONAL"))

		if !errors.Is(err, exception.ErrAppointmentNotAttendable) {
			t.Fatalf("%s: expected ErrAppointmentNotAttendable, got %v", status, err)
		}
	}
}

func TestStartConsultation_PassesOnTheAppointmentServiceErrors(t *testing.T) {
	for _, want := range []error{
		exception.ErrAppointmentNotFound, exception.ErrForbidden, exception.ErrAppointmentServiceUnavailable,
	} {
		uc, _, ar := startFixture("CONFIRMED")
		ar.Err = want

		_, _, err := uc.Execute(context.Background(), startCommand(startDoctorID, "PROFESSIONAL"))

		if !errors.Is(err, want) {
			t.Fatalf("expected %v, got %v", want, err)
		}
	}
}

func TestStartConsultation_RejectsAMalformedAppointmentID(t *testing.T) {
	uc, _, ar := startFixture("CONFIRMED")
	cmd := startCommand(startDoctorID, "PROFESSIONAL")
	cmd.AppointmentID = "not-a-uuid"

	_, _, err := uc.Execute(context.Background(), cmd)

	if !errors.Is(err, exception.ErrInvalidConsultation) {
		t.Fatalf("expected ErrInvalidConsultation, got %v", err)
	}
	if ar.LastToken != "" {
		t.Fatal("the appointment service must not be called")
	}
}
