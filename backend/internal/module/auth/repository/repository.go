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

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (entity.User, error) {
	var user entity.User
	if err := r.db.GetContext(ctx, &user, `
		SELECT id, email, password_hash, name, role, active, created_at, updated_at
		FROM users WHERE email = $1
	`, email); err != nil {
		return entity.User{}, err
	}
	return user, nil
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

// GetUnitIDsByUserID retorna as unidades vinculadas ao usuário — usado pra
// montar os claims do JWT. Não filtra por unit.active: um vínculo existente
// com uma unidade desativada ainda deve aparecer nos claims (a unidade some
// das listagens, não o vínculo do usuário).
func (r *Repository) GetUnitIDsByUserID(ctx context.Context, userID string) ([]string, error) {
	var ids []string
	if err := r.db.SelectContext(ctx, &ids, `
		SELECT unit_id FROM user_unit_relation WHERE user_id = $1
	`, userID); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *Repository) CreateRefreshToken(ctx context.Context, rt entity.RefreshToken) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, ip, user_agent)
		VALUES (:id, :user_id, :token_hash, :expires_at, :ip, :user_agent)
	`, rt)
	return err
}

func (r *Repository) GetRefreshTokenByHash(ctx context.Context, hash string) (entity.RefreshToken, error) {
	var rt entity.RefreshToken
	if err := r.db.GetContext(ctx, &rt, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, ip, user_agent, created_at
		FROM refresh_tokens WHERE token_hash = $1
	`, hash); err != nil {
		return entity.RefreshToken{}, err
	}
	return rt, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1`, id)
	return err
}

// RevokeAllUserRefreshTokens revoga todas as sessões ativas do usuário.
// É o mecanismo de logout: POST /auth/logout não recebe o refresh_token no
// corpo (só o Bearer do access token), então não há como revogar uma sessão
// específica — o contrato pede "denylist ou versão de sessão", e revogar
// tudo é a única leitura possível dado o input disponível (docs/BACKEND-CONTRACT.md §1).
func (r *Repository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}
