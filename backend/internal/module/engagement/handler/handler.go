package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/go-chi/chi/v5"
)

type UseCase interface {
	List(ctx context.Context, f model.ListFilter) (model.CursorPage, error)
	Get(ctx context.Context, id string, access usecase.Access) (model.SocialEngagement, error)
	Reply(ctx context.Context, id string, req model.ReplyRequest, access usecase.Access) (model.SocialEngagement, error)
	Dismiss(ctx context.Context, id string, access usecase.Access) (model.SocialEngagement, error)
	Convert(ctx context.Context, id string, access usecase.Access) (model.SocialEngagement, error)
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
	page, err := h.uc.List(r.Context(), model.ListFilter{
		UnitID:           q.Get("unit_id"),
		Channel:          q.Get("channel"),
		Type:             q.Get("type"),
		Status:           q.Get("status"),
		Cursor:           q.Get("cursor"),
		Limit:            parseIntOrDefault(q.Get("limit"), 30),
		Unscoped:         claims.Role == string(entity.RoleAdmin),
		RequesterUnitIDs: claims.UnitIDs,
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", page)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	e, err := h.uc.Get(r.Context(), chi.URLParam(r, "id"), access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", e)
}

func (h *Handler) Reply(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	var req model.ReplyRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	e, err := h.uc.Reply(r.Context(), chi.URLParam(r, "id"), req, access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", e)
}

func (h *Handler) Dismiss(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	e, err := h.uc.Dismiss(r.Context(), chi.URLParam(r, "id"), access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", e)
}

func (h *Handler) Convert(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	e, err := h.uc.Convert(r.Context(), chi.URLParam(r, "id"), access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}
	httppkg.Success(w, http.StatusOK, "", e)
}

func accessFromContext(r *http.Request) (usecase.Access, bool) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		return usecase.Access{}, false
	}
	return usecase.Access{Role: claims.Role, UnitIDs: claims.UnitIDs}, true
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
