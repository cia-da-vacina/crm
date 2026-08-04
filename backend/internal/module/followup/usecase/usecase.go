package usecase

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/followup/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/followup/repository"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/cursor"
)

type Repository interface {
	List(ctx context.Context, f model.ListFilter) ([]repository.Row, bool, error)
	GetByID(ctx context.Context, id string) (repository.Row, error)
	UpdateStatus(ctx context.Context, id string, status entity.FollowUpStatus, completedAt *time.Time) error
}

// Access replica o shape usado no módulo conversation — cada módulo define o
// seu, mantendo-os independentes (ver backend/ARCHITECTURE.md).
type Access struct {
	Role    string
	UnitIDs []string
}

func (a Access) isAdmin() bool { return a.Role == string(entity.RoleAdmin) }

func (a Access) canAccessUnit(unitID string) bool {
	if a.isAdmin() {
		return true
	}
	for _, id := range a.UnitIDs {
		if id == unitID {
			return true
		}
	}
	return false
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

	items := make([]model.FollowUp, len(rows))
	for i, row := range rows {
		items[i] = toModel(row)
	}

	var nextCursor *string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		c := cursor.Encode(last.DueAt, last.ID)
		nextCursor = &c
	}

	return model.CursorPage{Items: items, NextCursor: nextCursor}, nil
}

func (uc *UseCase) Complete(ctx context.Context, id string, access Access) (model.FollowUp, error) {
	return uc.setStatus(ctx, id, entity.FollowUpDone, access)
}

func (uc *UseCase) Cancel(ctx context.Context, id string, access Access) (model.FollowUp, error) {
	return uc.setStatus(ctx, id, entity.FollowUpCanceled, access)
}

func (uc *UseCase) setStatus(ctx context.Context, id string, status entity.FollowUpStatus, access Access) (model.FollowUp, error) {
	row, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.FollowUp{}, apperrors.NewNotFoundError("follow-up")
		}
		return model.FollowUp{}, apperrors.NewDatabaseError(err)
	}
	if !access.canAccessUnit(row.UnitID) {
		return model.FollowUp{}, apperrors.NewNotFoundError("follow-up")
	}

	now := time.Now()
	if err := uc.repo.UpdateStatus(ctx, id, status, &now); err != nil {
		return model.FollowUp{}, apperrors.NewDatabaseError(err)
	}

	row.Status = status
	row.CompletedAt = &now
	return toModel(row), nil
}

func toModel(row repository.Row) model.FollowUp {
	return model.FollowUp{
		ID:             row.ID,
		ConversationID: row.ConversationID,
		CustomerID:     row.CustomerID,
		CustomerName:   row.CustomerName,
		CustomerPhone:  row.CustomerPhone,
		UnitID:         row.UnitID,
		PipelineStage:  string(row.PipelineStage),
		DueAt:          row.DueAt,
		Status:         string(row.Status),
		Note:           row.Note,
		CreatedAt:      row.CreatedAt,
		CompletedAt:    row.CompletedAt,
	}
}
