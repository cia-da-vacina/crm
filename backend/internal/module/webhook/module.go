// Package webhook recebe eventos da Meta (WhatsApp/Instagram/Facebook) —
// registrado fora de /api/v1 (docs/BACKEND-CONTRACT.md §8: "nenhum destes
// endpoints é chamado pelo frontend, direta ou indiretamente via proxy").
// Sem RequireAuth: a autenticação aqui é a assinatura HMAC, não um Bearer.
package webhook

import (
	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement"
	"github.com/cia-da-vacina/crm/backend/internal/module/pricing"
	"github.com/cia-da-vacina/crm/backend/internal/module/triage"
	"github.com/cia-da-vacina/crm/backend/internal/module/webhook/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/webhook/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/webhook/usecase"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type Module struct {
	handler *handler.Handler
}

func New(a *app.App) *Module {
	repo := repository.New(a.DB)
	// Instâncias próprias dos usecases de triagem e engagement (não as dos
	// módulos triage/engagement) — stateless, duplicar a montagem é
	// inofensivo e mantém os módulos desacoplados (mesma convenção de
	// CustomerReader em conversation/module.go).
	triageUC := triage.NewUseCase(a)
	engagementUC := engagement.NewUseCase(a)
	pricingUC := pricing.NewUseCase(a)
	uc := usecase.New(repo, triageUC, engagementUC, a.SSE, pricingUC)
	h := handler.New(uc)
	return &Module{handler: h}
}

func (m *Module) Register(r *httppkg.Router) {
	r.Group("/webhooks/meta", func(r *httppkg.Router) {
		r.Get("/{channel}", m.handler.Verify)
		r.Post("/{channel}", m.handler.Receive)
	})
}
