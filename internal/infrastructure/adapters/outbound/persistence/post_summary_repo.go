package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
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

func (r *PostSummaryRepo) SaveWithOutbox(ctx context.Context, s *model.PostSummary, correlationID string) (*model.PostSummary, error) {
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

	// 2. Construir el envelope comun y el payload con el ID real.
	//    eventId es determinista a partir de summaryId + eventType:
	//    re-generar el summary para la misma consulta produce el mismo
	//    eventId, y el consumidor puede deduplicar por eventId.
	eventType := "PostConsultationSummaryGenerated"
	eventID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(saved.ID+":"+eventType)).String()

	envelopePayload, err := json.Marshal(map[string]interface{}{
		"eventId":       eventID,
		"eventType":     eventType,
		"occurredAt":    time.Now().UTC().Format(time.RFC3339),
		"correlationId": correlationID,
		"version":       1,
		"payload": map[string]interface{}{
			"summaryId":      saved.ID,
			"consultationId": saved.ConsultationID,
			"appointmentId":  saved.AppointmentID,
			"patientId":      saved.PatientID,
			"generatedAt":    saved.GeneratedAt.UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal outbox payload: %w", err)
	}

	// 3. Insertar outbox_event con event_id y correlation_id reales.
	const qInsertOutbox = `
		INSERT INTO outbox_event (event_id, event_type, payload, correlation_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'PENDING', now(), now())
	`
	_, err = tx.Exec(ctx, qInsertOutbox, eventID, eventType, envelopePayload, correlationID)
	if err != nil {
		return nil, fmt.Errorf("insert outbox_event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return saved, nil
}
