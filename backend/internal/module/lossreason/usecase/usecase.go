package usecase

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
)

type Repository interface {
	List(ctx context.Context) ([]entity.LossReason, error)
}

type UseCase struct {
	repo Repository
}

func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) List(ctx context.Context) ([]entity.LossReason, error) {
	reasons, err := uc.repo.List(ctx)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	return reasons, nil
}
