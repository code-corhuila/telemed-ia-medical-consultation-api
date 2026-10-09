package persistence

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// AttentionRepo implements out.AttentionSummaryRepositoryPort on PostgreSQL.
//
// medications is stored as JSONB. The domain uses []MedicationInstruction,
// so the adapter serializes/deserializes it at the boundary.
type AttentionRepo struct {
	pool *pgxpool.Pool
}

func NewAttentionRepo(pool *pgxpool.Pool) *AttentionRepo {
	return &AttentionRepo{pool: pool}
}

const attentionColumns = `
	id, consultation_id, diagnosis, recommendations, medications,
	observations, referral, follow_up, created_at, updated_at
`

func (r *AttentionRepo) Save(ctx context.Context, s *model.AttentionSummary) (*model.AttentionSummary, error) {
	meds, err := json.Marshal(s.Medications)
	if err != nil {
		return nil, err
	}
	const q = `
		INSERT INTO attention_summaries
			(consultation_id, diagnosis, recommendations, medications,
			 observations, referral, follow_up, created_at, updated_at)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7, $8, $9)
		RETURNING ` + attentionColumns

	row := r.pool.QueryRow(ctx, q,
		s.ConsultationID, s.Diagnosis, s.Recommendations, meds,
		s.Observations, s.Referral, s.FollowUp, s.CreatedAt, s.UpdatedAt,
	)
	return scanAttention(row)
}

func (r *AttentionRepo) FindByID(ctx context.Context, id int64) (*model.AttentionSummary, error) {
	q := `SELECT ` + attentionColumns + ` FROM attention_summaries WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanAttention(row)
}

func (r *AttentionRepo) FindByConsultationID(ctx context.Context, consultationID int64) (*model.AttentionSummary, error) {
	q := `SELECT ` + attentionColumns + ` FROM attention_summaries WHERE consultation_id = $1`
	row := r.pool.QueryRow(ctx, q, consultationID)
	return scanAttention(row)
}

func scanAttention(row pgx.Row) (*model.AttentionSummary, error) {
	var s model.AttentionSummary
	var meds []byte
	err := row.Scan(
		&s.ID, &s.ConsultationID, &s.Diagnosis, &s.Recommendations, &meds,
		&s.Observations, &s.Referral, &s.FollowUp, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(meds) > 0 {
		if err := json.Unmarshal(meds, &s.Medications); err != nil {
			return nil, err
		}
	} else {
		s.Medications = []model.MedicationInstruction{}
	}
	return &s, nil
}
