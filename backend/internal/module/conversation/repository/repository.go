// Package repository acessa as tabelas conversations/messages/loss_reasons.
// phone_verifications tem seu próprio arquivo (phone.go) por ser um
// subdomínio conceitualmente separado (OTP), embora viva no mesmo módulo.
package repository

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/model"
	"github.com/cia-da-vacina/crm/backend/pkg/cursor"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

// Row é a projeção usada tanto em List quanto em GetByID — junta a
// conversation com os campos do customer que o ConversationSummary precisa
// (customer_name/phone/identification), mais unread_count computado em SQL.
type Row struct {
	ID                 string                        `db:"id"`
	CustomerID         string                        `db:"customer_id"`
	CustomerName       string                        `db:"customer_name"`
	CustomerPhone      *string                       `db:"customer_phone"`
	Identification     entity.CustomerIdentification `db:"identification"`
	Channel            entity.Channel                `db:"channel"`
	ChannelThreadID    *string                       `db:"channel_thread_id"`
	UnitID             string                        `db:"unit_id"`
	PipelineStage      entity.PipelineStage          `db:"pipeline_stage"`
	Mode               entity.ConversationMode       `db:"mode"`
	OwnerID            *string                       `db:"owner_id"`
	Intent             *string                       `db:"intent"`
	AISummary          *string                       `db:"ai_summary"`
	TriageNotes        *string                       `db:"triage_notes"`
	PhoneGate          entity.PhoneGate              `db:"phone_gate"`
	WindowExpiresAt    *time.Time                    `db:"window_expires_at"`
	LastMessagePreview string                        `db:"last_message_preview"`
	LastMessageAt      *time.Time                    `db:"last_message_at"`
	CreatedAt          time.Time                     `db:"created_at"`
	UpdatedAt          time.Time                     `db:"updated_at"`
	UnreadCount        int                           `db:"unread_count"`
	// SortAt = COALESCE(last_message_at, created_at) — chave de ordenação e
	// de cursor. Evita o caso degenerado de last_message_at NULL (conversa
	// sem nenhuma mensagem ainda) quebrar a paginação por cursor.
	SortAt time.Time `db:"sort_at"`
}

const rowColumns = `
	c.id, c.customer_id,
	cu.display_name AS customer_name,
	cu.primary_phone AS customer_phone,
	cu.identification,
	c.channel, c.channel_thread_id, c.unit_id, c.pipeline_stage, c.mode,
	c.owner_id, c.intent, c.ai_summary, c.triage_notes, c.phone_gate,
	c.window_expires_at, c.last_message_preview, c.last_message_at,
	c.created_at, c.updated_at,
	COALESCE(c.last_message_at, c.created_at) AS sort_at,
	COALESCE((
		SELECT COUNT(*) FROM messages m
		WHERE m.conversation_id = c.id AND m.direction = 'in'
		  AND m.created_at > COALESCE((
		      SELECT MAX(m2.created_at) FROM messages m2
		      WHERE m2.conversation_id = c.id AND m2.direction = 'out'
		  ), '-infinity'::timestamptz)
	), 0) AS unread_count
`

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

// List busca uma página + 1 registro extra (pra saber se há próxima página
// sem precisar de um COUNT separado — comum em paginação por cursor).
func (r *Repository) List(ctx context.Context, f model.InboxFilter) ([]Row, bool, error) {
	decodedTime, decodedID, err := cursor.Decode(f.Cursor)
	if err != nil {
		return nil, false, err
	}
	hasCursor := f.Cursor != ""

	// Passa ponteiros nil (NULL de verdade) quando não há cursor — um "" em
	// $9::uuid quebraria com erro de cast no Postgres antes mesmo do OR
	// avaliar, já que SQL não faz short-circuit de erro de tipo.
	var cursorTime *time.Time
	var cursorID *string
	if hasCursor {
		cursorTime = &decodedTime
		cursorID = &decodedID
	}

	scopeIDs := f.RequesterUnitIDs
	if scopeIDs == nil {
		scopeIDs = []string{}
	}

	var unitFilter *string
	if f.UnitID != "" {
		unitFilter = &f.UnitID
	}
	var stageFilter, channelFilter, modeFilter *string
	if f.Stage != "" {
		stageFilter = &f.Stage
	}
	if f.Channel != "" {
		channelFilter = &f.Channel
	}
	if f.Mode != "" {
		modeFilter = &f.Mode
	}

	var rows []Row
	err = r.db.SelectContext(ctx, &rows, `
		SELECT `+rowColumns+`
		FROM conversations c
		JOIN customers cu ON cu.id = c.customer_id
		WHERE ($1::uuid IS NULL OR c.unit_id = $1::uuid)
		  AND ($2::text IS NULL OR c.pipeline_stage = $2::text)
		  AND ($3::text IS NULL OR c.channel = $3::text)
		  AND ($4::text IS NULL OR c.mode = $4::text)
		  AND ($5 OR c.unit_id = ANY($6))
		  AND ($7::boolean IS FALSE OR (COALESCE(c.last_message_at, c.created_at), c.id) < ($8::timestamptz, $9::uuid))
		ORDER BY sort_at DESC, c.id DESC
		LIMIT $10
	`, unitFilter, stageFilter, channelFilter, modeFilter, f.Unscoped, scopeIDs,
		hasCursor, cursorTime, cursorID, f.Limit+1)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(rows) > f.Limit
	if hasMore {
		rows = rows[:f.Limit]
	}
	return rows, hasMore, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Row, error) {
	var row Row
	err := r.db.GetContext(ctx, &row, `
		SELECT `+rowColumns+`
		FROM conversations c
		JOIN customers cu ON cu.id = c.customer_id
		WHERE c.id = $1
	`, id)
	return row, err
}

// GetConversation busca a entity crua (sem join) — usado pelos usecases de
// mutação (claim, pipeline, mensagens, phone) que precisam só dos campos da
// própria conversation pra decidir e persistir.
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

// UpdateConversation cobre todos os campos mutáveis pelos usecases deste
// módulo — inclui intent/ai_summary/triage_notes/collected_fields (fase 7,
// triagem) mesmo esse módulo não sendo dono da lógica de IA: são colunas de
// conversations, e todo usecase que muta a linha usa este mesmo método.
func (r *Repository) UpdateConversation(ctx context.Context, c entity.Conversation) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE conversations SET
			customer_id = :customer_id, pipeline_stage = :pipeline_stage, mode = :mode,
			owner_id = :owner_id, phone_gate = :phone_gate,
			intent = :intent, ai_summary = :ai_summary, triage_notes = :triage_notes,
			collected_fields = :collected_fields, triage_confidence = :triage_confidence,
			triage_ready_for_handoff = :triage_ready_for_handoff,
			loss_reason_code = :loss_reason_code, loss_reason_text = :loss_reason_text,
			last_message_preview = :last_message_preview, last_message_at = :last_message_at,
			updated_at = :updated_at
		WHERE id = :id
	`, c)
	return err
}

func (r *Repository) ListMessages(ctx context.Context, conversationID string, before *time.Time, limit int) ([]entity.Message, bool, error) {
	var messages []entity.Message
	err := r.db.SelectContext(ctx, &messages, `
		SELECT id, conversation_id, direction, sender_type, sender_user_id, kind, channel,
		       body, status, meta_message_id, media_url, media_mime_type, template_name, created_at,
		       pricing_category, pricing_billable, pricing_model, pricing_confirmed, cost_brl
		FROM messages
		WHERE conversation_id = $1 AND ($2::timestamptz IS NULL OR created_at < $2::timestamptz)
		ORDER BY created_at DESC
		LIMIT $3
	`, conversationID, before, limit+1)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	return messages, hasMore, nil
}

func (r *Repository) CreateMessage(ctx context.Context, msg entity.Message) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO messages (id, conversation_id, direction, sender_type, sender_user_id, kind,
		                       channel, body, status, meta_message_id, created_at,
		                       pricing_category, pricing_billable, pricing_model, pricing_confirmed, cost_brl)
		VALUES (:id, :conversation_id, :direction, :sender_type, :sender_user_id, :kind,
		        :channel, :body, :status, :meta_message_id, :created_at,
		        :pricing_category, :pricing_billable, :pricing_model, :pricing_confirmed, :cost_brl)
	`, msg)
	return err
}

func (r *Repository) GetActiveLossReason(ctx context.Context, code string) (entity.LossReason, error) {
	var lr entity.LossReason
	err := r.db.GetContext(ctx, &lr, `SELECT code, label FROM loss_reasons WHERE code = $1 AND active`, code)
	return lr, err
}
