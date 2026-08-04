package usecase

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/domain/vo"
	"github.com/cia-da-vacina/crm/backend/internal/module/user/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/audit"
	"github.com/google/uuid"
)

type CreateUserInput struct {
	Email       string
	Password    string
	Name        string
	Role        string
	UnitIDs     []string
	ActorUserID string
}

// UpdateUserInput.RequesterID/RequesterRole vêm dos claims da request, nunca
// do body — é o que garante que um agente não consegue se promover a admin
// mandando role no PATCH do próprio usuário.
type UpdateUserInput struct {
	TargetID      string
	RequesterID   string
	RequesterRole string
	Name          *string
	Role          *string
	Active        *bool
	Password      *string
}

type ListUsersInput struct {
	RequesterRole    string
	RequesterUnitIDs []string
	Page             int
	PageSize         int
}

type Repository interface {
	Create(ctx context.Context, user entity.User, unitIDs []string) error
	GetByID(ctx context.Context, id string) (entity.User, error)
	GetUnitIDs(ctx context.Context, userID string) ([]string, error)
	GetUnitIDsBulk(ctx context.Context, userIDs []string) (map[string][]string, error)
	List(ctx context.Context, unscoped bool, scopeUnitIDs []string, page, pageSize int) ([]entity.User, int, error)
	Update(ctx context.Context, user entity.User) error
	SetUnits(ctx context.Context, userID string, unitIDs []string) error
}

type UseCase struct {
	repo  Repository
	audit *audit.Logger
}

func New(repo Repository, auditLogger *audit.Logger) *UseCase {
	return &UseCase{repo: repo, audit: auditLogger}
}

func (uc *UseCase) logAudit(ctx context.Context, actorUserID, action, resourceID string, metadata map[string]any) {
	if uc.audit == nil {
		return
	}
	uc.audit.Log(ctx, audit.Entry{
		ActorUserID: &actorUserID, Action: action, ResourceType: "user", ResourceID: resourceID, Metadata: metadata,
	})
}

func (uc *UseCase) Create(ctx context.Context, input CreateUserInput) (model.User, error) {
	hash, err := vo.HashPassword(input.Password, vo.DefaultPasswordConfig())
	if err != nil {
		return model.User{}, apperrors.NewInternalError(err)
	}

	now := time.Now()
	user := entity.User{
		ID:           uuid.Must(uuid.NewV7()).String(),
		Email:        input.Email,
		PasswordHash: hash,
		Name:         input.Name,
		Role:         entity.UserRole(input.Role),
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.repo.Create(ctx, user, input.UnitIDs); err != nil {
		return model.User{}, apperrors.MapDBError(err, map[string]string{
			"users_email_unique": "email",
		})
	}

	uc.logAudit(ctx, input.ActorUserID, "user.create", user.ID, map[string]any{"email": user.Email, "role": string(user.Role)})

	return toModel(user, input.UnitIDs), nil
}

func (uc *UseCase) Get(ctx context.Context, id string) (model.User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.User{}, apperrors.NewNotFoundError("user")
		}
		return model.User{}, apperrors.NewDatabaseError(err)
	}

	unitIDs, err := uc.repo.GetUnitIDs(ctx, id)
	if err != nil {
		return model.User{}, apperrors.NewDatabaseError(err)
	}

	return toModel(user, unitIDs), nil
}

// List aplica "visão de unidade" pra quem não é admin (docs/PRODUCT-V2.md
// §2: manager/supervisor têm visão de unidade, não da base inteira) — só
// admin enxerga usuários fora das próprias unidades.
func (uc *UseCase) List(ctx context.Context, input ListUsersInput) (model.ListUsersResult, error) {
	unscoped := input.RequesterRole == string(entity.RoleAdmin)

	users, total, err := uc.repo.List(ctx, unscoped, input.RequesterUnitIDs, input.Page, input.PageSize)
	if err != nil {
		return model.ListUsersResult{}, apperrors.NewDatabaseError(err)
	}

	ids := make([]string, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	unitIDsByUser, err := uc.repo.GetUnitIDsBulk(ctx, ids)
	if err != nil {
		return model.ListUsersResult{}, apperrors.NewDatabaseError(err)
	}

	items := make([]model.User, len(users))
	for i, u := range users {
		items[i] = toModel(u, unitIDsByUser[u.ID])
	}

	return model.ListUsersResult{Items: items, Total: total}, nil
}

// Update: admin pode alterar qualquer campo em qualquer usuário. Um usuário
// não-admin só pode alterar a si mesmo, e só name/password — role/active
// exigem admin (evita auto-promoção e auto-reativação/desativação).
func (uc *UseCase) Update(ctx context.Context, input UpdateUserInput) (model.User, error) {
	isAdmin := input.RequesterRole == string(entity.RoleAdmin)
	isSelf := input.RequesterID == input.TargetID

	if !isAdmin {
		if !isSelf {
			return model.User{}, apperrors.NewForbiddenError("")
		}
		if input.Role != nil || input.Active != nil {
			return model.User{}, apperrors.NewForbiddenError("only an admin can change role or active status")
		}
	}

	user, err := uc.repo.GetByID(ctx, input.TargetID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.User{}, apperrors.NewNotFoundError("user")
		}
		return model.User{}, apperrors.NewDatabaseError(err)
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Role != nil {
		user.Role = entity.UserRole(*input.Role)
	}
	if input.Active != nil {
		user.Active = *input.Active
	}
	if input.Password != nil {
		hash, err := vo.HashPassword(*input.Password, vo.DefaultPasswordConfig())
		if err != nil {
			return model.User{}, apperrors.NewInternalError(err)
		}
		user.PasswordHash = hash
	}

	user.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, user); err != nil {
		return model.User{}, apperrors.NewDatabaseError(err)
	}

	unitIDs, err := uc.repo.GetUnitIDs(ctx, user.ID)
	if err != nil {
		return model.User{}, apperrors.NewDatabaseError(err)
	}

	return toModel(user, unitIDs), nil
}

// Delete é soft-delete (active=false) — preserva o histórico de auditoria
// (docs/BACKEND-CONTRACT.md §2: "Preferir soft-delete a exclusão física").
func (uc *UseCase) Delete(ctx context.Context, id string, actorUserID string) error {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("user")
		}
		return apperrors.NewDatabaseError(err)
	}

	user.Active = false
	user.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, user); err != nil {
		return apperrors.NewDatabaseError(err)
	}

	uc.logAudit(ctx, actorUserID, "user.delete", user.ID, map[string]any{"email": user.Email})

	return nil
}

func (uc *UseCase) SetUnits(ctx context.Context, userID string, unitIDs []string) error {
	if _, err := uc.repo.GetByID(ctx, userID); err != nil {
		if apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("user")
		}
		return apperrors.NewDatabaseError(err)
	}

	if err := uc.repo.SetUnits(ctx, userID, unitIDs); err != nil {
		return apperrors.MapDBError(err, map[string]string{
			"user_unit_relation_unit_id_fkey": "unit",
		})
	}
	return nil
}

func toModel(user entity.User, unitIDs []string) model.User {
	return model.User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      string(user.Role),
		Active:    user.Active,
		UnitIDs:   unitIDs,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
