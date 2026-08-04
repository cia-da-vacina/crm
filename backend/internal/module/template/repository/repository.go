// Package repository acessa message_templates — o catálogo próprio de
// templates WhatsApp (docs/WHATSAPP-2026-ADAPTATION-PLAN.md, Frente B).
package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

const columns = `id, name, category, language_code, body, variable_count, approval_status, active, created_at, updated_at`

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, category string) ([]entity.MessageTemplate, error) {
	var templates []entity.MessageTemplate
	err := r.db.SelectContext(ctx, &templates, `
		SELECT `+columns+`
		FROM message_templates
		WHERE ($1 = '' OR category = $1)
		ORDER BY name
	`, category)
	return templates, err
}

func (r *Repository) GetByID(ctx context.Context, id string) (entity.MessageTemplate, error) {
	var t entity.MessageTemplate
	err := r.db.GetContext(ctx, &t, `SELECT `+columns+` FROM message_templates WHERE id = $1`, id)
	return t, err
}

// GetApprovedByNameAndLanguage é usado por quem for enviar um template
// (ex.: conversation/usecase no futuro, quando SendMessage passar a aceitar
// kind=template) pra validar que o nome existe no catálogo, está aprovado, e
// devolver a categoria real pra computar o custo — não confiar num
// template_name texto livre.
func (r *Repository) GetApprovedByNameAndLanguage(ctx context.Context, name, languageCode string) (entity.MessageTemplate, error) {
	var t entity.MessageTemplate
	err := r.db.GetContext(ctx, &t, `
		SELECT `+columns+`
		FROM message_templates
		WHERE name = $1 AND language_code = $2 AND approval_status = 'approved' AND active
	`, name, languageCode)
	return t, err
}

func (r *Repository) Create(ctx context.Context, t entity.MessageTemplate) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO message_templates (id, name, category, language_code, body, variable_count, approval_status, active, created_at, updated_at)
		VALUES (:id, :name, :category, :language_code, :body, :variable_count, :approval_status, :active, :created_at, :updated_at)
	`, t)
	return err
}

func (r *Repository) Update(ctx context.Context, t entity.MessageTemplate) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE message_templates SET name = :name, category = :category, language_code = :language_code,
		       body = :body, variable_count = :variable_count, approval_status = :approval_status,
		       active = :active, updated_at = :updated_at
		WHERE id = :id
	`, t)
	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM message_templates WHERE id = $1`, id)
	return err
}
