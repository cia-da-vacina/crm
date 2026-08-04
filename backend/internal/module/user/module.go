package user

import (
	"net/http"

	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	"github.com/cia-da-vacina/crm/backend/internal/module/user/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/user/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/user/usecase"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func New(a *app.App) *Module {
	repo := repository.New(a.DB)
	uc := usecase.New(repo, a.Audit)
	h := handler.New(uc)
	return &Module{handler: h, mw: middleware.New(a.JWT)}
}

func (m *Module) Register(r *httppkg.Router) {
	admin := string(entity.RoleAdmin)
	manager := string(entity.RoleManager)

	r.Group("/users", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)

		// GET /users/:id fica aberto a qualquer usuário autenticado (ex.: exibir
		// nome de quem assumiu uma conversa) — só a listagem e o CRUD são restritos.
		r.Get("/{id}", m.handler.Get)

		r.Method(http.MethodGet, "/", middleware.RequireRole(admin, manager)(http.HandlerFunc(m.handler.List)))
		r.Method(http.MethodPost, "/", middleware.RequireRole(admin)(http.HandlerFunc(m.handler.Create)))
		r.Patch("/{id}", m.handler.Update) // self vs admin decidido no usecase
		r.Method(http.MethodDelete, "/{id}", middleware.RequireRole(admin)(http.HandlerFunc(m.handler.Delete)))
		r.Method(http.MethodPut, "/{id}/units", middleware.RequireRole(admin)(http.HandlerFunc(m.handler.SetUnits)))
	})
}
