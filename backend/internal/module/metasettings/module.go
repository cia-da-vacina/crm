package metasettings

import (
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/metasettings/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/metasettings/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/metasettings/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func New(a *app.App) *Module {
	repo := repository.New(a.DB)
	uc := usecase.New(repo, a.Crypto, a.Audit)
	h := handler.New(uc)
	return &Module{handler: h, mw: middleware.New(a.JWT)}
}

func (m *Module) Register(r *httppkg.Router) {
	admin := string(entity.RoleAdmin)

	r.Group("/settings/meta", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)
		r.Use(middleware.RequireRole(admin))

		r.Get("/", m.handler.Get)
		r.Method(http.MethodPut, "/", http.HandlerFunc(m.handler.Update))
	})
}
