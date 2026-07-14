package domain

import (
	"time"

	"github.com/google/uuid"
)

type PipelineStage string

const (
	StageEmAtendimento         PipelineStage = "em_atendimento"
	StageEmNegociacao          PipelineStage = "em_negociacao"
	StageAguardandoFechamento  PipelineStage = "aguardando_fechamento"
	StageFechado               PipelineStage = "fechado"
	StageNaoFechado            PipelineStage = "nao_fechado"
)

type ConversationMode string

const (
	ModeAITriage ConversationMode = "ai_triage"
	ModeHuman    ConversationMode = "human"
)

type Conversation struct {
	ID                 uuid.UUID        `json:"id"`
	ContactName        string           `json:"contact_name"`
	ContactPhone       string           `json:"contact_phone"`
	UnitID             uuid.UUID        `json:"unit_id"`
	PipelineStage      PipelineStage    `json:"pipeline_stage"`
	Mode               ConversationMode `json:"mode"`
	OwnerID            *uuid.UUID       `json:"owner_id"`
	Intent             *string          `json:"intent"`
	AISummary          *string          `json:"ai_summary"`
	LastMessagePreview string           `json:"last_message_preview"`
	LastMessageAt      time.Time        `json:"last_message_at"`
	WindowExpiresAt    *time.Time       `json:"window_expires_at,omitempty"`
}

type Message struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	Direction      string    `json:"direction"`
	SenderType     string    `json:"sender_type"`
	Body           string    `json:"body"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type FollowUp struct {
	ID             uuid.UUID     `json:"id"`
	ConversationID uuid.UUID     `json:"conversation_id"`
	ContactName    string        `json:"contact_name"`
	ContactPhone   string        `json:"contact_phone"`
	UnitID         uuid.UUID     `json:"unit_id"`
	PipelineStage  PipelineStage `json:"pipeline_stage"`
	DueAt          time.Time     `json:"due_at"`
	Status         string        `json:"status"`
	Note           string        `json:"note"`
}

type Pop struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	IntentTags []string  `json:"intent_tags"`
	Active     bool      `json:"active"`
}

type LossReason struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

// AICampaign is seasonal copy the triage bot may mention to customers.
type AICampaign struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartsOn    string `json:"starts_on"` // YYYY-MM-DD
	EndsOn      string `json:"ends_on"`   // YYYY-MM-DD
	Active      bool   `json:"active"`
}

type WhatsAppSettings struct {
	WABAID          string       `json:"waba_id"`
	PhoneNumberID   string       `json:"phone_number_id"`
	DisplayPhone    string       `json:"display_phone"`
	TokenMasked     string       `json:"token_masked"`
	AIEnabled       bool         `json:"ai_enabled"`
	WebhookVerified bool         `json:"webhook_verified"`
	AISystemPrompt  string       `json:"ai_system_prompt"`
	AIContext       string       `json:"ai_context"`
	AICampaigns     []AICampaign `json:"ai_campaigns"`
}

type DashboardSummary struct {
	OpenConversations int                        `json:"open_conversations"`
	ByStage           map[PipelineStage]int      `json:"by_stage"`
	Closed            int                        `json:"closed"`
	NotClosed         int                        `json:"not_closed"`
	ConversionRate    int                        `json:"conversion_rate"`
	AITriage          int                        `json:"ai_triage"`
	Human             int                        `json:"human"`
	AwaitingFollowup  int                        `json:"awaiting_followup"`
	Units             []DashboardUnitSummary     `json:"units"`
}

type DashboardUnitSummary struct {
	UnitID         uuid.UUID `json:"unit_id"`
	UnitName       string    `json:"unit_name"`
	Open           int       `json:"open"`
	Closed         int       `json:"closed"`
	ConversionRate int       `json:"conversion_rate"`
}
