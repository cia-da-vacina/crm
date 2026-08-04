package conversation

import (
	"github.com/cia-da-vacina/crm/backend/internal/app"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/handler"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/usecase"
	customerrepository "github.com/cia-da-vacina/crm/backend/internal/module/customer/repository"
	customerusecase "github.com/cia-da-vacina/crm/backend/internal/module/customer/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	"github.com/cia-da-vacina/crm/backend/internal/module/pricing"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
)

type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func New(a *app.App) *Module {
	repo := repository.New(a.DB)

	// CustomerReader próprio (não o módulo customer inteiro) — só pra montar
	// ConversationDetail.Customer. Instância independente da usada em
	// internal/module/customer/module.go; ambas são stateless (só um
	// *database.DB por baixo), então duplicar a montagem é inofensivo e
	// mantém os módulos desacoplados.
	customerReader := customerusecase.New(customerrepository.New(a.DB))

	// PricingReader próprio (não o módulo pricing inteiro) — só pra estimar
	// Message.CostBRL no envio (Frente A do plano de adaptação WhatsApp
	// 2026). Mesma convenção de instância independente de customerReader
	// acima.
	pricingReader := pricing.NewUseCase(a)

	uc := usecase.New(repo, customerReader, a.SSE, a.Meta, a.Audit, pricingReader)
	h := handler.New(uc, a.SSE)
	return &Module{handler: h, mw: middleware.New(a.JWT)}
}

func (m *Module) Register(r *httppkg.Router) {
	r.Group("/inbox", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)

		r.Get("/", m.handler.ListInbox)
		r.Get("/stream", m.handler.Stream)
	})

	r.Group("/conversations", func(r *httppkg.Router) {
		r.Use(m.mw.RequireAuth)

		r.Get("/{id}", m.handler.Get)
		r.Get("/{id}/messages", m.handler.ListMessages)
		r.Post("/{id}/messages", m.handler.SendMessage)
		r.Post("/{id}/claim", m.handler.Claim)
		r.Patch("/{id}/pipeline", m.handler.UpdatePipeline)
		r.Post("/{id}/phone", m.handler.InitiatePhone)
		r.Post("/{id}/phone/resend", m.handler.ResendPhone)
		r.Post("/{id}/phone/confirm", m.handler.ConfirmPhone)
	})
}
