// Package sse implementa Server-Sent Events: um Hub de broadcast em memória
// (sem Redis/Kafka, conforme ADR 0001) e um helper HTTP que escreve o
// protocolo text/event-stream com heartbeat — necessário porque proxies
// reversos derrubam conexões longas sem tráfego (ver backend/ARCHITECTURE.md,
// nota do ADR sobre SSE exigir cuidado com timeouts).
package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Event é publicado no Hub e entregue a toda conexão SSE ativa. UnitID
// define o escopo de autorização — Serve só escreve o evento no wire se
// `allowed(UnitID)` for true pra aquela conexão (ver módulo conversation,
// que é quem decide quais unidades cada usuário pode ver).
type Event struct {
	Name   string
	UnitID string
	Data   any
}

// Hub distribui eventos pra todas as conexões SSE abertas no processo. A
// filtragem por unidade acontece em Serve, não aqui — o Hub não sabe nada
// sobre RBAC.
type Hub struct {
	mu   sync.RWMutex
	subs map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[chan Event]struct{})}
}

func (h *Hub) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 16)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		if _, ok := h.subs[ch]; ok {
			delete(h.subs, ch)
			close(ch)
		}
		h.mu.Unlock()
	}

	return ch, cancel
}

// Publish nunca bloqueia: um subscriber lento/travado simplesmente perde o
// evento (canal com buffer, select non-blocking) em vez de travar quem
// publicou — SSE aqui é "best effort real-time", não uma fila confiável.
func (h *Hub) Publish(e Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subs {
		select {
		case ch <- e:
		default:
		}
	}
}

// Serve escreve o protocolo text/event-stream até o contexto da request
// terminar (client desconectou) ou o canal fechar. allowed decide, por
// evento, se aquela conexão pode ver aquela unidade.
func Serve(w http.ResponseWriter, r *http.Request, ch <-chan Event, allowed func(unitID string) bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()

		case event, ok := <-ch:
			if !ok {
				return
			}
			if event.UnitID != "" && !allowed(event.UnitID) {
				continue
			}

			payload, err := json.Marshal(event.Data)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Name, payload); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
