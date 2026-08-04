// Package repository toca conversations/messages/customers/app_settings/
// ai_campaigns/pops diretamente — módulo autocontido, mesma convenção do
// resto do projeto (ver backend/ARCHITECTURE.md).
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

func (r *Repository) GetConversation(ctx context.Context, id string) (entity.Conversation, error) {
	var c entity.Conversation
	err := r.db.GetContext(ctx, &c, `
		SELECT id, customer_id, channel, channel_thread_id, unit_id, pipeline_stage, mode,
		       owner_id, intent, ai_summary, triage_notes, phone_gate,
		       loss_reason_code, loss_reason_text, window_expires_at,
		       last_message_preview, last_message_at, collected_fields,
		       triage_confidence, triage_ready_for_handoff, created_at, updated_at
		FROM conversations WHERE id = $1
	`, id)
	return c, err
}

func (r *Repository) UpdateConversation(ctx context.Context, c entity.Conversation) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE conversations SET
			intent = :intent, ai_summary = :ai_summary, triage_notes = :triage_notes,
			phone_gate = :phone_gate, collected_fields = :collected_fields,
			triage_confidence = :triage_confidence, triage_ready_for_handoff = :triage_ready_for_handoff,
			last_message_preview = :last_message_preview, last_message_at = :last_message_at,
			updated_at = :updated_at
		WHERE id = :id
	`, c)
	return err
}

func (r *Repository) GetCustomerIdentification(ctx context.Context, customerID string) (entity.CustomerIdentification, error) {
	var v entity.CustomerIdentification
	err := r.db.GetContext(ctx, &v, `SELECT identification FROM customers WHERE id = $1`, customerID)
	return v, err
}

func (r *Repository) GetCustomerIdentityExternalID(ctx context.Context, customerID string, channel entity.Channel) (string, error) {
	var externalID string
	err := r.db.GetContext(ctx, &externalID, `
		SELECT external_id FROM customer_identities WHERE customer_id = $1 AND channel = $2
	`, customerID, channel)
	return externalID, err
}

// GetRecentMessages devolve em ordem cronológica (mais antiga primeiro) —
// é assim que um histórico de chat faz sentido pro prompt do modelo, mesmo
// a query internamente buscando "as N mais recentes" via DESC + reverse.
func (r *Repository) GetRecentMessages(ctx context.Context, conversationID string, limit int) ([]entity.Message, error) {
	var messages []entity.Message
	if err := r.db.SelectContext(ctx, &messages, `
		SELECT id, conversation_id, direction, sender_type, sender_user_id, kind, channel,
		       body, status, meta_message_id, media_url, media_mime_type, template_name, created_at
		FROM messages WHERE conversation_id = $1
		ORDER BY created_at DESC LIMIT $2
	`, conversationID, limit); err != nil {
		return nil, err
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func (r *Repository) CreateMessage(ctx context.Context, msg entity.Message) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO messages (id, conversation_id, direction, sender_type, kind, channel, body, status, meta_message_id, created_at)
		VALUES (:id, :conversation_id, :direction, :sender_type, :kind, :channel, :body, :status, :meta_message_id, :created_at)
	`, msg)
	return err
}

func (r *Repository) GetAppSettings(ctx context.Context) (entity.AppSettings, error) {
	var s entity.AppSettings
	err := r.db.GetContext(ctx, &s, `
		SELECT id, ai_enabled, ai_system_prompt, ai_context, triage_enabled, triage_handoff_intents, default_unit_id, updated_at
		FROM app_settings WHERE id = 1
	`)
	return s, err
}

// GetActiveCampaigns só traz campanhas dentro da janela de datas e ativas —
// é contexto pro prompt, não precisa das encerradas/futuras.
func (r *Repository) GetActiveCampaigns(ctx context.Context) ([]entity.AICampaign, error) {
	var campaigns []entity.AICampaign
	err := r.db.SelectContext(ctx, &campaigns, `
		SELECT id, title, description, starts_on, ends_on, active, created_at, updated_at
		FROM ai_campaigns
		WHERE active AND starts_on <= CURRENT_DATE AND ends_on >= CURRENT_DATE
		ORDER BY starts_on
	`)
	return campaigns, err
}

func (r *Repository) GetSuggestedPopIDs(ctx context.Context, intent string) ([]string, error) {
	var ids []string
	err := r.db.SelectContext(ctx, &ids, `
		SELECT id FROM pops WHERE active AND intent_tags @> $1::jsonb ORDER BY title
	`, `["`+intent+`"]`)
	return ids, err
}

func (r *Repository) GetActivePendingVerification(ctx context.Context, conversationID string) (entity.PhoneVerification, error) {
	var pv entity.PhoneVerification
	err := r.db.GetContext(ctx, &pv, `
		SELECT id, conversation_id, phone_e164, code_hash, attempts, resend_count, expires_at, confirmed_at, created_at
		FROM phone_verifications WHERE conversation_id = $1 AND confirmed_at IS NULL
	`, conversationID)
	return pv, err
}
