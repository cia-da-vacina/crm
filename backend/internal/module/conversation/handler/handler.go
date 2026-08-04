package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/cia-da-vacina/crm/backend/pkg/sse"
	"github.com/go-chi/chi/v5"
)

type UseCase interface {
	List(ctx context.Context, f model.InboxFilter) (model.CursorPage[model.ConversationSummary], error)
	Get(ctx context.Context, id string, access usecase.Access) (model.ConversationDetail, error)
	ListMessages(ctx context.Context, conversationID, cursor string, limit int, access usecase.Access) (model.CursorPage[model.Message], error)
	SendMessage(ctx context.Context, conversationID string, req model.SendMessageRequest, access usecase.Access) (model.Message, error)
	Claim(ctx context.Context, conversationID string, access usecase.Access) (model.ConversationDetail, error)
	UpdatePipeline(ctx context.Context, conversationID string, req model.UpdatePipelineRequest, access usecase.Access) (model.ConversationDetail, error)
	InitiatePhoneVerification(ctx context.Context, conversationID string, req model.StartPhoneVerificationRequest, access usecase.Access) (model.ConversationDetail, error)
	ResendPhoneVerification(ctx context.Context, conversationID string, access usecase.Access) (model.ConversationDetail, error)
	ConfirmPhoneVerification(ctx context.Context, conversationID string, req model.ConfirmPhoneVerificationRequest, access usecase.Access) (model.ConversationDetail, error)
}

type Handler struct {
	uc  UseCase
	hub *sse.Hub
}

func New(uc UseCase, hub *sse.Hub) *Handler {
	return &Handler{uc: uc, hub: hub}
}

func (h *Handler) ListInbox(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	q := r.URL.Query()
	page, err := h.uc.List(r.Context(), model.InboxFilter{
		UnitID:           q.Get("unit_id"),
		Stage:            q.Get("stage"),
		Channel:          q.Get("channel"),
		Mode:             q.Get("mode"),
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

func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	isAdmin := claims.Role == string(entity.RoleAdmin)
	allowed := make(map[string]struct{}, len(claims.UnitIDs))
	for _, id := range claims.UnitIDs {
		allowed[id] = struct{}{}
	}

	ch, cancel := h.hub.Subscribe()
	defer cancel()

	sse.Serve(w, r, ch, func(unitID string) bool {
		if isAdmin {
			return true
		}
		_, ok := allowed[unitID]
		return ok
	})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	detail, err := h.uc.Get(r.Context(), chi.URLParam(r, "id"), access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", detail)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	q := r.URL.Query()
	page, err := h.uc.ListMessages(r.Context(), chi.URLParam(r, "id"), q.Get("cursor"), parseIntOrDefault(q.Get("limit"), 50), access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", page)
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	var req model.SendMessageRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	msg, err := h.uc.SendMessage(r.Context(), chi.URLParam(r, "id"), req, access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Created(w, msg, "")
}

func (h *Handler) Claim(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	detail, err := h.uc.Claim(r.Context(), chi.URLParam(r, "id"), access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", detail)
}

func (h *Handler) UpdatePipeline(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	var req model.UpdatePipelineRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	detail, err := h.uc.UpdatePipeline(r.Context(), chi.URLParam(r, "id"), req, access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", detail)
}

func (h *Handler) InitiatePhone(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	var req model.StartPhoneVerificationRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	detail, err := h.uc.InitiatePhoneVerification(r.Context(), chi.URLParam(r, "id"), req, access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusAccepted, "", detail)
}

func (h *Handler) ResendPhone(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	detail, err := h.uc.ResendPhoneVerification(r.Context(), chi.URLParam(r, "id"), access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusAccepted, "", detail)
}

func (h *Handler) ConfirmPhone(w http.ResponseWriter, r *http.Request) {
	access, ok := accessFromContext(r)
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	var req model.ConfirmPhoneVerificationRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	detail, err := h.uc.ConfirmPhoneVerification(r.Context(), chi.URLParam(r, "id"), req, access)
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", detail)
}

func accessFromContext(r *http.Request) (usecase.Access, bool) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		return usecase.Access{}, false
	}
	return usecase.Access{UserID: claims.Subject, Role: claims.Role, UnitIDs: claims.UnitIDs}, true
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
