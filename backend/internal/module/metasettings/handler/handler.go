package handler

import (
	"context"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/module/metasettings/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type UseCase interface {
	Get(ctx context.Context) (model.Settings, error)
	Update(ctx context.Context, req model.UpdateSettingsRequest, actorUserID string) (model.Settings, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := h.uc.Get(r.Context())
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", settings)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateSettingsRequest
	if err := httppkg.ParseAndValidateOptional(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	settings, err := h.uc.Update(r.Context(), req, claims.Subject)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", settings)
}
