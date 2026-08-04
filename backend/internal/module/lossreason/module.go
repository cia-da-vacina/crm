package lossreason

import (
	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/module/lossreason/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/lossreason/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/lossreason/usecase"
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
	r.Group("/loss-reasons", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)
		r.Get("/", m.handler.List)
	})
}
