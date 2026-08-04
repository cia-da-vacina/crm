package repository

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog/model"
	"github.com/cia-da-vacina/crm/backend/pkg/cursor"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

type Row struct {
	ID           string            `db:"id"`
	ActorUserID  *string           `db:"actor_user_id"`
	ActorName    *string           `db:"actor_name"`
	Action       string            `db:"action"`
	ResourceType string            `db:"resource_type"`
	ResourceID   string            `db:"resource_id"`
	UnitID       *string           `db:"unit_id"`
	Metadata     entity.JSONObject `db:"metadata"`
	CreatedAt    time.Time         `db:"created_at"`
}

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

// List é só-leitura (audit_logs é append-only, escrita fica em pkg/audit) —
// admin-only, sem escopo por unidade: auditoria existe pra fiscalizar
// atendentes/gestores, então filtrar por unit_id no query param é
// suficiente, não precisa de RBAC por unidade como o resto do sistema.
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

	var actionFilter, resourceTypeFilter, actorFilter, unitFilter *string
	if f.Action != "" {
		actionFilter = &f.Action
	}
	if f.ResourceType != "" {
		resourceTypeFilter = &f.ResourceType
	}
	if f.ActorUserID != "" {
		actorFilter = &f.ActorUserID
	}
	if f.UnitID != "" {
		unitFilter = &f.UnitID
	}

	var rows []Row
	err = r.db.SelectContext(ctx, &rows, `
		SELECT a.id, a.actor_user_id, u.name AS actor_name, a.action, a.resource_type, a.resource_id,
		       a.unit_id, a.metadata, a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.actor_user_id
		WHERE ($1::text IS NULL OR a.action = $1::text)
		  AND ($2::text IS NULL OR a.resource_type = $2::text)
		  AND ($3::uuid IS NULL OR a.actor_user_id = $3::uuid)
		  AND ($4::uuid IS NULL OR a.unit_id = $4::uuid)
		  AND ($5::boolean IS FALSE OR (a.created_at, a.id) < ($6::timestamptz, $7::uuid))
		ORDER BY a.created_at DESC, a.id DESC
		LIMIT $8
	`, actionFilter, resourceTypeFilter, actorFilter, unitFilter, hasCursor, cursorTime, cursorID, f.Limit+1)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(rows) > f.Limit
	if hasMore {
		rows = rows[:f.Limit]
	}
	return rows, hasMore, nil
}
