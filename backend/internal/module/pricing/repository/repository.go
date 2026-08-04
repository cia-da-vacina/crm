// Package repository acessa message_pricing_rates — a tabela de preço local
// usada pra converter categoria de cobrança WhatsApp em BRL (ver migration
// 000021 e docs/WHATSAPP-2026-ADAPTATION-PLAN.md §2.2).
package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

const columns = `category, rate_brl, billable, updated_at`

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]entity.MessagePricingRate, error) {
	var rates []entity.MessagePricingRate
	err := r.db.SelectContext(ctx, &rates, `SELECT `+columns+` FROM message_pricing_rates ORDER BY category`)
	return rates, err
}

func (r *Repository) GetByCategory(ctx context.Context, category entity.PricingCategory) (entity.MessagePricingRate, error) {
	var rate entity.MessagePricingRate
	err := r.db.GetContext(ctx, &rate, `SELECT `+columns+` FROM message_pricing_rates WHERE category = $1`, category)
	return rate, err
}

func (r *Repository) Update(ctx context.Context, rate entity.MessagePricingRate) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE message_pricing_rates SET rate_brl = :rate_brl, billable = :billable, updated_at = :updated_at
		WHERE category = :category
	`, rate)
	return err
}
