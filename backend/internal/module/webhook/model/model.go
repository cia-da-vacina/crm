package model

import (
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
)

// InboundMessage é a forma unificada que os parsers de cada canal produzem
// — o resto da ingestão não sabe mais nada sobre o shape original da Meta
// depois disso.
type InboundMessage struct {
	Channel       entity.Channel
	ExternalID    string // wa_id | IGSID | PSID
	DisplayHandle string // nome de perfil, quando o payload traz (raro fora do WhatsApp)
	MetaMessageID string
	Body          string
	Timestamp     time.Time
	PhoneNumberID string // só preenchido pro WhatsApp — usado pra achar a unidade dona
}

// InboundStatus é a forma unificada de um evento do array "statuses" do
// webhook de status da WhatsApp Cloud API — só existe pro WhatsApp (Frente A
// do plano de adaptação WhatsApp 2026: a mudança de cobrança de out/2026 é
// específica desse canal). Category/Billable/PricingModel vêm nil quando o
// evento de status não tem objeto "pricing" (ex.: um "sent"/"delivered" que
// ainda não chegou em "read", ou quando a mensagem não é billable).
type InboundStatus struct {
	MetaMessageID string
	Status        string // sent | delivered | read | failed
	Category      *string
	Billable      *bool
	PricingModel  *string
	Timestamp     time.Time
}

// InboundEngagement é a forma unificada que os parsers de comentário/story
// produzem, análoga a InboundMessage — só existe pra Instagram/Facebook
// (WhatsApp não tem posts, stories nem comments). UnitID é preenchido depois
// do parse, pela mesma resolução de unidade centralizada usada pra mensagens
// (docs/BACKEND-CONTRACT.md §5).
type InboundEngagement struct {
	Channel          entity.Channel
	Type             entity.EngagementType
	ExternalID       string // id nativo do comentário/story na Meta
	AuthorExternalID string // IGSID/PSID de quem comentou/respondeu
	Body             string
	MediaID          *string
	MediaURL         *string
	MediaCaption     *string
	Timestamp        time.Time
}
