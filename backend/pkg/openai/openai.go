// Package openai abstrai o provedor de LLM usado pela triagem (fase 7) —
// interface genérica o bastante pra não vazar detalhe de API específico da
// OpenAI em quem consome (docs/decisions.md D-02: "gpt-4o-mini ou
// equivalente barato", atrás de feature flag — a troca de provedor um dia
// não deveria exigir reescrever o usecase de triagem).
package openai

import "context"

type Message struct {
	Role    string // "system" | "user" | "assistant"
	Content string
}

type CompletionRequest struct {
	Model string
	// Messages já vem pronta (system prompt + contexto + histórico) — quem
	// monta o prompt é o usecase de triagem, não este pacote.
	Messages []Message
	// JSONResponse pede response_format: {"type":"json_object"} — a triagem
	// sempre quer JSON estruturado de volta, nunca prosa livre.
	JSONResponse bool
	Temperature  float64
}

type CompletionResult struct {
	Content string
}

// Client é implementado por HTTPClient (chamada real) e MockClient (loga e
// devolve uma resposta configurável, sem rede) — mesma dualidade de
// pkg/meta.Sender/MockClient.
type Client interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResult, error)
}
