package repository

import (
	"context"
	"encoding/json"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

const columns = `id, title, body, intent_tags, active, created_at, updated_at`

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

// List filtra por intent quando informado — intent_tags @> ["x"] testa se o
// array JSONB contém aquele elemento. Quando intent == "", o filtro é
// ignorado (o parâmetro ainda precisa ser um JSON válido pro cast não
// falhar, mesmo sem ser usado).
func (r *Repository) List(ctx context.Context, intent string) ([]entity.Pop, error) {
	filterJSON, err := json.Marshal([]string{intent})
	if err != nil {
		return nil, err
	}

	var pops []entity.Pop
	err = r.db.SelectContext(ctx, &pops, `
		SELECT `+columns+`
		FROM pops
		WHERE ($1 = '' OR intent_tags @> $2::jsonb)
		ORDER BY title
	`, intent, filterJSON)
	return pops, err
}

func (r *Repository) GetByID(ctx context.Context, id string) (entity.Pop, error) {
	var p entity.Pop
	err := r.db.GetContext(ctx, &p, `SELECT `+columns+` FROM pops WHERE id = $1`, id)
	return p, err
}

func (r *Repository) Create(ctx context.Context, p entity.Pop) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO pops (id, title, body, intent_tags, active, created_at, updated_at)
		VALUES (:id, :title, :body, :intent_tags, :active, :created_at, :updated_at)
	`, p)
	return err
}

func (r *Repository) Update(ctx context.Context, p entity.Pop) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE pops SET title = :title, body = :body, intent_tags = :intent_tags,
		                active = :active, updated_at = :updated_at
		WHERE id = :id
	`, p)
	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM pops WHERE id = $1`, id)
	return err
}
