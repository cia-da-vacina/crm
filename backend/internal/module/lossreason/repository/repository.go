package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]entity.LossReason, error) {
	var reasons []entity.LossReason
	err := r.db.SelectContext(ctx, &reasons, `SELECT code, label FROM loss_reasons WHERE active ORDER BY label`)
	return reasons, err
}
