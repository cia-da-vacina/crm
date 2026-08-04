// Package usecase implementa a leitura/edição do rate card local (Frente A
// do plano de adaptação WhatsApp 2026) e expõe GetRate — consumido por
// outros módulos (conversation, webhook) através da interface PricingReader
// que eles próprios declaram, mesma convenção de CustomerReader/Triage/
// Engagement já usada no projeto pra dependência cross-módulo sem acoplar
// pacotes inteiros.
package usecase

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/pricing/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
)

type Repository interface {
	List(ctx context.Context) ([]entity.MessagePricingRate, error)
	GetByCategory(ctx context.Context, category entity.PricingCategory) (entity.MessagePricingRate, error)
	Update(ctx context.Context, rate entity.MessagePricingRate) error
}

type UseCase struct {
	repo Repository
}

func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) List(ctx context.Context) ([]model.Rate, error) {
	rates, err := uc.repo.List(ctx)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	out := make([]model.Rate, len(rates))
	for i, r := range rates {
		out[i] = toModel(r)
	}
	return out, nil
}

// GetRate é o método consumido via PricingReader por conversation/webhook
// pra computar Message.CostBRL no momento do envio/reconciliação — devolve
// a entity crua (não o model de API), já que quem chama não fala HTTP.
func (uc *UseCase) GetRate(ctx context.Context, category entity.PricingCategory) (entity.MessagePricingRate, error) {
	rate, err := uc.repo.GetByCategory(ctx, category)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return entity.MessagePricingRate{}, apperrors.NewNotFoundError("pricing rate")
		}
		return entity.MessagePricingRate{}, apperrors.NewDatabaseError(err)
	}
	return rate, nil
}

func (uc *UseCase) UpdateRate(ctx context.Context, category string, req model.UpdateRateRequest) (model.Rate, error) {
	cat := entity.PricingCategory(category)
	if !entity.ValidPricingCategories[cat] {
		return model.Rate{}, apperrors.NewBadRequestError("invalid pricing category")
	}

	rate, err := uc.repo.GetByCategory(ctx, cat)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.Rate{}, apperrors.NewNotFoundError("pricing rate")
		}
		return model.Rate{}, apperrors.NewDatabaseError(err)
	}

	if req.RateBRL != nil {
		rate.RateBRL = *req.RateBRL
	}
	if req.Billable != nil {
		rate.Billable = *req.Billable
	}
	rate.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, rate); err != nil {
		return model.Rate{}, apperrors.NewDatabaseError(err)
	}
	return toModel(rate), nil
}

func toModel(r entity.MessagePricingRate) model.Rate {
	return model.Rate{
		Category:  string(r.Category),
		RateBRL:   r.RateBRL,
		Billable:  r.Billable,
		UpdatedAt: r.UpdatedAt,
	}
}
