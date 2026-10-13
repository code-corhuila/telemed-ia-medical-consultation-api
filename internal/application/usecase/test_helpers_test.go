package usecase_test

import (
	"context"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase/fakes"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

func seedConsultation(ctx context.Context, repo *fakes.ConsultationRepo, id int64) error {
	c, err := model.NewConsultation(id, 1, 1)
	if err != nil {
		return err
	}
	c.ID = id
	_, err = repo.Save(ctx, c)
	return err
}
