// Package template expõe o CRUD do catálogo de templates WhatsApp
// (docs/WHATSAPP-2026-ADAPTATION-PLAN.md, Frente B — decisão D5: cadastro
// próprio no CRM, backend é fonte de verdade de categoria/variáveis/status
// de aprovação). Mesmo padrão de RBAC do módulo pop: leitura pra qualquer
// autenticado, escrita restrita a admin/manager.
package template

import (
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	"github.com/cia-da-vacina/crm/backend/internal/module/template/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/template/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/template/usecase"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func New(a *app.App) *Module {
	repo := repository.New(a.DB)
	uc := usecase.New(repo)
	h := handler.New(uc)
	return &Module{handler: h, mw: middleware.New(a.JWT)}
}

func (m *Module) Register(r *httppkg.Router) {
	admin := string(entity.RoleAdmin)
	manager := string(entity.RoleManager)

	r.Group("/templates", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)

		r.Get("/", m.handler.List)
		r.Get("/{id}", m.handler.Get)
		r.Method(http.MethodPost, "/", middleware.RequireRole(admin, manager)(http.HandlerFunc(m.handler.Create)))
		r.Method(http.MethodPatch, "/{id}", middleware.RequireRole(admin, manager)(http.HandlerFunc(m.handler.Update)))
		r.Method(http.MethodDelete, "/{id}", middleware.RequireRole(admin, manager)(http.HandlerFunc(m.handler.Delete)))
	})
}
