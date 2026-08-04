package usecase

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/pop/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/google/uuid"
)

type CreatePopInput struct {
	Title      string
	Body       string
	IntentTags []string
	Active     *bool
}

type UpdatePopInput struct {
	ID         string
	Title      *string
	Body       *string
	IntentTags []string
	Active     *bool
}

type Repository interface {
	List(ctx context.Context, intent string) ([]entity.Pop, error)
	GetByID(ctx context.Context, id string) (entity.Pop, error)
	Create(ctx context.Context, p entity.Pop) error
	Update(ctx context.Context, p entity.Pop) error
	Delete(ctx context.Context, id string) error
}

type UseCase struct {
	repo Repository
}

func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) List(ctx context.Context, intent string) ([]model.Pop, error) {
	pops, err := uc.repo.List(ctx, intent)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	items := make([]model.Pop, len(pops))
	for i, p := range pops {
		items[i] = toModel(p)
	}
	return items, nil
}

func (uc *UseCase) Get(ctx context.Context, id string) (model.Pop, error) {
	p, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.Pop{}, apperrors.NewNotFoundError("pop")
		}
		return model.Pop{}, apperrors.NewDatabaseError(err)
	}
	return toModel(p), nil
}

func (uc *UseCase) Create(ctx context.Context, input CreatePopInput) (model.Pop, error) {
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	now := time.Now()

	p := entity.Pop{
		ID:         uuid.Must(uuid.NewV7()).String(),
		Title:      input.Title,
		Body:       input.Body,
		IntentTags: entity.IntentTags(input.IntentTags),
		Active:     active,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := uc.repo.Create(ctx, p); err != nil {
		return model.Pop{}, apperrors.NewDatabaseError(err)
	}
	return toModel(p), nil
}

func (uc *UseCase) Update(ctx context.Context, input UpdatePopInput) (model.Pop, error) {
	p, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.Pop{}, apperrors.NewNotFoundError("pop")
		}
		return model.Pop{}, apperrors.NewDatabaseError(err)
	}

	if input.Title != nil {
		p.Title = *input.Title
	}
	if input.Body != nil {
		p.Body = *input.Body
	}
	if input.IntentTags != nil {
		p.IntentTags = entity.IntentTags(input.IntentTags)
	}
	if input.Active != nil {
		p.Active = *input.Active
	}
	p.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, p); err != nil {
		return model.Pop{}, apperrors.NewDatabaseError(err)
	}
	return toModel(p), nil
}

func (uc *UseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.repo.GetByID(ctx, id); err != nil {
		if apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("pop")
		}
		return apperrors.NewDatabaseError(err)
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func toModel(p entity.Pop) model.Pop {
	tags := []string(p.IntentTags)
	if tags == nil {
		tags = []string{}
	}
	return model.Pop{
		ID:         p.ID,
		Title:      p.Title,
		Body:       p.Body,
		IntentTags: tags,
		Active:     p.Active,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}
