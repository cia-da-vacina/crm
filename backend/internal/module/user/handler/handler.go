package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	"github.com/cia-da-vacina/crm/backend/internal/module/user/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/user/usecase"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/go-chi/chi/v5"
)

type UseCase interface {
	Create(ctx context.Context, input usecase.CreateUserInput) (model.User, error)
	Get(ctx context.Context, id string) (model.User, error)
	List(ctx context.Context, input usecase.ListUsersInput) (model.ListUsersResult, error)
	Update(ctx context.Context, input usecase.UpdateUserInput) (model.User, error)
	Delete(ctx context.Context, id string, actorUserID string) error
	SetUnits(ctx context.Context, userID string, unitIDs []string) error
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

	page := parseIntOrDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntOrDefault(r.URL.Query().Get("page_size"), 20)

	result, err := h.uc.List(r.Context(), usecase.ListUsersInput{
		RequesterRole:    claims.Role,
		RequesterUnitIDs: claims.UnitIDs,
		Page:             page,
		PageSize:         pageSize,
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", result)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.uc.Get(r.Context(), id)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", user)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	var req model.CreateUserRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	user, err := h.uc.Create(r.Context(), usecase.CreateUserInput{
		Email:       req.Email,
		Password:    req.Password,
		Name:        req.Name,
		Role:        req.Role,
		UnitIDs:     req.UnitIDs,
		ActorUserID: claims.Subject,
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Created(w, user, "")
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	var req model.UpdateUserRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	user, err := h.uc.Update(r.Context(), usecase.UpdateUserInput{
		TargetID:      chi.URLParam(r, "id"),
		RequesterID:   claims.Subject,
		RequesterRole: claims.Role,
		Name:          req.Name,
		Role:          req.Role,
		Active:        req.Active,
		Password:      req.Password,
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", user)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	id := chi.URLParam(r, "id")

	if err := h.uc.Delete(r.Context(), id, claims.Subject); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SetUnits(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req model.SetUnitsRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	if err := h.uc.SetUnits(r.Context(), id, req.UnitIDs); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseIntOrDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return fallback
	}
	return v
}
