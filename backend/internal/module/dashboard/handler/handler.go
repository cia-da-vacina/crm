package handler

import (
	"context"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/module/dashboard/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/dashboard/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type UseCase interface {
	Get(ctx context.Context, unitIDParam string, access usecase.Access) (model.Summary, error)
	GetCosts(ctx context.Context, unitIDParam string, access usecase.Access) (model.CostSummary, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	summary, err := h.uc.Get(r.Context(), r.URL.Query().Get("unit_id"), usecase.Access{
		Role:    claims.Role,
		UnitIDs: claims.UnitIDs,
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", summary)
}

func (h *Handler) GetCosts(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	summary, err := h.uc.GetCosts(r.Context(), r.URL.Query().Get("unit_id"), usecase.Access{
		Role:    claims.Role,
		UnitIDs: claims.UnitIDs,
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", summary)
}
