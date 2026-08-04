// Package audit grava o log append-only de ações sensíveis exigido por
// docs/BACKEND-CONTRACT.md §9 (claim, mudança de pipeline_stage, alteração
// de settings/meta, criação/remoção de usuário). É infra cross-cutting
// chamada diretamente pelos usecases que precisam auditar algo — não um
// middleware genérico, porque só o usecase sabe o que de fato mudou
// (ex.: de qual stage pra qual stage), não só que a rota foi chamada.
package audit

import (
	"context"
	"log"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
	"github.com/google/uuid"
)

// Entry é o que o caller monta; Logger preenche ID/CreatedAt.
type Entry struct {
	ActorUserID  *string
	Action       string // "<recurso>.<verbo>", ex.: "conversation.claim"
	ResourceType string
	ResourceID   string
	UnitID       *string
	Metadata     map[string]any
}

type Logger struct {
	db *database.DB
}

func New(db *database.DB) *Logger {
	return &Logger{db: db}
}

// Log grava o registro e nunca retorna erro pro caller: uma falha ao
// auditar não deve derrubar a ação de negócio que está sendo auditada (o
// claim/mudança de pipeline já aconteceu quando isso é chamado) — só fica
// logada pra investigação manual.
func (l *Logger) Log(ctx context.Context, e Entry) {
	row := entity.AuditLog{
		ID:           uuid.Must(uuid.NewV7()).String(),
		ActorUserID:  e.ActorUserID,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		UnitID:       e.UnitID,
		Metadata:     entity.JSONObject(e.Metadata),
		CreatedAt:    time.Now(),
	}
	if _, err := l.db.NamedExecContext(ctx, `
		INSERT INTO audit_logs (id, actor_user_id, action, resource_type, resource_id, unit_id, metadata, created_at)
		VALUES (:id, :actor_user_id, :action, :resource_type, :resource_id, :unit_id, :metadata, :created_at)
	`, row); err != nil {
		log.Printf("audit: failed to log action=%s resource=%s/%s: %v", e.Action, e.ResourceType, e.ResourceID, err)
	}
}
