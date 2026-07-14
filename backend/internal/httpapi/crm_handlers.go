package httpapi

import (
	"errors"
	"net/http"

	"github.com/cia-vacina/crm/backend/internal/crmdemo"
	"github.com/cia-vacina/crm/backend/internal/domain"
	"github.com/google/uuid"
)

func (s *Server) ListInbox(w http.ResponseWriter, r *http.Request) {
	var unitID *uuid.UUID
	if v := r.URL.Query().Get("unit_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "unit_id inválido")
			return
		}
		unitID = &id
	}
	var stage *domain.PipelineStage
	if v := r.URL.Query().Get("stage"); v != "" {
		st := domain.PipelineStage(v)
		stage = &st
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": s.CRM.ListInbox(unitID, stage), "next_cursor": nil})
}

func (s *Server) GetConversation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("conversationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	c, ok := s.CRM.GetConversation(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Conversa não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) ListMessages(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("conversationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": s.CRM.ListMessages(id)})
}

func (s *Server) SendMessage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("conversationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	msg, err := s.CRM.SendMessage(id, req.Body)
	if errors.Is(err, crmdemo.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Conversa não encontrada")
		return
	}
	if errors.Is(err, crmdemo.ErrForbidden) {
		writeError(w, http.StatusForbidden, "forbidden", "Assuma a conversa antes de responder")
		return
	}
	if errors.Is(err, crmdemo.ErrBadRequest) {
		writeError(w, http.StatusBadRequest, "bad_request", "Mensagem vazia")
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

func (s *Server) ClaimConversation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("conversationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	uid, _ := UserIDFromContext(r.Context())
	user, err := s.Store.GetUserByID(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Usuário inválido")
		return
	}
	c, err := s.CRM.Claim(id, uid, user.Name)
	if errors.Is(err, crmdemo.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Conversa não encontrada")
		return
	}
	if errors.Is(err, crmdemo.ErrConflict) {
		writeError(w, http.StatusConflict, "conflict", "Conversa já atribuída")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) UpdatePipeline(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("conversationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	var req struct {
		Stage      domain.PipelineStage `json:"stage"`
		ReasonCode string               `json:"reason_code"`
		ReasonText string               `json:"reason_text"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	c, err := s.CRM.UpdatePipeline(id, req.Stage, req.ReasonCode, req.ReasonText)
	if errors.Is(err, crmdemo.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Conversa não encontrada")
		return
	}
	if errors.Is(err, crmdemo.ErrUnprocessable) {
		writeError(w, http.StatusUnprocessableEntity, "unprocessable", "Motivo obrigatório para Não fechado")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) ListFollowUps(w http.ResponseWriter, r *http.Request) {
	var unitID *uuid.UUID
	if v := r.URL.Query().Get("unit_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "unit_id inválido")
			return
		}
		unitID = &id
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "open"
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": s.CRM.ListFollowUps(unitID, status)})
}

func (s *Server) CompleteFollowUp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("followUpId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	f, err := s.CRM.CompleteFollowUp(id)
	if errors.Is(err, crmdemo.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Follow-up não encontrado")
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (s *Server) ListPops(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.CRM.ListPops(r.URL.Query().Get("intent"))})
}

func (s *Server) CreatePop(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title      string   `json:"title"`
		Body       string   `json:"body"`
		IntentTags []string `json:"intent_tags"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	pop, err := s.CRM.CreatePop(req.Title, req.Body, req.IntentTags)
	if errors.Is(err, crmdemo.ErrBadRequest) {
		writeError(w, http.StatusBadRequest, "bad_request", "Título e texto são obrigatórios")
		return
	}
	writeJSON(w, http.StatusCreated, pop)
}

func (s *Server) UpdatePop(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("popId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	var req struct {
		Title      string   `json:"title"`
		Body       string   `json:"body"`
		IntentTags []string `json:"intent_tags"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	pop, err := s.CRM.UpdatePop(id, req.Title, req.Body, req.IntentTags)
	if errors.Is(err, crmdemo.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "POP não encontrado")
		return
	}
	if errors.Is(err, crmdemo.ErrBadRequest) {
		writeError(w, http.StatusBadRequest, "bad_request", "Título e texto são obrigatórios")
		return
	}
	writeJSON(w, http.StatusOK, pop)
}

func (s *Server) DeletePop(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("popId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "ID inválido")
		return
	}
	if err := s.CRM.DeletePop(id); errors.Is(err, crmdemo.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "POP não encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ListLossReasons(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.CRM.LossReasons()})
}

func (s *Server) DashboardSummary(w http.ResponseWriter, r *http.Request) {
	var unitID *uuid.UUID
	if v := r.URL.Query().Get("unit_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "unit_id inválido")
			return
		}
		unitID = &id
	}
	writeJSON(w, http.StatusOK, s.CRM.Dashboard(unitID))
}

func (s *Server) GetWhatsAppSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.CRM.Settings())
}

func (s *Server) PutWhatsAppSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		domain.WhatsAppSettings
		Token string `json:"token"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "JSON inválido")
		return
	}
	writeJSON(w, http.StatusOK, s.CRM.UpdateSettings(req.WhatsAppSettings, req.Token))
}
