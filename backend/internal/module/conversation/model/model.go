// Package model define os DTOs de request/response do módulo conversation —
// inbox, conversas, mensagens e o fluxo de verificação de telefone.
package model

import (
	"time"

	customermodel "github.com/cia-da-vacina/crm/backend/internal/module/customer/model"
)

// ConversationSummary = InboxItem do contrato (docs/BACKEND-CONTRACT.md §4) —
// mesmo shape usado em GET /inbox e como base de ConversationDetail.
type ConversationSummary struct {
	ID                 string     `json:"id"`
	CustomerID         string     `json:"customer_id"`
	CustomerName       string     `json:"customer_name"`
	CustomerPhone      *string    `json:"customer_phone,omitempty"`
	Identification     string     `json:"identification"`
	PhoneGate          string     `json:"phone_gate"`
	PendingPhoneMasked *string    `json:"pending_phone_masked,omitempty"`
	Channel            string     `json:"channel"`
	ChannelThreadID    *string    `json:"channel_thread_id,omitempty"`
	UnitID             string     `json:"unit_id"`
	PipelineStage      string     `json:"pipeline_stage"`
	Mode               string     `json:"mode"`
	Status             string     `json:"status"`
	OwnerID            *string    `json:"owner_id,omitempty"`
	Intent             *string    `json:"intent,omitempty"`
	AISummary          *string    `json:"ai_summary,omitempty"`
	TriageNotes        *string    `json:"triage_notes,omitempty"`
	LastMessagePreview string     `json:"last_message_preview"`
	LastMessageAt      *time.Time `json:"last_message_at,omitempty"`
	WindowExpiresAt    *time.Time `json:"window_expires_at,omitempty"`
	UnreadCount        int        `json:"unread_count,omitempty"`
}

// ConversationDetail = ConversationSummary + customer completo (docs/BACKEND-CONTRACT.md §4).
type ConversationDetail struct {
	ConversationSummary
	Customer  customermodel.Customer `json:"customer"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type Message struct {
	ID                  string  `json:"id"`
	ConversationID      string  `json:"conversation_id"`
	Direction           string  `json:"direction"`
	SenderType          string  `json:"sender_type"`
	Kind                string  `json:"kind"`
	Channel             string  `json:"channel"`
	Body                string  `json:"body"`
	Status              string  `json:"status"`
	MetaMessageID       *string `json:"meta_message_id,omitempty"`
	ReplyToEngagementID *string `json:"reply_to_engagement_id,omitempty"`
	MediaURL            *string `json:"media_url,omitempty"`
	MediaMimeType       *string `json:"media_mime_type,omitempty"`
	TemplateName        *string `json:"template_name,omitempty"`
	// Bloco de custo (docs/WHATSAPP-2026-ADAPTATION-PLAN.md, Frente A) — só
	// preenchido em mensagens de saída; PricingConfirmed=false indica
	// estimativa local (sem client Meta real pra reconciliar via webhook de
	// status ainda, ver backend/ARCHITECTURE.md §5).
	PricingCategory  *string   `json:"pricing_category,omitempty"`
	PricingBillable  *bool     `json:"pricing_billable,omitempty"`
	PricingConfirmed bool      `json:"pricing_confirmed"`
	CostBRL          *float64  `json:"cost_brl,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type CursorPage[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

// InboxFilter carrega os filtros de query de GET /inbox + o escopo de
// visibilidade do requester (mesma "visão de unidade" de users/units/customers).
type InboxFilter struct {
	UnitID           string
	Stage            string
	Channel          string
	Mode             string
	Cursor           string
	Limit            int
	Unscoped         bool
	RequesterUnitIDs []string
}

type SendMessageRequest struct {
	Body string `json:"body" validate:"required"`
	Kind string `json:"kind"`
}

type UpdatePipelineRequest struct {
	Stage      string `json:"stage" validate:"required"`
	ReasonCode string `json:"reason_code"`
	ReasonText string `json:"reason_text"`
}

type StartPhoneVerificationRequest struct {
	PhoneE164 string `json:"phone_e164" validate:"required,e164"`
}

type ConfirmPhoneVerificationRequest struct {
	Code string `json:"code" validate:"required"`
}
