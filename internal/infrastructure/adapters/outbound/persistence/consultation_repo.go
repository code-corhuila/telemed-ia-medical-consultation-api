package persistence

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// ConsultationRepo implements out.ConsultationRepositoryPort on PostgreSQL.
type ConsultationRepo struct {
	pool *pgxpool.Pool
}

func NewConsultationRepo(pool *pgxpool.Pool) *ConsultationRepo {
	return &ConsultationRepo{pool: pool}
}

const consultationColumns = `
	id, appointment_id, patient_id, professional_id, status,
	completed_at, created_at, updated_at, deleted_at
`

func (r *ConsultationRepo) Save(ctx context.Context, c *model.Consultation) (*model.Consultation, error) {
	const q = `
		INSERT INTO consultations
			(appointment_id, patient_id, professional_id, status, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + consultationColumns

	row := r.pool.QueryRow(ctx, q,
		c.AppointmentID, c.PatientID, c.ProfessionalID,
		c.Status, c.CompletedAt, c.CreatedAt, c.UpdatedAt,
	)
	return scanConsultation(row)
}

func (r *ConsultationRepo) FindByID(ctx context.Context, id int64) (*model.Consultation, error) {
	q := `SELECT ` + consultationColumns + ` FROM consultations WHERE id = $1 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, q, id)
	return scanConsultation(row)
}

func (r *ConsultationRepo) FindByAppointmentID(ctx context.Context, appointmentID int64) (*model.Consultation, error) {
	q := `SELECT ` + consultationColumns + ` FROM consultations WHERE appointment_id = $1 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, q, appointmentID)
	return scanConsultation(row)
}

func (r *ConsultationRepo) Complete(ctx context.Context, id int64) (*model.Consultation, error) {
	const q = `
		UPDATE consultations
		SET status = $2, completed_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + consultationColumns

	row := r.pool.QueryRow(ctx, q, id, model.StatusCompleted)
	return scanConsultation(row)
}

func scanConsultation(row pgx.Row) (*model.Consultation, error) {
	var c model.Consultation
	err := row.Scan(
		&c.ID, &c.AppointmentID, &c.PatientID, &c.ProfessionalID,
		&c.Status, &c.CompletedAt, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
