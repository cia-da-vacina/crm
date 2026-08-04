package handler

import (
	"context"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/module/template/model"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/go-chi/chi/v5"
)

type UseCase interface {
	List(ctx context.Context, category string) ([]model.Template, error)
	Get(ctx context.Context, id string) (model.Template, error)
	Create(ctx context.Context, req model.CreateTemplateRequest) (model.Template, error)
	Update(ctx context.Context, id string, req model.UpdateTemplateRequest) (model.Template, error)
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	templates, err := h.uc.List(r.Context(), r.URL.Query().Get("category"))
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", templates)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	t, err := h.uc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", t)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTemplateRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	t, err := h.uc.Create(r.Context(), req)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Created(w, t, "")
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateTemplateRequest
	if err := httppkg.ParseAndValidateOptional(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	t, err := h.uc.Update(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", t)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
