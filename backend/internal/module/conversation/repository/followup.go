package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
)

// CreateFollowUpIfNotOpen toca a tabela follow_ups diretamente em vez de
// depender do módulo followup — mesma convenção de módulo autocontido usada
// em customer.go (merge) e auth/repository (queries de users). O índice
// único parcial follow_ups_open_per_conversation_idx garante no banco que
// só existe um follow-up aberto por conversa; ON CONFLICT DO NOTHING deixa
// isso silencioso em vez de virar erro quando já existe um.
func (r *Repository) CreateFollowUpIfNotOpen(ctx context.Context, fu entity.FollowUp) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO follow_ups (id, conversation_id, customer_id, unit_id, pipeline_stage, due_at, status, note, created_at)
		VALUES (:id, :conversation_id, :customer_id, :unit_id, :pipeline_stage, :due_at, :status, :note, :created_at)
		ON CONFLICT (conversation_id) WHERE status = 'open' DO NOTHING
	`, fu)
	return err
}
