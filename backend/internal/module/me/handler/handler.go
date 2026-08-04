package handler

import (
	"context"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/module/me/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type UseCase interface {
	Get(ctx context.Context, userID string) (model.Me, error)
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

	me, err := h.uc.Get(r.Context(), claims.Subject)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", me)
}
