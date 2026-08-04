package usecase

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog/repository"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/cursor"
)

type Repository interface {
	List(ctx context.Context, f model.ListFilter) ([]repository.Row, bool, error)
}

type UseCase struct {
	repo Repository
}

func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) List(ctx context.Context, f model.ListFilter) (model.CursorPage, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 30
	}

	rows, hasMore, err := uc.repo.List(ctx, f)
	if err != nil {
		return model.CursorPage{}, apperrors.NewDatabaseError(err)
	}

	items := make([]model.AuditLog, len(rows))
	for i, row := range rows {
		items[i] = model.AuditLog{
			ID:           row.ID,
			ActorUserID:  row.ActorUserID,
			ActorName:    row.ActorName,
			Action:       row.Action,
			ResourceType: row.ResourceType,
			ResourceID:   row.ResourceID,
			UnitID:       row.UnitID,
			Metadata:     map[string]any(row.Metadata),
			CreatedAt:    row.CreatedAt,
		}
	}

	var nextCursor *string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		c := cursor.Encode(last.CreatedAt, last.ID)
		nextCursor = &c
	}

	return model.CursorPage{Items: items, NextCursor: nextCursor}, nil
}
