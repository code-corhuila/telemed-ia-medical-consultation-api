package persistence

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// PostSummaryRepo implements out.PostSummaryRepositoryPort on PostgreSQL.
type PostSummaryRepo struct {
	pool *pgxpool.Pool
}

func NewPostSummaryRepo(pool *pgxpool.Pool) *PostSummaryRepo {
	return &PostSummaryRepo{pool: pool}
}

const postSummaryColumns = `
	id, consultation_id, appointment_id, patient_id,
	preconsultation_summary_id, attention_summary_id, generated_at
`

func (r *PostSummaryRepo) Save(ctx context.Context, s *model.PostSummary) (*model.PostSummary, error) {
	const q = `
		INSERT INTO post_summaries
			(consultation_id, appointment_id, patient_id,
			 preconsultation_summary_id, attention_summary_id, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + postSummaryColumns

	row := r.pool.QueryRow(ctx, q,
		s.ConsultationID, s.AppointmentID, s.PatientID,
		s.PreconsultationSummaryID, s.AttentionSummaryID, s.GeneratedAt,
	)
	return scanPostSummary(row)
}

func (r *PostSummaryRepo) FindByID(ctx context.Context, id int64) (*model.PostSummary, error) {
	q := `SELECT ` + postSummaryColumns + ` FROM post_summaries WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanPostSummary(row)
}

func (r *PostSummaryRepo) FindByConsultationID(ctx context.Context, consultationID int64) (*model.PostSummary, error) {
	q := `SELECT ` + postSummaryColumns + ` FROM post_summaries WHERE consultation_id = $1`
	row := r.pool.QueryRow(ctx, q, consultationID)
	return scanPostSummary(row)
}

func scanPostSummary(row pgx.Row) (*model.PostSummary, error) {
	var s model.PostSummary
	err := row.Scan(
		&s.ID, &s.ConsultationID, &s.AppointmentID, &s.PatientID,
		&s.PreconsultationSummaryID, &s.AttentionSummaryID, &s.GeneratedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
