package repository

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/followup/model"
	"github.com/cia-da-vacina/crm/backend/pkg/cursor"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

type Row struct {
	ID             string                `db:"id"`
	ConversationID string                `db:"conversation_id"`
	CustomerID     string                `db:"customer_id"`
	CustomerName   string                `db:"customer_name"`
	CustomerPhone  *string               `db:"customer_phone"`
	UnitID         string                `db:"unit_id"`
	PipelineStage  entity.PipelineStage  `db:"pipeline_stage"`
	DueAt          time.Time             `db:"due_at"`
	Status         entity.FollowUpStatus `db:"status"`
	Note           string                `db:"note"`
	CreatedAt      time.Time             `db:"created_at"`
	CompletedAt    *time.Time            `db:"completed_at"`
}

const rowColumns = `
	f.id, f.conversation_id, f.customer_id,
	cu.display_name AS customer_name, cu.primary_phone AS customer_phone,
	f.unit_id, f.pipeline_stage, f.due_at, f.status, f.note, f.created_at, f.completed_at
`

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

// List ordena por due_at ASC — é uma fila acionável (o que vence primeiro
// aparece primeiro), diferente do /inbox (que ordena por atividade recente).
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

	var unitFilter, statusFilter, stageFilter *string
	if f.UnitID != "" {
		unitFilter = &f.UnitID
	}
	if f.Status != "" {
		statusFilter = &f.Status
	}
	if f.Stage != "" {
		stageFilter = &f.Stage
	}

	var rows []Row
	err = r.db.SelectContext(ctx, &rows, `
		SELECT `+rowColumns+`
		FROM follow_ups f
		JOIN customers cu ON cu.id = f.customer_id
		WHERE ($1::uuid IS NULL OR f.unit_id = $1::uuid)
		  AND ($2::text IS NULL OR f.status = $2::text)
		  AND ($3::text IS NULL OR f.pipeline_stage = $3::text)
		  AND ($4 OR f.unit_id = ANY($5))
		  AND ($6::boolean IS FALSE OR (f.due_at, f.id) > ($7::timestamptz, $8::uuid))
		ORDER BY f.due_at ASC, f.id ASC
		LIMIT $9
	`, unitFilter, statusFilter, stageFilter, f.Unscoped, scopeIDs,
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
		FROM follow_ups f
		JOIN customers cu ON cu.id = f.customer_id
		WHERE f.id = $1
	`, id)
	return row, err
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status entity.FollowUpStatus, completedAt *time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE follow_ups SET status = $1, completed_at = $2 WHERE id = $3
	`, status, completedAt, id)
	return err
}
