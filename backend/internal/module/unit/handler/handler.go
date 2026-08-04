package handler

import (
	"context"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	"github.com/cia-da-vacina/crm/backend/internal/module/unit/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/unit/usecase"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/go-chi/chi/v5"
)

type UseCase interface {
	List(ctx context.Context, isAdmin bool, requesterUnitIDs []string) ([]entity.Unit, error)
	Get(ctx context.Context, id string) (entity.Unit, error)
	Create(ctx context.Context, input usecase.CreateUnitInput) (entity.Unit, error)
	Update(ctx context.Context, input usecase.UpdateUnitInput) (entity.Unit, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	units, err := h.uc.List(r.Context(), claims.Role == string(entity.RoleAdmin), claims.UnitIDs)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", map[string]any{"items": units})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	unit, err := h.uc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", unit)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUnitRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	unit, err := h.uc.Create(r.Context(), usecase.CreateUnitInput{
		Name:       req.Name,
		Code:       req.Code,
		City:       req.City,
		Address:    req.Address,
		Timezone:   req.Timezone,
		Active:     req.Active,
		District:   req.District,
		Complement: req.Complement,
		Reference:  req.Reference,
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Created(w, unit, "")
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateUnitRequest
	if err := httppkg.ParseAndValidateOptional(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	unit, err := h.uc.Update(r.Context(), usecase.UpdateUnitInput{
		ID:         chi.URLParam(r, "id"),
		Name:       req.Name,
		Code:       req.Code,
		City:       req.City,
		Address:    req.Address,
		Timezone:   req.Timezone,
		Active:     req.Active,
		District:   req.District,
		Complement: req.Complement,
		Reference:  req.Reference,
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", unit)
}
