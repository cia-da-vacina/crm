package engagement

import (
	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func New(a *app.App) *Module {
	repo := repository.New(a.DB)
	uc := usecase.New(repo, a.Meta, a.SSE)
	h := handler.New(uc)
	return &Module{handler: h, mw: middleware.New(a.JWT)}
}

// NewUseCase expõe o usecase pra outros módulos construírem sua própria
// instância independente (ex.: webhook, na ingestão de comentário/story) —
// mesma convenção de triage.NewUseCase (ver backend/ARCHITECTURE.md).
func NewUseCase(a *app.App) *usecase.UseCase {
	return usecase.New(repository.New(a.DB), a.Meta, a.SSE)
}

func (m *Module) Register(r *httppkg.Router) {
	r.Group("/engagements", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)

		r.Get("/", m.handler.List)
		r.Get("/{id}", m.handler.Get)
		r.Post("/{id}/reply", m.handler.Reply)
		r.Post("/{id}/dismiss", m.handler.Dismiss)
		r.Post("/{id}/convert", m.handler.Convert)
	})
}
