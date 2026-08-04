package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
)

// GetActivePendingVerification busca a pendência não confirmada da conversa
// (no máximo uma, ver índice único parcial na migration). sql.ErrNoRows
// (via apperrors.IsNotFound) significa "sem pendência ativa".
func (r *Repository) GetActivePendingVerification(ctx context.Context, conversationID string) (entity.PhoneVerification, error) {
	var pv entity.PhoneVerification
	err := r.db.GetContext(ctx, &pv, `
		SELECT id, conversation_id, phone_e164, code_hash, attempts, resend_count, expires_at, confirmed_at, created_at
		FROM phone_verifications
		WHERE conversation_id = $1 AND confirmed_at IS NULL
	`, conversationID)
	return pv, err
}

// UpsertPendingVerification cria a pendência ou substitui a existente — só
// pode haver uma não confirmada por conversa (índice único parcial), então
// "iniciar" com um novo número sempre sobrescreve qualquer pendência antiga
// em vez de acumular.
func (r *Repository) UpsertPendingVerification(ctx context.Context, pv entity.PhoneVerification) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO phone_verifications (id, conversation_id, phone_e164, code_hash, attempts, resend_count, expires_at)
		VALUES (:id, :conversation_id, :phone_e164, :code_hash, :attempts, :resend_count, :expires_at)
		ON CONFLICT (conversation_id) WHERE confirmed_at IS NULL
		DO UPDATE SET phone_e164 = :phone_e164, code_hash = :code_hash,
		              attempts = :attempts, resend_count = :resend_count, expires_at = :expires_at
	`, pv)
	return err
}

func (r *Repository) UpdatePendingVerification(ctx context.Context, pv entity.PhoneVerification) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE phone_verifications
		SET attempts = :attempts, resend_count = :resend_count, expires_at = :expires_at, confirmed_at = :confirmed_at
		WHERE id = :id
	`, pv)
	return err
}

// DeletePendingVerification é usada quando uma pendência estoura tentativas
// ou expira e volta pro estado "required" — não faz sentido manter a linha
// (o índice único parcial permitiria uma nova de qualquer forma, mas apagar
// evita lixo acumulado de tentativas abandonadas).
func (r *Repository) DeletePendingVerification(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM phone_verifications WHERE id = $1`, id)
	return err
}

// GetActivePendingVerificationsBulk busca a pendência ativa de várias
// conversas de uma vez — usado em List pra montar pending_phone_masked sem
// N+1 (mesma técnica de internal/module/user/repository).
func (r *Repository) GetActivePendingVerificationsBulk(ctx context.Context, conversationIDs []string) (map[string]entity.PhoneVerification, error) {
	result := make(map[string]entity.PhoneVerification, len(conversationIDs))
	if len(conversationIDs) == 0 {
		return result, nil
	}

	var rows []entity.PhoneVerification
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, conversation_id, phone_e164, code_hash, attempts, resend_count, expires_at, confirmed_at, created_at
		FROM phone_verifications
		WHERE conversation_id = ANY($1) AND confirmed_at IS NULL
	`, conversationIDs); err != nil {
		return nil, err
	}

	for _, pv := range rows {
		result[pv.ConversationID] = pv
	}
	return result, nil
}
