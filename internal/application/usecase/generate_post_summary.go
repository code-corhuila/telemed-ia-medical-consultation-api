package usecase

import (
	"encoding/json"
	"time"
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// GeneratePostSummary creates the post-consultation summary for a
// consultation that is already COMPLETED.
type GeneratePostSummary struct {
	consultations out.ConsultationRepositoryPort
	summaries     out.AttentionSummaryRepositoryPort
	postSummaries out.PostSummaryRepositoryPort
}

func NewGeneratePostSummary(
	consultations out.ConsultationRepositoryPort,
	summaries out.AttentionSummaryRepositoryPort,
	postSummaries out.PostSummaryRepositoryPort,
) *GeneratePostSummary {
	return &GeneratePostSummary{
		consultations: consultations,
		summaries:     summaries,
		postSummaries: postSummaries,
	}
}

func (uc *GeneratePostSummary) Execute(ctx context.Context, cmd dto.GeneratePostSummaryCommand) (*dto.PostSummaryResponse, error) {
	if cmd.ConsultationID == "" {
		return nil, exception.ErrInvalidPostSummary
	}
	c, err := uc.consultations.FindByID(ctx, cmd.ConsultationID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, exception.ErrConsultationNotFound
	}
	if c.Status != model.StatusCompleted {
		return nil, exception.ErrConsultationNotCompleted
	}

	att, err := uc.summaries.FindByConsultationID(ctx, cmd.ConsultationID)
	if err != nil {
		return nil, err
	}
	var attID *string
	if att != nil {
		attID = &att.ID
	}

	ps, err := model.NewPostSummary(c.ID, c.AppointmentID, c.PatientID, cmd.PreconsultationSummaryID, attID)
	if err != nil {
		return nil, exception.ErrInvalidPostSummary
	}
	// Construir el evento de outbox
	payload, _ := json.Marshal(map[string]interface{}{
		"summaryId":      ps.ID,
		"consultationId": ps.ConsultationID,
		"patientId":      ps.PatientID,
	})
	
	event := &model.OutboxEvent{
		EventType:     "PostConsultationSummaryGenerated",
		Payload:       payload,
		CorrelationID: "", // Se podría inyectar desde el contexto
		Status:        "PENDING",
		CreatedAt:     time.Now().UTC(),
	}

	// Guardar PostSummary y OutboxEvent en la misma transacción
	saved, err := uc.postSummaries.SaveWithOutbox(ctx, ps, event)
	if err != nil {
		return nil, err
	}
	return toPostSummaryResponse(saved), nil
}
