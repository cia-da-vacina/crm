// Package repository toca customers/customer_identities/conversations/
// messages/meta_channel_configs/app_settings diretamente — não depende de
// nenhum outro módulo (mesma convenção de módulo autocontido usada em todo
// o projeto, ver backend/ARCHITECTURE.md). É o repository com mais tabelas
// tocadas do projeto porque ingestão de webhook é genuinamente cross-
// cutting: resolve identidade, unidade, conversa e mensagem numa tacada só.
package repository

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetCustomerIDByIdentity(ctx context.Context, channel entity.Channel, externalID string) (string, error) {
	var id string
	err := r.db.GetContext(ctx, &id, `
		SELECT customer_id FROM customer_identities WHERE channel = $1 AND external_id = $2
	`, channel, externalID)
	return id, err
}

func (r *Repository) GetCustomerByPhone(ctx context.Context, phone string) (entity.Customer, error) {
	var c entity.Customer
	err := r.db.GetContext(ctx, &c, `
		SELECT id, display_name, identification, primary_phone, unit_id, created_at, updated_at
		FROM customers WHERE primary_phone = $1
	`, phone)
	return c, err
}

func (r *Repository) CreateCustomer(ctx context.Context, c entity.Customer) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO customers (id, display_name, identification, primary_phone, unit_id, created_at, updated_at)
		VALUES (:id, :display_name, :identification, :primary_phone, :unit_id, :created_at, :updated_at)
	`, c)
	return err
}

func (r *Repository) CreateCustomerIdentity(ctx context.Context, ci entity.CustomerIdentity) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO customer_identities (id, customer_id, channel, external_id, display_handle, phone_e164, verified_at, created_at)
		VALUES (:id, :customer_id, :channel, :external_id, :display_handle, :phone_e164, :verified_at, :created_at)
	`, ci)
	return err
}

// GetUnitIDByPhoneNumberID é o roteamento automático do WhatsApp (D-01: 1
// número por unidade) — o payload do webhook já traz o phone_number_id que
// recebeu a mensagem.
func (r *Repository) GetUnitIDByPhoneNumberID(ctx context.Context, phoneNumberID string) (string, error) {
	var unitID string
	err := r.db.GetContext(ctx, &unitID, `
		SELECT unit_id FROM meta_channel_configs WHERE channel = 'whatsapp' AND phone_number_id = $1
	`, phoneNumberID)
	return unitID, err
}

// GetTriageEnabled decide o mode de conversa nova: false força mode:"human"
// direto, pulando a IA por completo, sem quebrar o resto do fluxo
// (docs/PRODUCT-V2.md §6, kill-switch).
func (r *Repository) GetTriageEnabled(ctx context.Context) (bool, error) {
	var enabled bool
	err := r.db.GetContext(ctx, &enabled, `SELECT triage_enabled FROM app_settings WHERE id = 1`)
	return enabled, err
}

// GetDefaultUnitID é pra onde cai conversa nova de canal centralizado
// (Instagram/Facebook) — decisão confirmada com o stakeholder, ver
// backend/ARCHITECTURE.md §5. Pode ser nil se o admin ainda não configurou.
func (r *Repository) GetDefaultUnitID(ctx context.Context) (*string, error) {
	var row struct {
		DefaultUnitID *string `db:"default_unit_id"`
	}
	if err := r.db.GetContext(ctx, &row, `SELECT default_unit_id FROM app_settings WHERE id = 1`); err != nil {
		return nil, err
	}
	return row.DefaultUnitID, nil
}

// FindConversation reusa a thread mais recente do par (customer, channel) em
// vez de criar uma nova a cada mensagem — mesmo cliente mandando várias
// mensagens seguidas no mesmo canal continua na mesma conversa.
func (r *Repository) FindConversation(ctx context.Context, customerID string, channel entity.Channel) (entity.Conversation, error) {
	var c entity.Conversation
	err := r.db.GetContext(ctx, &c, `
		SELECT id, customer_id, channel, channel_thread_id, unit_id, pipeline_stage, mode, owner_id,
		       intent, ai_summary, triage_notes, phone_gate, loss_reason_code, loss_reason_text,
		       window_expires_at, last_message_preview, last_message_at, created_at, updated_at
		FROM conversations WHERE customer_id = $1 AND channel = $2
		ORDER BY created_at DESC LIMIT 1
	`, customerID, channel)
	return c, err
}

func (r *Repository) CreateConversation(ctx context.Context, c entity.Conversation) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO conversations (id, customer_id, channel, unit_id, pipeline_stage, mode, phone_gate,
		                            last_message_preview, last_message_at, created_at, updated_at)
		VALUES (:id, :customer_id, :channel, :unit_id, :pipeline_stage, :mode, :phone_gate,
		        :last_message_preview, :last_message_at, :created_at, :updated_at)
	`, c)
	return err
}

func (r *Repository) UpdateConversationAfterMessage(ctx context.Context, id, preview string, at, windowExpiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE conversations SET last_message_preview = $1, last_message_at = $2, window_expires_at = $3, updated_at = $2
		WHERE id = $4
	`, preview, at, windowExpiresAt, id)
	return err
}

// CreateMessage retorna created=false (sem erro) quando meta_message_id já
// existe — é a idempotência exigida pelo contrato (a Meta reenvia webhooks
// que não foram confirmados a tempo).
func (r *Repository) CreateMessage(ctx context.Context, msg entity.Message) (bool, error) {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO messages (id, conversation_id, direction, sender_type, kind, channel, body, status, meta_message_id, created_at)
		VALUES (:id, :conversation_id, :direction, :sender_type, :kind, :channel, :body, :status, :meta_message_id, :created_at)
	`, msg)
	if err != nil {
		if apperrors.IsUniqueViolation(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
