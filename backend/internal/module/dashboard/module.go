package dashboard

import (
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/dashboard/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/dashboard/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/dashboard/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
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

	r.Group("/dashboard", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)
		r.Get("/summary", m.handler.Get)
		// Custo é dado financeiro — restrito a admin/manager, diferente do
		// resumo operacional acima (aberto a qualquer papel autenticado).
		r.Method(http.MethodGet, "/costs", middleware.RequireRole(admin, manager)(http.HandlerFunc(m.handler.GetCosts)))
	})
}
