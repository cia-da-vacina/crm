package usecase

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/dashboard/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/dashboard/repository"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
)

type Repository interface {
	GetCounts(ctx context.Context, unscoped bool, unitIDs []string) (repository.Counts, error)
	GetAwaitingReply(ctx context.Context, unscoped bool, unitIDs []string) (int, error)
	GetOpenEngagements(ctx context.Context, unscoped bool, unitIDs []string) (int, error)
	GetByStage(ctx context.Context, unscoped bool, unitIDs []string) (map[string]int, error)
	GetByChannel(ctx context.Context, unscoped bool, unitIDs []string) (map[string]int, error)
	GetByIntent(ctx context.Context, unscoped bool, unitIDs []string) (map[string]int, error)
	GetClosedByChannel(ctx context.Context, unscoped bool, unitIDs []string) (map[string]int, error)
	GetNotClosedByChannel(ctx context.Context, unscoped bool, unitIDs []string) (map[string]int, error)
	GetFollowUpCounts(ctx context.Context, unscoped bool, unitIDs []string) (awaiting, overdue int, err error)
	GetUnitBreakdown(ctx context.Context, unitIDs []string) ([]repository.UnitRow, error)
	GetAwaitingFollowupByUnit(ctx context.Context, unitIDs []string) (map[string]int, error)
	ListAllUnitIDs(ctx context.Context) ([]string, error)
}

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

// Get monta o snapshot. unitIDParam vazio = visão consolidada de todas as
// unidades acessíveis (docs/BACKEND-CONTRACT.md §6); informado = escopa só
// aquela unidade nos números agregados — mas Units[] sempre lista todas as
// unidades acessíveis, com suas próprias métricas, independente do filtro.
func (uc *UseCase) Get(ctx context.Context, unitIDParam string, access Access) (model.Summary, error) {
	unscoped, unitIDs, err := uc.resolveScope(unitIDParam, access)
	if err != nil {
		return model.Summary{}, err
	}

	counts, err := uc.repo.GetCounts(ctx, unscoped, unitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}
	awaitingReply, err := uc.repo.GetAwaitingReply(ctx, unscoped, unitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}
	openEngagements, err := uc.repo.GetOpenEngagements(ctx, unscoped, unitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}
	byStage, err := uc.repo.GetByStage(ctx, unscoped, unitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}
	byChannel, err := uc.repo.GetByChannel(ctx, unscoped, unitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}
	byIntent, err := uc.repo.GetByIntent(ctx, unscoped, unitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}
	closedByChannel, err := uc.repo.GetClosedByChannel(ctx, unscoped, unitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}
	notClosedByChannel, err := uc.repo.GetNotClosedByChannel(ctx, unscoped, unitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}
	awaitingFollowup, overdueFollowups, err := uc.repo.GetFollowUpCounts(ctx, unscoped, unitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}

	// Units[] independe do filtro unit_id? — sempre as unidades acessíveis.
	breakdownUnitIDs := unitIDs
	if access.isAdmin() {
		breakdownUnitIDs, err = uc.repo.ListAllUnitIDs(ctx)
		if err != nil {
			return model.Summary{}, apperrors.NewDatabaseError(err)
		}
	} else if unitIDParam != "" {
		breakdownUnitIDs = access.UnitIDs
	}

	unitRows, err := uc.repo.GetUnitBreakdown(ctx, breakdownUnitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}
	awaitingByUnit, err := uc.repo.GetAwaitingFollowupByUnit(ctx, breakdownUnitIDs)
	if err != nil {
		return model.Summary{}, apperrors.NewDatabaseError(err)
	}

	units := make([]model.UnitSummary, len(unitRows))
	for i, row := range unitRows {
		units[i] = model.UnitSummary{
			UnitID:           row.UnitID,
			UnitName:         row.UnitName,
			Open:             row.Open,
			Closed:           row.Closed,
			NotClosed:        row.NotClosed,
			ConversionRate:   conversionRate(row.Closed, row.NotClosed),
			Unclaimed:        row.Unclaimed,
			AwaitingFollowup: awaitingByUnit[row.UnitID],
		}
	}

	return model.Summary{
		OpenConversations:  counts.OpenConversations,
		ByStage:            byStage,
		ByChannel:          byChannel,
		Closed:             counts.Closed,
		NotClosed:          counts.NotClosed,
		Decided:            counts.Closed + counts.NotClosed,
		ConversionRate:     conversionRate(counts.Closed, counts.NotClosed),
		AITriage:           counts.AITriage,
		Human:              counts.Human,
		Unclaimed:          counts.Unclaimed,
		AwaitingReply:      awaitingReply,
		StaleOpen:          counts.StaleOpen,
		AwaitingPhone:      counts.AwaitingPhone,
		WindowExpiring:     counts.WindowExpiring,
		AwaitingFollowup:   awaitingFollowup,
		OverdueFollowups:   overdueFollowups,
		OpenEngagements:    openEngagements,
		ByIntent:           byIntent,
		ClosedByChannel:    closedByChannel,
		NotClosedByChannel: notClosedByChannel,
		Units:              units,
	}, nil
}

func (uc *UseCase) resolveScope(unitIDParam string, access Access) (unscoped bool, unitIDs []string, err error) {
	if unitIDParam != "" {
		if !access.canAccessUnit(unitIDParam) {
			return false, nil, apperrors.NewNotFoundError("unit")
		}
		return false, []string{unitIDParam}, nil
	}
	if access.isAdmin() {
		return true, nil, nil
	}
	return false, access.UnitIDs, nil
}

func conversionRate(closed, notClosed int) float64 {
	decided := closed + notClosed
	if decided == 0 {
		return 0
	}
	return float64(closed) / float64(decided)
}
