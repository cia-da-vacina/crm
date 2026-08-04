package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

const selectColumns = `id, name, code, timezone, active, address, city, district, complement, reference, created_at, updated_at`

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

// List retorna todas as unidades quando ids é nil (admin — "recebe todas");
// caso contrário, só as unidades cujo id está em ids.
func (r *Repository) List(ctx context.Context, ids []string) ([]entity.Unit, error) {
	var units []entity.Unit

	if ids == nil {
		if err := r.db.SelectContext(ctx, &units, `SELECT `+selectColumns+` FROM units ORDER BY name`); err != nil {
			return nil, err
		}
		return units, nil
	}

	if len(ids) == 0 {
		return units, nil
	}

	if err := r.db.SelectContext(ctx, &units, `SELECT `+selectColumns+` FROM units WHERE id = ANY($1) ORDER BY name`, ids); err != nil {
		return nil, err
	}
	return units, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (entity.Unit, error) {
	var unit entity.Unit
	if err := r.db.GetContext(ctx, &unit, `SELECT `+selectColumns+` FROM units WHERE id = $1`, id); err != nil {
		return entity.Unit{}, err
	}
	return unit, nil
}

// Create/Update recebem created_at/updated_at já setados pelo usecase (em vez
// de depender do DEFAULT now() do Postgres) pra a struct devolvida na
// resposta HTTP bater exatamente com o que foi persistido.
func (r *Repository) Create(ctx context.Context, unit entity.Unit) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO units (id, name, code, timezone, active, address, city, district, complement, reference, created_at, updated_at)
		VALUES (:id, :name, :code, :timezone, :active, :address, :city, :district, :complement, :reference, :created_at, :updated_at)
	`, unit)
	return err
}

func (r *Repository) Update(ctx context.Context, unit entity.Unit) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE units SET
			name = :name, code = :code, timezone = :timezone, active = :active,
			address = :address, city = :city, district = :district,
			complement = :complement, reference = :reference, updated_at = :updated_at
		WHERE id = :id
	`, unit)
	return err
}
