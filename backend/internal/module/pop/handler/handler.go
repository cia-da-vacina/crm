package handler

import (
	"context"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/module/pop/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/pop/usecase"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/go-chi/chi/v5"
)

type UseCase interface {
	List(ctx context.Context, intent string) ([]model.Pop, error)
	Get(ctx context.Context, id string) (model.Pop, error)
	Create(ctx context.Context, input usecase.CreatePopInput) (model.Pop, error)
	Update(ctx context.Context, input usecase.UpdatePopInput) (model.Pop, error)
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	pops, err := h.uc.List(r.Context(), r.URL.Query().Get("intent"))
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", pops)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	pop, err := h.uc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", pop)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePopRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	pop, err := h.uc.Create(r.Context(), usecase.CreatePopInput{
		Title:      req.Title,
		Body:       req.Body,
		IntentTags: req.IntentTags,
		Active:     req.Active,
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Created(w, pop, "")
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdatePopRequest
	if err := httppkg.ParseAndValidateOptional(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	pop, err := h.uc.Update(r.Context(), usecase.UpdatePopInput{
		ID:         chi.URLParam(r, "id"),
		Title:      req.Title,
		Body:       req.Body,
		IntentTags: req.IntentTags,
		Active:     req.Active,
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", pop)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
