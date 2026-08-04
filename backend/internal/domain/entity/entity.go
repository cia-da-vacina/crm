// Package entity contém as structs planas 1:1 com as tabelas do banco (tags
// `db` para sqlx). Cada module tem sua própria camada de model/DTO para o
// shape exposto pela API — entity nunca é serializada diretamente numa
// resposta HTTP.
package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleManager    UserRole = "manager"
	RoleSupervisor UserRole = "supervisor"
	RoleAgent      UserRole = "agent"
)

type User struct {
	ID           string    `db:"id"            json:"id"`
	Email        string    `db:"email"         json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Name         string    `db:"name"          json:"name"`
	Role         UserRole  `db:"role"          json:"role"`
	Active       bool      `db:"active"        json:"active"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}

type Unit struct {
	ID         string    `db:"id"         json:"id"`
	Name       string    `db:"name"       json:"name"`
	Code       string    `db:"code"       json:"code"`
	Timezone   string    `db:"timezone"   json:"timezone"`
	Active     bool      `db:"active"     json:"active"`
	Address    string    `db:"address"    json:"address"`
	City       string    `db:"city"       json:"city"`
	District   *string   `db:"district"   json:"district,omitempty"`
	Complement *string   `db:"complement" json:"complement,omitempty"`
	Reference  *string   `db:"reference"  json:"reference,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// RefreshToken nunca guarda o token bruto — só o hash (SHA-256). O valor
// bruto só existe em memória entre a geração e a entrega ao client.
type RefreshToken struct {
	ID        string     `db:"id"         json:"id"`
	UserID    string     `db:"user_id"    json:"user_id"`
	TokenHash string     `db:"token_hash" json:"-"`
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
	IP        string     `db:"ip"         json:"ip"`
	UserAgent string     `db:"user_agent" json:"user_agent"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

// Channel identifica o canal Meta de origem de uma identidade/mensagem.
// Duplica pkg/meta.ChannelType de propósito: entity é domínio puro e não
// deve depender de um pacote de integração (pkg/meta fica livre pra ser
// standalone); a conversão é um cast de string na borda onde os dois se
// encontram (a partir da fase 6 do roadmap — ver backend/ARCHITECTURE.md).
type Channel string

const (
	ChannelWhatsApp  Channel = "whatsapp"
	ChannelInstagram Channel = "instagram"
	ChannelFacebook  Channel = "facebook"
)

type CustomerIdentification string

const (
	IdentificationAnonymous  CustomerIdentification = "anonymous"
	IdentificationIdentified CustomerIdentification = "identified"
)

// Customer é a identidade canônica do cliente no CRM. primary_phone (E.164)
// é a chave de negócio pra unificar canais — só existe depois que a "parede
// de privacidade" é atravessada (ver docs/BACKEND-CONTRACT.md §3). Até lá o
// cliente é anonymous e só tem identidades de canal (CustomerIdentity), sem
// telefone nem histórico cross-canal.
type Customer struct {
	ID             string                 `db:"id"              json:"id"`
	DisplayName    string                 `db:"display_name"    json:"display_name"`
	Identification CustomerIdentification `db:"identification"  json:"identification"`
	PrimaryPhone   *string                `db:"primary_phone"   json:"primary_phone"`
	UnitID         *string                `db:"unit_id"         json:"unit_id,omitempty"`
	CreatedAt      time.Time              `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time              `db:"updated_at"      json:"updated_at"`
}

// CustomerIdentity vincula N:1 o identificador nativo de um canal (wa_id,
// IGSID ou PSID) a um Customer. VerifiedAt é preenchido na criação para
// WhatsApp (a própria entrega do número pela Meta já prova posse) e só após
// confirmação de OTP para os demais canais.
type CustomerIdentity struct {
	ID            string     `db:"id"             json:"id"`
	CustomerID    string     `db:"customer_id"    json:"customer_id"`
	Channel       Channel    `db:"channel"        json:"channel"`
	ExternalID    string     `db:"external_id"    json:"external_id"`
	DisplayHandle *string    `db:"display_handle" json:"display_handle,omitempty"`
	PhoneE164     *string    `db:"phone_e164"     json:"phone_e164,omitempty"`
	VerifiedAt    *time.Time `db:"verified_at"    json:"verified_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at"     json:"created_at"`
}

type PipelineStage string

const (
	StageEmAtendimento        PipelineStage = "em_atendimento"
	StageEmNegociacao         PipelineStage = "em_negociacao"
	StageAguardandoFechamento PipelineStage = "aguardando_fechamento"
	StageFechado              PipelineStage = "fechado"
	StageNaoFechado           PipelineStage = "nao_fechado"
)

// ValidPipelineStages não define ordem — qualquer transição entre estágios
// válidos é permitida (decisão de produto: backend só valida pertinência ao
// enum, não sequência; ver backend/ARCHITECTURE.md).
var ValidPipelineStages = map[PipelineStage]bool{
	StageEmAtendimento:        true,
	StageEmNegociacao:         true,
	StageAguardandoFechamento: true,
	StageFechado:              true,
	StageNaoFechado:           true,
}

type ConversationMode string

const (
	ModeAITriage ConversationMode = "ai_triage"
	ModeHuman    ConversationMode = "human"
)

type PhoneGate string

const (
	PhoneGateNotNeeded           PhoneGate = "not_needed"
	PhoneGateRequired            PhoneGate = "required"
	PhoneGatePendingVerification PhoneGate = "pending_verification"
	PhoneGateCollected           PhoneGate = "collected"
)

// Conversation é uma thread de UM canal com UMA identidade de cliente — um
// Customer pode ter várias conversas abertas simultaneamente em canais
// diferentes (docs/BACKEND-CONTRACT.md §4).
type Conversation struct {
	ID                 string           `db:"id"                   json:"id"`
	CustomerID         string           `db:"customer_id"          json:"customer_id"`
	Channel            Channel          `db:"channel"              json:"channel"`
	ChannelThreadID    *string          `db:"channel_thread_id"    json:"channel_thread_id,omitempty"`
	UnitID             string           `db:"unit_id"              json:"unit_id"`
	PipelineStage      PipelineStage    `db:"pipeline_stage"       json:"pipeline_stage"`
	Mode               ConversationMode `db:"mode"                 json:"mode"`
	OwnerID            *string          `db:"owner_id"             json:"owner_id,omitempty"`
	Intent             *string          `db:"intent"               json:"intent,omitempty"`
	AISummary          *string          `db:"ai_summary"           json:"ai_summary,omitempty"`
	TriageNotes        *string          `db:"triage_notes"         json:"triage_notes,omitempty"`
	PhoneGate          PhoneGate        `db:"phone_gate"           json:"phone_gate"`
	LossReasonCode     *string          `db:"loss_reason_code"     json:"-"`
	LossReasonText     *string          `db:"loss_reason_text"     json:"-"`
	WindowExpiresAt    *time.Time       `db:"window_expires_at"    json:"window_expires_at,omitempty"`
	LastMessagePreview string           `db:"last_message_preview" json:"last_message_preview"`
	LastMessageAt      *time.Time       `db:"last_message_at"      json:"last_message_at,omitempty"`
	// CollectedFields são os campos estruturados que a IA de triagem extrai
	// da conversa (fase 7) — ex.: vacina desejada, unidade preferida.
	CollectedFields       JSONObject `db:"collected_fields"           json:"collected_fields,omitempty"`
	TriageConfidence      *float64   `db:"triage_confidence"          json:"-"`
	TriageReadyForHandoff bool       `db:"triage_ready_for_handoff"   json:"-"`
	CreatedAt             time.Time  `db:"created_at"                 json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at"        json:"updated_at"`
}

// JSONObject é um mapa JSONB genérico — mesmo padrão de Scan/Value de
// IntentTags, mas pra objeto em vez de array (usado por Conversation.
// CollectedFields, cujo conjunto de chaves varia por intenção e não é
// fechado no MVP).
type JSONObject map[string]any

func (o JSONObject) Value() (driver.Value, error) {
	if o == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(o)
}

func (o *JSONObject) Scan(src any) error {
	if src == nil {
		*o = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("JSONObject: unsupported scan type %T", src)
	}
	if len(data) == 0 {
		*o = nil
		return nil
	}
	return json.Unmarshal(data, o)
}

type MessageDirection string

const (
	DirectionIn  MessageDirection = "in"
	DirectionOut MessageDirection = "out"
)

type SenderType string

const (
	SenderContact SenderType = "contact"
	SenderAgent   SenderType = "agent"
	SenderAI      SenderType = "ai"
	SenderSystem  SenderType = "system"
)

// MessageKind lista os tipos previstos pelo contrato; só "text" é aceito em
// POST /conversations/:id/messages no MVP (texto-first — mídia é V1.1,
// ver docs/decisions.md). Os demais valores existem no enum pra não exigir
// migration quando mídia entrar.
type MessageKind string

const (
	KindText     MessageKind = "text"
	KindImage    MessageKind = "image"
	KindDocument MessageKind = "document"
	KindAudio    MessageKind = "audio"
	KindVideo    MessageKind = "video"
	KindTemplate MessageKind = "template"
	KindSystem   MessageKind = "system"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

type Message struct {
	ID             string           `db:"id"                json:"id"`
	ConversationID string           `db:"conversation_id"   json:"conversation_id"`
	Direction      MessageDirection `db:"direction"         json:"direction"`
	SenderType     SenderType       `db:"sender_type"       json:"sender_type"`
	SenderUserID   *string          `db:"sender_user_id"    json:"sender_user_id,omitempty"`
	Kind           MessageKind      `db:"kind"              json:"kind"`
	Channel        Channel          `db:"channel"           json:"channel"`
	Body           string           `db:"body"              json:"body"`
	Status         MessageStatus    `db:"status"            json:"status"`
	MetaMessageID  *string          `db:"meta_message_id"   json:"meta_message_id,omitempty"`
	MediaURL       *string          `db:"media_url"         json:"media_url,omitempty"`
	MediaMimeType  *string          `db:"media_mime_type"   json:"media_mime_type,omitempty"`
	TemplateName   *string          `db:"template_name"     json:"template_name,omitempty"`
	// ReplyToEngagementID liga a mensagem a um SocialEngagement de origem —
	// ex.: resposta enviada a partir de uma resposta de story (fase 8).
	ReplyToEngagementID *string   `db:"reply_to_engagement_id" json:"reply_to_engagement_id,omitempty"`
	CreatedAt           time.Time `db:"created_at"             json:"created_at"`
}

// PhoneVerification é a pendência de OTP de um número informado em IG/FB
// (docs/BACKEND-CONTRACT.md §3). CodeHash nunca guarda o código em texto
// plano — mesmo padrão de RefreshToken.TokenHash.
type PhoneVerification struct {
	ID             string     `db:"id"              json:"id"`
	ConversationID string     `db:"conversation_id" json:"conversation_id"`
	PhoneE164      string     `db:"phone_e164"      json:"phone_e164"`
	CodeHash       string     `db:"code_hash"       json:"-"`
	Attempts       int        `db:"attempts"        json:"attempts"`
	ResendCount    int        `db:"resend_count"    json:"resend_count"`
	ExpiresAt      time.Time  `db:"expires_at"      json:"expires_at"`
	ConfirmedAt    *time.Time `db:"confirmed_at"    json:"confirmed_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
}

// LossReason é o catálogo de motivos de não conversão — seed fixo em
// docs/decisions.md. Existe já na fase 4 porque PATCH .../pipeline precisa
// validar reason_code contra ele; o CRUD/GET /loss-reasons entra na fase 5.
type LossReason struct {
	Code  string `db:"code"  json:"code"`
	Label string `db:"label" json:"label"`
}

type FollowUpStatus string

const (
	FollowUpOpen     FollowUpStatus = "open"
	FollowUpDone     FollowUpStatus = "done"
	FollowUpCanceled FollowUpStatus = "canceled"
)

// FollowUp é gerado automaticamente pelo backend quando uma conversa entra
// em aguardando_fechamento/nao_fechado (docs/BACKEND-CONTRACT.md §6) — não
// há criação manual. PipelineStage é um snapshot do estágio que gerou o
// follow-up, não o estágio atual da conversa (que pode já ter mudado).
type FollowUp struct {
	ID             string         `db:"id"              json:"id"`
	ConversationID string         `db:"conversation_id" json:"conversation_id"`
	CustomerID     string         `db:"customer_id"     json:"customer_id"`
	UnitID         string         `db:"unit_id"         json:"unit_id"`
	PipelineStage  PipelineStage  `db:"pipeline_stage"  json:"pipeline_stage"`
	DueAt          time.Time      `db:"due_at"          json:"due_at"`
	Status         FollowUpStatus `db:"status"           json:"status"`
	Note           string         `db:"note"            json:"note"`
	CreatedAt      time.Time      `db:"created_at"      json:"created_at"`
	CompletedAt    *time.Time     `db:"completed_at"    json:"completed_at,omitempty"`
}

// IntentTags é armazenado como JSONB (não TEXT[]) — evita depender de scan
// de array do Postgres via pgx/sqlx, que exigiria um tipo Scanner dedicado
// de qualquer forma; aqui reusamos json.Marshal/Unmarshal direto.
type IntentTags []string

func (t IntentTags) Value() (driver.Value, error) {
	if t == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(t)
}

func (t *IntentTags) Scan(src any) error {
	if src == nil {
		*t = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("IntentTags: unsupported scan type %T", src)
	}
	if len(data) == 0 {
		*t = nil
		return nil
	}
	return json.Unmarshal(data, t)
}

type Pop struct {
	ID         string     `db:"id"          json:"id"`
	Title      string     `db:"title"       json:"title"`
	Body       string     `db:"body"        json:"body"`
	IntentTags IntentTags `db:"intent_tags" json:"intent_tags"`
	Active     bool       `db:"active"      json:"active"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
}

// MetaChannelConfig: WhatsApp é 1 linha por unidade (UnitID sempre presente);
// Instagram/Facebook são centralizados — 1 linha global por canal, UnitID
// nil (docs.: confirmado com o stakeholder, ver backend/ARCHITECTURE.md §5).
// TokenCiphertext nunca é exposto via API — só TokenMasked, calculado uma
// vez no momento da rotação.
type MetaChannelConfig struct {
	ID              string    `db:"id"                json:"id"`
	Channel         Channel   `db:"channel"           json:"channel"`
	UnitID          *string   `db:"unit_id"           json:"unit_id,omitempty"`
	Enabled         bool      `db:"enabled"           json:"enabled"`
	AccountID       string    `db:"account_id"        json:"account_id"`
	DisplayName     string    `db:"display_name"      json:"display_name"`
	PhoneNumberID   *string   `db:"phone_number_id"   json:"phone_number_id,omitempty"`
	WebhookVerified bool      `db:"webhook_verified"  json:"webhook_verified"`
	TokenCiphertext []byte    `db:"token_ciphertext"  json:"-"`
	TokenMasked     *string   `db:"token_masked"      json:"token_masked,omitempty"`
	CreatedAt       time.Time `db:"created_at"        json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"        json:"updated_at"`
}

type AICampaign struct {
	ID          string    `db:"id"          json:"id"`
	Title       string    `db:"title"       json:"title"`
	Description string    `db:"description" json:"description"`
	StartsOn    time.Time `db:"starts_on"   json:"starts_on"`
	EndsOn      time.Time `db:"ends_on"     json:"ends_on"`
	Active      bool      `db:"active"      json:"active"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

// AppSettings é singleton (sempre id=1) — configurações globais de IA/
// triagem + DefaultUnitID, a unidade pra onde cai conversa nova de canal
// centralizado (IG/FB) até um humano decidir a clínica física certa.
type AppSettings struct {
	ID                   int        `db:"id"                     json:"-"`
	AIEnabled            bool       `db:"ai_enabled"             json:"ai_enabled"`
	AISystemPrompt       string     `db:"ai_system_prompt"       json:"ai_system_prompt"`
	AIContext            string     `db:"ai_context"             json:"ai_context"`
	TriageEnabled        bool       `db:"triage_enabled"         json:"triage_enabled"`
	TriageHandoffIntents IntentTags `db:"triage_handoff_intents" json:"triage_handoff_intents"`
	DefaultUnitID        *string    `db:"default_unit_id"        json:"default_unit_id,omitempty"`
	UpdatedAt            time.Time  `db:"updated_at"             json:"updated_at"`
}

type EngagementType string

const (
	EngagementStoryReply   EngagementType = "story_reply"
	EngagementStoryMention EngagementType = "story_mention"
	EngagementPostComment  EngagementType = "post_comment"
	EngagementLiveComment  EngagementType = "live_comment"
	EngagementPrivateReply EngagementType = "private_reply"
)

type EngagementStatus string

const (
	EngagementOpen      EngagementStatus = "open"
	EngagementReplied   EngagementStatus = "replied"
	EngagementDismissed EngagementStatus = "dismissed"
	EngagementConverted EngagementStatus = "converted_to_conversation"
)

// SocialEngagement é uma interação Meta-nativa fora do fluxo 1:1 de
// mensagens — story reply/mention, comentário de post/live
// (docs/BACKEND-CONTRACT.md §5). CustomerID fica nil até author_external_id
// ser resolvido pra um Customer (mesma lógica de identidade de mensagens).
type SocialEngagement struct {
	ID               string           `db:"id"                  json:"id"`
	CustomerID       *string          `db:"customer_id"         json:"customer_id,omitempty"`
	Channel          Channel          `db:"channel"             json:"channel"`
	Type             EngagementType   `db:"type"                json:"type"`
	Status           EngagementStatus `db:"status"               json:"status"`
	UnitID           string           `db:"unit_id"             json:"unit_id"`
	MediaID          *string          `db:"media_id"            json:"media_id,omitempty"`
	MediaURL         *string          `db:"media_url"           json:"media_url,omitempty"`
	MediaCaption     *string          `db:"media_caption"       json:"media_caption,omitempty"`
	Body             string           `db:"body"                json:"body"`
	ExternalID       string           `db:"external_id"         json:"external_id"`
	AuthorExternalID string           `db:"author_external_id"  json:"author_external_id"`
	ConversationID   *string          `db:"conversation_id"     json:"conversation_id,omitempty"`
	CreatedAt        time.Time        `db:"created_at"          json:"created_at"`
	RepliedAt        *time.Time       `db:"replied_at"          json:"replied_at,omitempty"`
}

// AuditLog é append-only (docs/BACKEND-CONTRACT.md §9) — nada no backend
// faz UPDATE/DELETE nela, só INSERT via pkg/audit. Action segue a convenção
// "<recurso>.<verbo>" (ex.: "conversation.claim", "user.create").
type AuditLog struct {
	ID           string     `db:"id"            json:"id"`
	ActorUserID  *string    `db:"actor_user_id" json:"actor_user_id,omitempty"`
	ActorName    *string    `db:"actor_name"    json:"actor_name,omitempty"`
	Action       string     `db:"action"        json:"action"`
	ResourceType string     `db:"resource_type" json:"resource_type"`
	ResourceID   string     `db:"resource_id"   json:"resource_id"`
	UnitID       *string    `db:"unit_id"       json:"unit_id,omitempty"`
	Metadata     JSONObject `db:"metadata"      json:"metadata"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
}
