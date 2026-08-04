package auth

import (
	"net/http"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/module/auth/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/auth/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/auth/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/go-chi/httprate"
)

type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func New(a *app.App) *Module {
	repo := repository.New(a.DB)
	uc := usecase.New(repo, a.JWT)
	h := handler.New(uc)
	return &Module{handler: h, mw: middleware.New(a.JWT)}
}

func (m *Module) Register(r *httppkg.Router) {
	// login/refresh são as únicas rotas sem Bearer — rate limit por IP contra
	// brute-force (docs/BACKEND-CONTRACT.md §9). Aplicado por rota (não via
	// Group) porque /auth/logout precisa de um middleware diferente
	// (RequireAuth) no mesmo prefixo, e o Router.Group monta um sub-mux novo
	// a cada chamada — duas chamadas no mesmo prefixo colidiriam.
	loginRateLimit := httprate.LimitByIP(10, 1*time.Minute)

	r.Method(http.MethodPost, "/auth/login", loginRateLimit(http.HandlerFunc(m.handler.Login)))
	r.Method(http.MethodPost, "/auth/refresh", loginRateLimit(http.HandlerFunc(m.handler.Refresh)))
	r.Method(http.MethodPost, "/auth/logout", m.mw.RequireAuth(http.HandlerFunc(m.handler.Logout)))
}
