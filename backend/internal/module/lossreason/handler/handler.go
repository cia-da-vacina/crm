package handler

import (
	"context"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type UseCase interface {
	List(ctx context.Context) ([]entity.LossReason, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	reasons, err := h.uc.List(r.Context())
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", reasons)
}
