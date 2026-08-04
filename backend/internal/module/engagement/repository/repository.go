package repository

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/cursor"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

type Row struct {
	ID               string                  `db:"id"`
	CustomerID       *string                 `db:"customer_id"`
	CustomerName     *string                 `db:"customer_name"`
	Channel          entity.Channel          `db:"channel"`
	Type             entity.EngagementType   `db:"type"`
	Status           entity.EngagementStatus `db:"status"`
	UnitID           string                  `db:"unit_id"`
	MediaID          *string                 `db:"media_id"`
	MediaURL         *string                 `db:"media_url"`
	MediaCaption     *string                 `db:"media_caption"`
	Body             string                  `db:"body"`
	ExternalID       string                  `db:"external_id"`
	AuthorExternalID string                  `db:"author_external_id"`
	ConversationID   *string                 `db:"conversation_id"`
	CreatedAt        time.Time               `db:"created_at"`
	RepliedAt        *time.Time              `db:"replied_at"`
}

const rowColumns = `
	e.id, e.customer_id, cu.display_name AS customer_name,
	e.channel, e.type, e.status, e.unit_id, e.media_id, e.media_url, e.media_caption,
	e.body, e.external_id, e.author_external_id, e.conversation_id, e.created_at, e.replied_at
`

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, f model.ListFilter) ([]Row, bool, error) {
	decodedTime, decodedID, err := cursor.Decode(f.Cursor)
	if err != nil {
		return nil, false, err
	}
	hasCursor := f.Cursor != ""

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

	var unitFilter, channelFilter, typeFilter, statusFilter *string
	if f.UnitID != "" {
		unitFilter = &f.UnitID
	}
	if f.Channel != "" {
		channelFilter = &f.Channel
	}
	if f.Type != "" {
		typeFilter = &f.Type
	}
	if f.Status != "" {
		statusFilter = &f.Status
	}

	var rows []Row
	err = r.db.SelectContext(ctx, &rows, `
		SELECT `+rowColumns+`
		FROM social_engagements e
		LEFT JOIN customers cu ON cu.id = e.customer_id
		WHERE ($1::uuid IS NULL OR e.unit_id = $1::uuid)
		  AND ($2::text IS NULL OR e.channel = $2::text)
		  AND ($3::text IS NULL OR e.type = $3::text)
		  AND ($4::text IS NULL OR e.status = $4::text)
		  AND ($5 OR e.unit_id = ANY($6))
		  AND ($7::boolean IS FALSE OR (e.created_at, e.id) < ($8::timestamptz, $9::uuid))
		ORDER BY e.created_at DESC, e.id DESC
		LIMIT $10
	`, unitFilter, channelFilter, typeFilter, statusFilter, f.Unscoped, scopeIDs,
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

// CreateEngagement retorna created=false (sem erro) quando (channel,
// external_id) já existe — mesma idempotência de CreateMessage no webhook, a
// Meta reenvia webhooks não confirmados a tempo.
func (r *Repository) CreateEngagement(ctx context.Context, e entity.SocialEngagement) (bool, error) {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO social_engagements (id, customer_id, channel, type, status, unit_id, media_id, media_url, media_caption,
		                                 body, external_id, author_external_id, conversation_id, created_at, replied_at)
		VALUES (:id, :customer_id, :channel, :type, :status, :unit_id, :media_id, :media_url, :media_caption,
		        :body, :external_id, :author_external_id, :conversation_id, :created_at, :replied_at)
	`, e)
	if err != nil {
		if apperrors.IsUniqueViolation(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Row, error) {
	var row Row
	err := r.db.GetContext(ctx, &row, `
		SELECT `+rowColumns+`
		FROM social_engagements e
		LEFT JOIN customers cu ON cu.id = e.customer_id
		WHERE e.id = $1
	`, id)
	return row, err
}

func (r *Repository) GetEntity(ctx context.Context, id string) (entity.SocialEngagement, error) {
	var e entity.SocialEngagement
	err := r.db.GetContext(ctx, &e, `
		SELECT id, customer_id, channel, type, status, unit_id, media_id, media_url, media_caption,
		       body, external_id, author_external_id, conversation_id, created_at, replied_at
		FROM social_engagements WHERE id = $1
	`, id)
	return e, err
}

func (r *Repository) UpdateEngagement(ctx context.Context, e entity.SocialEngagement) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE social_engagements SET
			customer_id = :customer_id, status = :status, conversation_id = :conversation_id,
			replied_at = :replied_at
		WHERE id = :id
	`, e)
	return err
}

// GetCustomerIDByIdentity, CreateCustomer, CreateCustomerIdentity, FindConversation
// e CreateConversation duplicam o que webhook/repository já faz — mesma
// convenção de módulo autocontido (ver backend/ARCHITECTURE.md): Convert
// precisa da mesma resolução de identidade/conversa que a ingestão de
// mensagem, mas engagement não deve depender do módulo webhook.
func (r *Repository) GetCustomerIDByIdentity(ctx context.Context, channel entity.Channel, externalID string) (string, error) {
	var id string
	err := r.db.GetContext(ctx, &id, `
		SELECT customer_id FROM customer_identities WHERE channel = $1 AND external_id = $2
	`, channel, externalID)
	return id, err
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
