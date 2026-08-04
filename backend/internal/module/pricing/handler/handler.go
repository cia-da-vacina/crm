package handler

import (
	"context"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/module/pricing/model"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/go-chi/chi/v5"
)

type UseCase interface {
	List(ctx context.Context) ([]model.Rate, error)
	UpdateRate(ctx context.Context, category string, req model.UpdateRateRequest) (model.Rate, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rates, err := h.uc.List(r.Context())
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", rates)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateRateRequest
	if err := httppkg.ParseAndValidateOptional(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	rate, err := h.uc.UpdateRate(r.Context(), chi.URLParam(r, "category"), req)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", rate)
}
