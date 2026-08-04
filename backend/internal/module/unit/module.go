package unit

import (
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	"github.com/cia-da-vacina/crm/backend/internal/module/unit/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/unit/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/unit/usecase"
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

	r.Group("/units", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)

		r.Get("/", m.handler.List)
		r.Get("/{id}", m.handler.Get)
		r.Method(http.MethodPost, "/", middleware.RequireRole(admin)(http.HandlerFunc(m.handler.Create)))
		r.Method(http.MethodPatch, "/{id}", middleware.RequireRole(admin)(http.HandlerFunc(m.handler.Update)))
	})
}
