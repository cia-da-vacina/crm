package usecase

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/google/uuid"
)

const defaultTimezone = "America/Sao_Paulo"

type CreateUnitInput struct {
	Name       string
	Code       string
	City       string
	Address    string
	Timezone   string
	Active     *bool
	District   *string
	Complement *string
	Reference  *string
}

type UpdateUnitInput struct {
	ID         string
	Name       *string
	Code       *string
	City       *string
	Address    *string
	Timezone   *string
	Active     *bool
	District   *string
	Complement *string
	Reference  *string
}

type Repository interface {
	List(ctx context.Context, ids []string) ([]entity.Unit, error)
	GetByID(ctx context.Context, id string) (entity.Unit, error)
	Create(ctx context.Context, unit entity.Unit) error
	Update(ctx context.Context, unit entity.Unit) error
}

type UseCase struct {
	repo Repository
}

func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

// List: admin recebe todas as unidades; demais papéis só as próprias
// (requesterUnitIDs, vindo dos claims) — docs/BACKEND-CONTRACT.md §2.
func (uc *UseCase) List(ctx context.Context, isAdmin bool, requesterUnitIDs []string) ([]entity.Unit, error) {
	var ids []string // nil = sem filtro (admin)
	if !isAdmin {
		ids = requesterUnitIDs
		if ids == nil {
			ids = []string{}
		}
	}

	units, err := uc.repo.List(ctx, ids)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	return units, nil
}

func (uc *UseCase) Get(ctx context.Context, id string) (entity.Unit, error) {
	unit, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return entity.Unit{}, apperrors.NewNotFoundError("unit")
		}
		return entity.Unit{}, apperrors.NewDatabaseError(err)
	}
	return unit, nil
}

func (uc *UseCase) Create(ctx context.Context, input CreateUnitInput) (entity.Unit, error) {
	timezone := input.Timezone
	if timezone == "" {
		timezone = defaultTimezone
	}
	active := true
	if input.Active != nil {
		active = *input.Active
	}

	now := time.Now()
	unit := entity.Unit{
		ID:         uuid.Must(uuid.NewV7()).String(),
		Name:       input.Name,
		Code:       input.Code,
		Timezone:   timezone,
		Active:     active,
		Address:    input.Address,
		City:       input.City,
		District:   input.District,
		Complement: input.Complement,
		Reference:  input.Reference,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.repo.Create(ctx, unit); err != nil {
		return entity.Unit{}, apperrors.MapDBError(err, map[string]string{
			"units_code_unique": "code",
		})
	}
	return unit, nil
}

func (uc *UseCase) Update(ctx context.Context, input UpdateUnitInput) (entity.Unit, error) {
	unit, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return entity.Unit{}, apperrors.NewNotFoundError("unit")
		}
		return entity.Unit{}, apperrors.NewDatabaseError(err)
	}

	if input.Name != nil {
		unit.Name = *input.Name
	}
	if input.Code != nil {
		unit.Code = *input.Code
	}
	if input.City != nil {
		unit.City = *input.City
	}
	if input.Address != nil {
		unit.Address = *input.Address
	}
	if input.Timezone != nil {
		unit.Timezone = *input.Timezone
	}
	if input.Active != nil {
		unit.Active = *input.Active
	}
	if input.District != nil {
		unit.District = input.District
	}
	if input.Complement != nil {
		unit.Complement = input.Complement
	}
	if input.Reference != nil {
		unit.Reference = input.Reference
	}

	unit.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, unit); err != nil {
		return entity.Unit{}, apperrors.MapDBError(err, map[string]string{
			"units_code_unique": "code",
		})
	}
	return unit, nil
}
