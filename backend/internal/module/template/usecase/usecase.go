// Package usecase implementa o CRUD do catálogo de templates
// (docs/WHATSAPP-2026-ADAPTATION-PLAN.md, Frente B) — inclui a validação de
// variáveis do guia (WhatsApp_API_Optimization_Guide.md §4, "Checklist de
// Aprovação de Templates": "Variáveis ({{1}}, {{2}}) sempre preenchidas,
// nunca vazias") como uma checagem de consistência entre VariableCount
// declarado e os placeholders {{n}} realmente presentes no body, feita aqui
// (não no banco) porque é regra de negócio, não constraint de schema.
package usecase

import (
	"context"
	"regexp"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/template/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/google/uuid"
)

var placeholderPattern = regexp.MustCompile(`\{\{\d+\}\}`)

type Repository interface {
	List(ctx context.Context, category string) ([]entity.MessageTemplate, error)
	GetByID(ctx context.Context, id string) (entity.MessageTemplate, error)
	Create(ctx context.Context, t entity.MessageTemplate) error
	Update(ctx context.Context, t entity.MessageTemplate) error
	Delete(ctx context.Context, id string) error
}

type UseCase struct {
	repo Repository
}

func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) List(ctx context.Context, category string) ([]model.Template, error) {
	templates, err := uc.repo.List(ctx, category)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	out := make([]model.Template, len(templates))
	for i, t := range templates {
		out[i] = toModel(t)
	}
	return out, nil
}

func (uc *UseCase) Get(ctx context.Context, id string) (model.Template, error) {
	t, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.Template{}, apperrors.NewNotFoundError("template")
		}
		return model.Template{}, apperrors.NewDatabaseError(err)
	}
	return toModel(t), nil
}

func (uc *UseCase) Create(ctx context.Context, req model.CreateTemplateRequest) (model.Template, error) {
	if err := validateVariableCount(req.Body, req.VariableCount); err != nil {
		return model.Template{}, err
	}

	language := req.LanguageCode
	if language == "" {
		language = "pt_BR"
	}
	now := time.Now()
	t := entity.MessageTemplate{
		ID:             uuid.Must(uuid.NewV7()).String(),
		Name:           req.Name,
		Category:       entity.TemplateCategory(req.Category),
		LanguageCode:   language,
		Body:           req.Body,
		VariableCount:  req.VariableCount,
		ApprovalStatus: entity.TemplateApprovalPending,
		Active:         true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := uc.repo.Create(ctx, t); err != nil {
		return model.Template{}, apperrors.MapDBError(err, map[string]string{
			"message_templates_name_language_unique_idx": "name",
		})
	}
	return toModel(t), nil
}

func (uc *UseCase) Update(ctx context.Context, id string, req model.UpdateTemplateRequest) (model.Template, error) {
	t, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.Template{}, apperrors.NewNotFoundError("template")
		}
		return model.Template{}, apperrors.NewDatabaseError(err)
	}

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Category != nil {
		t.Category = entity.TemplateCategory(*req.Category)
	}
	if req.LanguageCode != nil {
		t.LanguageCode = *req.LanguageCode
	}
	if req.Body != nil {
		t.Body = *req.Body
	}
	if req.VariableCount != nil {
		t.VariableCount = *req.VariableCount
	}
	if req.Body != nil || req.VariableCount != nil {
		if err := validateVariableCount(t.Body, t.VariableCount); err != nil {
			return model.Template{}, err
		}
	}
	if req.ApprovalStatus != nil {
		t.ApprovalStatus = entity.TemplateApprovalStatus(*req.ApprovalStatus)
	}
	if req.Active != nil {
		t.Active = *req.Active
	}
	t.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, t); err != nil {
		return model.Template{}, apperrors.MapDBError(err, map[string]string{
			"message_templates_name_language_unique_idx": "name",
		})
	}
	return toModel(t), nil
}

func (uc *UseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.repo.GetByID(ctx, id); err != nil {
		if apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("template")
		}
		return apperrors.NewDatabaseError(err)
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// validateVariableCount conta os placeholders {{n}} no body e exige que
// bata exatamente com VariableCount declarado — pega tanto "declarei 2 mas
// só uso {{1}}" quanto "uso {{1}} e {{2}} mas declarei 0", os dois casos que
// o guia aponta como causa comum de rejeição pela Meta.
func validateVariableCount(body string, declared int) error {
	found := len(placeholderPattern.FindAllString(body, -1))
	if found != declared {
		return apperrors.NewBadRequestError("variable_count does not match the number of {{n}} placeholders found in body")
	}
	return nil
}

func toModel(t entity.MessageTemplate) model.Template {
	return model.Template{
		ID:             t.ID,
		Name:           t.Name,
		Category:       string(t.Category),
		LanguageCode:   t.LanguageCode,
		Body:           t.Body,
		VariableCount:  t.VariableCount,
		ApprovalStatus: string(t.ApprovalStatus),
		Active:         t.Active,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}
