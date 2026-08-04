// Package meta define o contrato de integração com a Meta (WhatsApp Cloud
// API, Instagram e Facebook Messenger Send API, Private Reply API) usado
// pelos usecases de conversation/engagement.
//
// Nenhum usecase instancia um client HTTP diretamente — todos dependem só de
// Sender/CommentResponder, resolvidos em runtime via Registry. Isso permite
// desenvolver e testar o backend inteiro (claim, pipeline, roteamento de
// canal, SSE, webhooks de status) com MockClient, antes de existir qualquer
// credencial Meta real. Os clients HTTP reais (fase 6 do roadmap — ver
// ../../ARCHITECTURE.md) implementam a mesma interface e só precisam ser
// registrados no lugar do mock em internal/app/app.go.
//
// O pacote não decide regras de negócio (ex.: se a janela de 24h já fechou e
// portanto é preciso usar template) — isso é do usecase, que já sabe o
// window_expires_at da conversa. O client só executa o envio que foi pedido.
package meta

import (
	"context"
	"time"
)

// ChannelType identifica o canal Meta de uma mensagem/engagement.
type ChannelType string

const (
	ChannelWhatsApp  ChannelType = "whatsapp"
	ChannelInstagram ChannelType = "instagram"
	ChannelFacebook  ChannelType = "facebook"
)

// Recipient endereça o destinatário do jeito que cada canal exige — nunca o
// Customer.id do CRM, que é interno. WhatsApp usa o wa_id (E.164 sem o "+");
// Instagram e Messenger usam o external_id escopado do canal (IGSID/PSID).
// Ver docs/BACKEND-CONTRACT.md §3 (Customers).
type Recipient struct {
	Channel    ChannelType
	ExternalID string
}

// SendTextInput é um envio de texto livre 1:1 — só válido dentro da janela de
// atendimento (24h no WhatsApp; sem limite formal em IG/FB, mas sujeito às
// políticas da Meta).
type SendTextInput struct {
	Recipient Recipient
	Body      string
}

// SendTemplateInput cobre tanto templates de negócio (envio fora da janela de
// 24h no WhatsApp) quanto o template de autenticação usado no fluxo de OTP de
// posse de número (docs/BACKEND-CONTRACT.md §3, "Verificação WhatsApp").
type SendTemplateInput struct {
	Recipient    Recipient
	TemplateName string
	LanguageCode string   // ex.: "pt_BR"
	Params       []string // parâmetros posicionais do template já aprovado na Meta
}

// ReplyCommentInput referencia o comentário/story/live original — usado tanto
// pra resposta pública quanto pra Private Reply.
type ReplyCommentInput struct {
	Channel           ChannelType
	CommentExternalID string // id nativo do comentário/story na Meta
	Body              string
}

// SendResult é o retorno comum de qualquer envio. MetaMessageID é o que o
// backend persiste em Message.meta_message_id (wamid.* / mid.*) pra
// reconciliar status de entrega/leitura vindos por webhook.
type SendResult struct {
	MetaMessageID string
	SentAt        time.Time
}

// Sender é o contrato de envio direto (1:1), implementado por cada canal.
// O usecase de conversation resolve o Sender certo via Registry a partir do
// channel da conversa — nunca decide qual API da Meta chamar.
type Sender interface {
	Channel() ChannelType

	// SendText envia uma mensagem de texto livre dentro da janela de atendimento.
	SendText(ctx context.Context, input SendTextInput) (SendResult, error)

	// SendTemplate envia uma mensagem de template pré-aprovado — obrigatório no
	// WhatsApp fora da janela de 24h, e é o único jeito de disparar o OTP de
	// autenticação em qualquer canal.
	SendTemplate(ctx context.Context, input SendTemplateInput) (SendResult, error)
}

// CommentResponder é implementado só pelos canais com engagements Meta-nativos
// (Instagram, Facebook) — WhatsApp não tem posts, comments nem stories.
type CommentResponder interface {
	// ReplyPublic posta uma resposta visível publicamente sob o comentário original.
	ReplyPublic(ctx context.Context, input ReplyCommentInput) (SendResult, error)

	// ReplyPrivate envia uma DM via Private Reply API, referenciando o comentário
	// original — vira um Message com reply_to_engagement_id preenchido.
	ReplyPrivate(ctx context.Context, input ReplyCommentInput) (SendResult, error)
}
