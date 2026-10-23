package usecase_test

import (
	"context"
	"fmt"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase/fakes"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// uuidFor builds a deterministic UUID string from an integer seed, so tests
// can keep using small ids (1, 2, 3) while satisfying the UUID validation.
func uuidFor(seed int64) string {
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", seed)
}

func seedConsultation(ctx context.Context, repo *fakes.ConsultationRepo, id string) error {
	c, err := model.NewConsultation(id, uuidFor(1), uuidFor(2))
	if err != nil {
		return err
	}
	_, err = repo.Save(ctx, c)
	return err
}
