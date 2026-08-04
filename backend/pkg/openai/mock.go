package openai

import (
	"context"
	"log"
	"sync"
)

// defaultMockResponse é um JSON plausível de triagem — usado quando ninguém
// configurou Response explicitamente. Mantém o formato que
// triage/usecase espera parsear, pra o fluxo inteiro funcionar em dev sem
// chave real.
const defaultMockResponse = `{
	"intent": "duvidas",
	"confidence": 0.6,
	"summary": "Cliente entrou em contato, ainda sem intenção clara identificada.",
	"internal_notes": "",
	"phone_gate_required": false,
	"ready_for_handoff": false,
	"reply": "Oi! Sou a assistente virtual da Cia da Vacina. Como posso te ajudar hoje?",
	"collected_fields": {}
}`

// MockClient não chama rede nenhuma — devolve Response (configurável,
// thread-safe) e loga a chamada, pro resto do backend (triage/usecase,
// integração com pkg/meta pro envio da resposta) poder ser desenvolvido e
// testado de ponta a ponta antes de existir OPENAI_API_KEY real.
type MockClient struct {
	mu       sync.RWMutex
	response string
}

func NewMockClient() *MockClient {
	return &MockClient{response: defaultMockResponse}
}

// SetResponse troca o JSON devolvido pela próxima chamada — usado em testes
// pra exercitar caminhos diferentes (ex.: intent "agendar" com
// phone_gate_required=true).
func (m *MockClient) SetResponse(json string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.response = json
}

func (m *MockClient) Complete(ctx context.Context, req CompletionRequest) (CompletionResult, error) {
	m.mu.RLock()
	resp := m.response
	m.mu.RUnlock()

	log.Printf("[openai:mock] model=%s messages=%d json_response=%v -> canned response", req.Model, len(req.Messages), req.JSONResponse)
	return CompletionResult{Content: resp}, nil
}
