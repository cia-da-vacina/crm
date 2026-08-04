package usecase

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/customer/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
)

type Repository interface {
	List(ctx context.Context, filter model.ListFilter) ([]entity.Customer, int, error)
	GetByID(ctx context.Context, id string) (entity.Customer, error)
	GetIdentities(ctx context.Context, customerID string) ([]entity.CustomerIdentity, error)
	GetIdentitiesBulk(ctx context.Context, customerIDs []string) (map[string][]entity.CustomerIdentity, error)
}

type UseCase struct {
	repo Repository
}

func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) List(ctx context.Context, filter model.ListFilter) (model.ListResult, error) {
	customers, total, err := uc.repo.List(ctx, filter)
	if err != nil {
		return model.ListResult{}, apperrors.NewDatabaseError(err)
	}

	ids := make([]string, len(customers))
	for i, c := range customers {
		ids[i] = c.ID
	}
	identitiesByCustomer, err := uc.repo.GetIdentitiesBulk(ctx, ids)
	if err != nil {
		return model.ListResult{}, apperrors.NewDatabaseError(err)
	}

	items := make([]model.Customer, len(customers))
	for i, c := range customers {
		items[i] = toModel(c, identitiesByCustomer[c.ID])
	}

	return model.ListResult{Items: items, Total: total}, nil
}

func (uc *UseCase) Get(ctx context.Context, id string) (model.Customer, error) {
	customer, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.Customer{}, apperrors.NewNotFoundError("customer")
		}
		return model.Customer{}, apperrors.NewDatabaseError(err)
	}

	identities, err := uc.repo.GetIdentities(ctx, id)
	if err != nil {
		return model.Customer{}, apperrors.NewDatabaseError(err)
	}

	return toModel(customer, identities), nil
}

func (uc *UseCase) GetIdentities(ctx context.Context, customerID string) ([]entity.CustomerIdentity, error) {
	if _, err := uc.repo.GetByID(ctx, customerID); err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.NewNotFoundError("customer")
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	identities, err := uc.repo.GetIdentities(ctx, customerID)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	return identities, nil
}

func toModel(c entity.Customer, identities []entity.CustomerIdentity) model.Customer {
	if identities == nil {
		identities = []entity.CustomerIdentity{}
	}
	return model.Customer{
		ID:             c.ID,
		DisplayName:    c.DisplayName,
		Identification: string(c.Identification),
		PrimaryPhone:   c.PrimaryPhone,
		UnitID:         c.UnitID,
		Identities:     identities,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
