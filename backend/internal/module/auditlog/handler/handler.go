package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog/model"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type UseCase interface {
	List(ctx context.Context, f model.ListFilter) (model.CursorPage, error)
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := h.uc.List(r.Context(), model.ListFilter{
		Action:       q.Get("action"),
		ResourceType: q.Get("resource_type"),
		ActorUserID:  q.Get("actor_user_id"),
		UnitID:       q.Get("unit_id"),
		Cursor:       q.Get("cursor"),
		Limit:        parseIntOrDefault(q.Get("limit"), 30),
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", page)
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
