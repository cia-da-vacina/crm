package handler

import (
	"context"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/module/auth/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type UseCase interface {
	Login(ctx context.Context, input usecase.LoginInput) (usecase.LoginOutput, error)
	Refresh(ctx context.Context, input usecase.RefreshInput) (usecase.LoginOutput, error)
	Logout(ctx context.Context, userID string) error
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	output, err := h.uc.Login(r.Context(), usecase.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IP:        httppkg.ClientIP(r),
		UserAgent: httppkg.UserAgent(r),
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", output)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := httppkg.ParseAndValidate(r, &req); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	output, err := h.uc.Refresh(r.Context(), usecase.RefreshInput{
		RefreshToken: req.RefreshToken,
		IP:           httppkg.ClientIP(r),
		UserAgent:    httppkg.UserAgent(r),
	})
	if err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	httppkg.Success(w, http.StatusOK, "", output)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httppkg.Unauthorized(w, "")
		return
	}

	if err := h.uc.Logout(r.Context(), claims.Subject); err != nil {
		httppkg.Handle(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
