package persistence

import (
	"context"
	"errors"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	saved, err := scanPostSummary(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, exception.ErrPostSummaryAlreadyExists
		}
		return nil, err
	}
	return saved, nil
}

func (r *PostSummaryRepo) FindByID(ctx context.Context, id string) (*model.PostSummary, error) {
	q := `SELECT ` + postSummaryColumns + ` FROM post_summaries WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanPostSummary(row)
}

func (r *PostSummaryRepo) FindByConsultationID(ctx context.Context, consultationID string) (*model.PostSummary, error) {
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

func (r *PostSummaryRepo) SaveWithOutbox(ctx context.Context, s *model.PostSummary, e *model.OutboxEvent) (*model.PostSummary, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Insertar PostSummary
	const qInsertSummary = `
		INSERT INTO post_summaries
			(consultation_id, appointment_id, patient_id,
			 preconsultation_summary_id, attention_summary_id, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + postSummaryColumns

	row := tx.QueryRow(ctx, qInsertSummary,
		s.ConsultationID, s.AppointmentID, s.PatientID,
		s.PreconsultationSummaryID, s.AttentionSummaryID, s.GeneratedAt,
	)
	saved, err := scanPostSummary(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, exception.ErrPostSummaryAlreadyExists
		}
		return nil, err
	}

	// 2. Insertar OutboxEvent
	const qInsertOutbox = `
		INSERT INTO outbox_event (event_type, payload, correlation_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(ctx, qInsertOutbox, e.EventType, e.Payload, e.CorrelationID, e.Status, e.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return saved, nil
}
