// Package auditlog expõe GET /audit-logs, admin-only — a leitura do que
// pkg/audit grava. Não previsto literalmente no openapi.yaml (o contrato só
// exige que a escrita seja append-only, docs/BACKEND-CONTRACT.md §9), mas um
// log que ninguém consegue ler pela API não cumpre o propósito de auditoria
// — extensão de baixo risco, documentada em backend/ARCHITECTURE.md §5.
package auditlog

import (
	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog/usecase"
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
	r.Group("/audit-logs", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)
		r.Use(middleware.RequireRole(string(entity.RoleAdmin)))

		r.Get("/", m.handler.List)
	})
}
