package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/cia-vacina/crm/backend/internal/auth"
	"github.com/cia-vacina/crm/backend/internal/crmdemo"
	"github.com/cia-vacina/crm/backend/internal/domain"
	"github.com/cia-vacina/crm/backend/internal/store"
	"github.com/google/uuid"
)

type Server struct {
	Store  store.Store
	Tokens *auth.TokenService
	CRM    *crmdemo.Store
}

func (s *Server) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	user, err := s.Store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Credenciais inválidas")
		return
	}
	if !user.Active {
		writeError(w, http.StatusForbidden, "forbidden", "Usuário inativo")
		return
	}
	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Credenciais inválidas")
		return
	}
	pair, err := s.Tokens.Issue(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao emitir token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.ExpiresIn,
		"user":          sanitizeUser(user),
	})
}

func (s *Server) Logout(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil || req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "refresh_token obrigatório")
		return
	}
	claims, err := s.Tokens.Parse(req.RefreshToken, "refresh")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Refresh inválido")
		return
	}
	user, err := s.Store.GetUserByID(r.Context(), claims.UserID)
	if err != nil || !user.Active {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Usuário inválido")
		return
	}
	pair, err := s.Tokens.Issue(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao emitir token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.ExpiresIn,
		"user":          sanitizeUser(user),
	})
}

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Token ausente")
		return
	}
	user, err := s.Store.GetUserByID(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Usuário não encontrado")
		return
	}
	units, err := s.Store.ListUnitsForUser(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao listar unidades")
		return
	}
	writeJSON(w, http.StatusOK, domain.MeResponse{User: sanitizeUser(user), Units: units})
}

func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
	items, err := s.Store.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao listar usuários")
		return
	}
	out := make([]domain.User, 0, len(items))
	for _, u := range items {
		out = append(out, sanitizeUser(u))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out, "total": len(out)})
}

type createUserRequest struct {
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Name     string      `json:"name"`
	Role     domain.Role `json:"role"`
	UnitIDs  []uuid.UUID `json:"unit_ids"`
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	if req.Email == "" || len(req.Password) < 8 || req.Name == "" || req.Role == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "Campos obrigatórios inválidos")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao hashear senha")
		return
	}
	user, err := s.Store.CreateUser(r.Context(), domain.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Role:         req.Role,
		Active:       true,
		UnitIDs:      req.UnitIDs,
	})
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "conflict", "Email já cadastrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao criar usuário")
		return
	}
	writeJSON(w, http.StatusCreated, sanitizeUser(user))
}

func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	user, err := s.Store.GetUserByID(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Usuário não encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao obter usuário")
		return
	}
	writeJSON(w, http.StatusOK, sanitizeUser(user))
}

type updateUserRequest struct {
	Name     *string      `json:"name"`
	Role     *domain.Role `json:"role"`
	Active   *bool        `json:"active"`
	Password *string      `json:"password"`
}

func (s *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	existing, err := s.Store.GetUserByID(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Usuário não encontrado")
		return
	}
	var req updateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Role != nil {
		existing.Role = *req.Role
	}
	if req.Active != nil {
		existing.Active = *req.Active
	}
	if req.Password != nil {
		if len(*req.Password) < 8 {
			writeError(w, http.StatusBadRequest, "bad_request", "Senha deve ter no mínimo 8 caracteres")
			return
		}
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao hashear senha")
			return
		}
		existing.PasswordHash = hash
	}
	updated, err := s.Store.UpdateUser(r.Context(), existing)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao atualizar usuário")
		return
	}
	writeJSON(w, http.StatusOK, sanitizeUser(updated))
}

func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	if self, ok := UserIDFromContext(r.Context()); ok && self == id {
		writeError(w, http.StatusBadRequest, "bad_request", "Você não pode remover a própria conta")
		return
	}
	if err := s.Store.DeleteUser(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Usuário não encontrado")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao remover usuário")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type setUnitsRequest struct {
	UnitIDs []uuid.UUID `json:"unit_ids"`
}

func (s *Server) SetUserUnits(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	var req setUnitsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	if err := s.Store.SetUserUnits(r.Context(), id, req.UnitIDs); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Usuário não encontrado")
		return
	} else if errors.Is(err, store.ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, "bad_request", "Unidade inválida")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao vincular unidades")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ListUnits(w http.ResponseWriter, r *http.Request) {
	items, err := s.Store.ListUnits(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao listar unidades")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type createUnitRequest struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Timezone string `json:"timezone"`
}

func (s *Server) CreateUnit(w http.ResponseWriter, r *http.Request) {
	var req createUnitRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Code) == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "name e code obrigatórios")
		return
	}
	unit, err := s.Store.CreateUnit(r.Context(), domain.Unit{
		Name:     req.Name,
		Code:     req.Code,
		Timezone: req.Timezone,
		Active:   true,
	})
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "conflict", "Código de unidade já existe")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao criar unidade")
		return
	}
	writeJSON(w, http.StatusCreated, unit)
}

func (s *Server) GetUnit(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("unitId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	unit, err := s.Store.GetUnitByID(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Unidade não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, unit)
}

type updateUnitRequest struct {
	Name     *string `json:"name"`
	Timezone *string `json:"timezone"`
	Active   *bool   `json:"active"`
}

func (s *Server) UpdateUnit(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("unitId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	existing, err := s.Store.GetUnitByID(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Unidade não encontrada")
		return
	}
	var req updateUnitRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Timezone != nil {
		existing.Timezone = *req.Timezone
	}
	if req.Active != nil {
		existing.Active = *req.Active
	}
	updated, err := s.Store.UpdateUnit(r.Context(), existing)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Falha ao atualizar unidade")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func sanitizeUser(u domain.User) domain.User {
	u.PasswordHash = ""
	return u
}
