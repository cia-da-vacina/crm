package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Create insere o usuário e seus vínculos de unidade numa única transação —
// se a inserção de qualquer unit_id falhar (ex.: unidade inexistente), o
// usuário criado é revertido junto.
func (r *Repository) Create(ctx context.Context, user entity.User, unitIDs []string) error {
	return r.db.Transaction(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO users (id, email, password_hash, name, role, active, created_at, updated_at)
			VALUES (:id, :email, :password_hash, :name, :role, :active, :created_at, :updated_at)
		`, user); err != nil {
			return err
		}
		return insertUnitRelations(ctx, tx, user.ID, unitIDs)
	})
}

func (r *Repository) GetByID(ctx context.Context, id string) (entity.User, error) {
	var user entity.User
	if err := r.db.GetContext(ctx, &user, `
		SELECT id, email, password_hash, name, role, active, created_at, updated_at
		FROM users WHERE id = $1
	`, id); err != nil {
		return entity.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUnitIDs(ctx context.Context, userID string) ([]string, error) {
	var ids []string
	if err := r.db.SelectContext(ctx, &ids, `
		SELECT unit_id FROM user_unit_relation WHERE user_id = $1
	`, userID); err != nil {
		return nil, err
	}
	return ids, nil
}

// GetUnitIDsBulk busca os vínculos de vários usuários de uma vez (evita N+1
// em List) e agrupa em Go — mais simples e portável do que depender de
// array_agg + scan de array do Postgres via pgx/sqlx.
func (r *Repository) GetUnitIDsBulk(ctx context.Context, userIDs []string) (map[string][]string, error) {
	result := make(map[string][]string, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	type row struct {
		UserID string `db:"user_id"`
		UnitID string `db:"unit_id"`
	}
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT user_id, unit_id FROM user_unit_relation WHERE user_id = ANY($1)
	`, userIDs); err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], row.UnitID)
	}
	return result, nil
}

// List retorna uma página de usuários. Quando unscoped é false, só usuários
// com pelo menos um vínculo em scopeUnitIDs entram — usado pelo manager, que
// só enxerga "visão de unidade" (docs/PRODUCT-V2.md §2), nunca a base inteira.
func (r *Repository) List(ctx context.Context, unscoped bool, scopeUnitIDs []string, page, pageSize int) ([]entity.User, int, error) {
	offset := (page - 1) * pageSize

	if unscoped {
		var users []entity.User
		if err := r.db.SelectContext(ctx, &users, `
			SELECT id, email, password_hash, name, role, active, created_at, updated_at
			FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
		`, pageSize, offset); err != nil {
			return nil, 0, err
		}

		var total int
		if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM users`); err != nil {
			return nil, 0, err
		}
		return users, total, nil
	}

	if len(scopeUnitIDs) == 0 {
		return nil, 0, nil
	}

	var users []entity.User
	if err := r.db.SelectContext(ctx, &users, `
		SELECT DISTINCT u.id, u.email, u.password_hash, u.name, u.role, u.active, u.created_at, u.updated_at
		FROM users u
		JOIN user_unit_relation uur ON uur.user_id = u.id
		WHERE uur.unit_id = ANY($1)
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3
	`, scopeUnitIDs, pageSize, offset); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		JOIN user_unit_relation uur ON uur.user_id = u.id
		WHERE uur.unit_id = ANY($1)
	`, scopeUnitIDs); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update sobrescreve os campos mutáveis do usuário. O usecase decide o merge
// (campos ausentes no request mantêm o valor atual) e já seta UpdatedAt
// antes de chamar isto, pra a resposta devolvida ao client bater exatamente
// com o que foi persistido (evita ler NOW() de volta do banco).
func (r *Repository) Update(ctx context.Context, user entity.User) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE users
		SET name = :name, role = :role, active = :active, password_hash = :password_hash, updated_at = :updated_at
		WHERE id = :id
	`, user)
	return err
}

// SetUnits substitui integralmente o vínculo usuário×unidade (não incremental).
func (r *Repository) SetUnits(ctx context.Context, userID string, unitIDs []string) error {
	return r.db.Transaction(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM user_unit_relation WHERE user_id = $1`, userID); err != nil {
			return err
		}
		return insertUnitRelations(ctx, tx, userID, unitIDs)
	})
}

func insertUnitRelations(ctx context.Context, tx *sqlx.Tx, userID string, unitIDs []string) error {
	for _, unitID := range unitIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO user_unit_relation (id, user_id, unit_id) VALUES ($1, $2, $3)
		`, uuid.Must(uuid.NewV7()).String(), userID, unitID); err != nil {
			return err
		}
	}
	return nil
}
