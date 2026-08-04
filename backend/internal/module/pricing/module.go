// Package pricing expõe o rate card local usado pro núcleo de custo
// (docs/WHATSAPP-2026-ADAPTATION-PLAN.md, Frente A) — GET/PATCH
// /settings/pricing-rates, admin-only, mesmo padrão de RBAC de
// /settings/meta. NewUseCase (não só New) é exportado pra outros módulos
// (conversation, webhook) montarem sua própria instância do usecase como
// PricingReader, sem depender do Module inteiro — mesma convenção de
// customerusecase.New em conversation/module.go.
package pricing

import (
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	"github.com/cia-da-vacina/crm/backend/internal/module/pricing/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/pricing/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/pricing/usecase"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

// NewUseCase monta um *usecase.UseCase isolado a partir de *app.App — usado
// por outros módulos que só precisam de GetRate (PricingReader), não da rota
// HTTP.
func NewUseCase(a *app.App) *usecase.UseCase {
	return usecase.New(repository.New(a.DB))
}

func New(a *app.App) *Module {
	uc := NewUseCase(a)
	h := handler.New(uc)
	return &Module{handler: h, mw: middleware.New(a.JWT)}
}

func (m *Module) Register(r *httppkg.Router) {
	admin := string(entity.RoleAdmin)

	r.Group("/settings/pricing-rates", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)
		r.Use(middleware.RequireRole(admin))

		r.Get("/", m.handler.List)
		r.Method(http.MethodPatch, "/{category}", http.HandlerFunc(m.handler.Update))
	})
}
