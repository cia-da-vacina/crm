package handler

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/env"
	"github.com/cia-da-vacina/crm/backend/pkg/meta"
	"github.com/go-chi/chi/v5"
)

type UseCase interface {
	IngestPayload(ctx context.Context, channel entity.Channel, rawBody []byte) error
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

var validChannels = map[string]entity.Channel{
	"whatsapp":  entity.ChannelWhatsApp,
	"instagram": entity.ChannelInstagram,
	"facebook":  entity.ChannelFacebook,
}

// Verify é o handshake GET que a Meta faz uma vez, na hora de configurar o
// webhook no App Dashboard — responde hub.challenge em texto puro se
// hub.verify_token bater com META_WEBHOOK_VERIFY_TOKEN (docs/BACKEND-CONTRACT.md §8).
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("hub.mode") != "subscribe" || q.Get("hub.verify_token") != env.GetOrDefault("META_WEBHOOK_VERIFY_TOKEN", "") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(q.Get("hub.challenge")))
}

// Receive valida a assinatura antes de processar qualquer coisa — payload
// não assinado corretamente nunca é persistido, sem exceção mesmo em dev
// (docs/BACKEND-CONTRACT.md §8/§9).
func (h *Handler) Receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !meta.VerifySignature(env.GetOrDefault("META_APP_SECRET", ""), body, r.Header.Get("X-Hub-Signature-256")) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	channel, ok := validChannels[chi.URLParam(r, "channel")]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Ingestão roda síncrona antes do 200: são poucas queries (resolver
	// identidade + conversa + inserir mensagem), rápido o bastante pra não
	// esbarrar no timeout da Meta — não precisou de fila assíncrona real
	// (ver backend/ARCHITECTURE.md §2, nota sobre pkg/jobs não ter sido
	// necessário até agora). Erro de parse (payload malformado) é só
	// logado: reprocessar não ajudaria, a Meta reenviaria pra sempre.
	if err := h.uc.IngestPayload(r.Context(), channel, body); err != nil {
		log.Printf("webhook: failed to ingest payload: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}
