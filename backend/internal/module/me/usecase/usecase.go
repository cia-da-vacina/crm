package usecase

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/me/model"
	usermodel "github.com/cia-da-vacina/crm/backend/internal/module/user/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
)

type Repository interface {
	GetUserByID(ctx context.Context, id string) (entity.User, error)
	ListUnits(ctx context.Context, isAdmin bool, userID string) ([]entity.Unit, error)
}

type UseCase struct {
	repo Repository
}

func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Get(ctx context.Context, userID string) (model.Me, error) {
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.Me{}, apperrors.NewUnauthorizedError("")
		}
		return model.Me{}, apperrors.NewDatabaseError(err)
	}

	units, err := uc.repo.ListUnits(ctx, user.Role == entity.RoleAdmin, userID)
	if err != nil {
		return model.Me{}, apperrors.NewDatabaseError(err)
	}

	unitIDs := make([]string, len(units))
	for i, u := range units {
		unitIDs[i] = u.ID
	}

	return model.Me{
		User: usermodel.User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      string(user.Role),
			Active:    user.Active,
			UnitIDs:   unitIDs,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Units: units,
	}, nil
}
