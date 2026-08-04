package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/customer/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/go-chi/chi/v5"
)

type UseCase interface {
	List(ctx context.Context, filter model.ListFilter) (model.ListResult, error)
	Get(ctx context.Context, id string) (model.Customer, error)
	GetIdentities(ctx context.Context, customerID string) ([]entity.CustomerIdentity, error)
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

	q := r.URL.Query()
	result, err := h.uc.List(r.Context(), model.ListFilter{
		Query:            q.Get("q"),
		Identification:   q.Get("identification"),
		UnitID:           q.Get("unit_id"),
		Unscoped:         claims.Role == string(entity.RoleAdmin),
		RequesterUnitIDs: claims.UnitIDs,
		Page:             parseIntOrDefault(q.Get("page"), 1),
		PageSize:         parseIntOrDefault(q.Get("page_size"), 20),
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", result)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	customer, err := h.uc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", customer)
}

func (h *Handler) GetIdentities(w http.ResponseWriter, r *http.Request) {
	identities, err := h.uc.GetIdentities(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", identities)
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
