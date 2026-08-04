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

func (r *Repository) GetUserByID(ctx context.Context, id string) (entity.User, error) {
	var user entity.User
	if err := r.db.GetContext(ctx, &user, `
		SELECT id, email, password_hash, name, role, active, created_at, updated_at
		FROM users WHERE id = $1
	`, id); err != nil {
		return entity.User{}, err
	}
	return user, nil
}

const unqualifiedColumns = `id, name, code, timezone, active, address, city, district, complement, reference, created_at, updated_at`

// unitJoinColumns qualifica cada coluna com "u." — necessário no JOIN abaixo
// porque user_unit_relation também tem uma coluna "id" (sua própria PK), o
// que tornaria "id" ambíguo sem prefixo.
const unitJoinColumns = `u.id, u.name, u.code, u.timezone, u.active, u.address, u.city, u.district, u.complement, u.reference, u.created_at, u.updated_at`

// ListUnits: admin recebe todas as unidades (mesma regra de GET /units);
// demais papéis só as vinculadas em user_unit_relation.
func (r *Repository) ListUnits(ctx context.Context, isAdmin bool, userID string) ([]entity.Unit, error) {
	var units []entity.Unit
	if isAdmin {
		if err := r.db.SelectContext(ctx, &units, `SELECT `+unqualifiedColumns+` FROM units ORDER BY name`); err != nil {
			return nil, err
		}
		return units, nil
	}

	if err := r.db.SelectContext(ctx, &units, `
		SELECT `+unitJoinColumns+`
		FROM units u
		JOIN user_unit_relation uur ON uur.unit_id = u.id
		WHERE uur.user_id = $1
		ORDER BY u.name
	`, userID); err != nil {
		return nil, err
	}
	return units, nil
}
