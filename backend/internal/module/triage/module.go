package triage

import (
	"log"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	"github.com/cia-da-vacina/crm/backend/internal/module/triage/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/triage/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/triage/usecase"
	"github.com/cia-da-vacina/crm/backend/pkg/env"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/cia-da-vacina/crm/backend/pkg/openai"
)

type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func New(a *app.App) *Module {
	uc := NewUseCase(a)
	h := handler.New(uc)
	return &Module{handler: h, mw: middleware.New(a.JWT)}
}

// NewUseCase é exportado porque o módulo webhook também precisa disparar
// RunTriage depois de ingerir uma mensagem — instância própria, independente
// desta (mesma convenção de módulo autocontido de todo o projeto: stateless,
// duplicar a montagem é inofensivo e mantém os módulos desacoplados — ver
// backend/ARCHITECTURE.md).
func NewUseCase(a *app.App) *usecase.UseCase {
	repo := repository.New(a.DB)
	model := env.GetOrDefault("OPENAI_MODEL", "gpt-4o-mini")
	return usecase.New(repo, newOpenAIClient(), a.Meta, a.SSE, model)
}

// newOpenAIClient usa o client real se OPENAI_API_KEY estiver configurada;
// senão cai pro mock (mesma decisão de app.go pro pkg/meta) — liga sozinho
// quando alguém configurar a chave, sem mudar código.
func newOpenAIClient() openai.Client {
	apiKey := env.GetOrDefault("OPENAI_API_KEY", "")
	if apiKey == "" {
		log.Println("triage: OPENAI_API_KEY not set — using mock client (no real OpenAI calls)")
		return openai.NewMockClient()
	}
	log.Println("triage: OPENAI_API_KEY configured — using real OpenAI client")
	return openai.NewHTTPClient(apiKey)
}

func (m *Module) Register(r *httppkg.Router) {
	r.Method(http.MethodGet, "/conversations/{id}/triage", m.mw.RequireAuth(http.HandlerFunc(m.handler.Get)))
}
