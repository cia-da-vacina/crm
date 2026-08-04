// Package model define os DTOs de GET/PUT /settings/meta.
//
// Duas extensões deliberadas em relação ao shape documentado em
// docs/BACKEND-CONTRACT.md §7, confirmadas com o stakeholder (ver
// backend/ARCHITECTURE.md §5) porque o contrato original assumia canais
// sempre por unidade:
//   - ChannelConfig ganha unit_id (obrigatório pra whatsapp, ausente pra
//     instagram/facebook — que são centralizados, uma conta pra todas as
//     5 unidades).
//   - Settings ganha default_unit_id: unidade pra onde cai conversa nova de
//     canal centralizado, já que não há phone_number_id pra rotear sozinho.
//   - channel_tokens (mapa global do contrato original) vira um campo
//     Token por item de Channels — necessário porque "whatsapp" sozinho não
//     identifica qual das 5 unidades rotacionar.
package model

import "time"

type ChannelConfig struct {
	Channel         string  `json:"channel"`
	UnitID          *string `json:"unit_id,omitempty"`
	Enabled         bool    `json:"enabled"`
	AccountID       string  `json:"account_id"`
	DisplayName     string  `json:"display_name"`
	PhoneNumberID   *string `json:"phone_number_id,omitempty"`
	WebhookVerified bool    `json:"webhook_verified"`
	TokenMasked     *string `json:"token_masked,omitempty"`
}

type Campaign struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartsOn    time.Time `json:"starts_on"`
	EndsOn      time.Time `json:"ends_on"`
	Active      bool      `json:"active"`
}

type Settings struct {
	Channels             []ChannelConfig `json:"channels"`
	AIEnabled            bool            `json:"ai_enabled"`
	AISystemPrompt       string          `json:"ai_system_prompt"`
	AIContext            string          `json:"ai_context"`
	AICampaigns          []Campaign      `json:"ai_campaigns"`
	TriageEnabled        bool            `json:"triage_enabled"`
	TriageHandoffIntents []string        `json:"triage_handoff_intents"`
	DefaultUnitID        *string         `json:"default_unit_id,omitempty"`
}

// ChannelUpdateItem: Token, quando presente (mesmo vazio não — só quando o
// campo aparece no JSON), rotaciona o token daquele canal/unidade. É o único
// jeito de setar um token; nunca é ecoado de volta.
type ChannelUpdateItem struct {
	Channel       string  `json:"channel" validate:"required,oneof=whatsapp instagram facebook"`
	UnitID        *string `json:"unit_id"`
	Enabled       *bool   `json:"enabled"`
	AccountID     *string `json:"account_id"`
	DisplayName   *string `json:"display_name"`
	PhoneNumberID *string `json:"phone_number_id"`
	Token         *string `json:"token"`
}

// CampaignUpdateItem: ID ausente/nil cria uma campanha nova; ID presente
// atualiza a existente. Não há semântica de "array completo apaga o que
// falta" — cada item é um upsert independente (evita perda de dado acidental
// se o client mandar uma lista parcial); exclusão continua sendo uma
// operação futura fora de escopo (mesmo status de "merge review" em customers).
type CampaignUpdateItem struct {
	ID          *string `json:"id"`
	Title       string  `json:"title"       validate:"required"`
	Description string  `json:"description"`
	StartsOn    string  `json:"starts_on"   validate:"required"` // "2026-05-01"
	EndsOn      string  `json:"ends_on"     validate:"required"`
	Active      *bool   `json:"active"`
}

type UpdateSettingsRequest struct {
	// dive: sem isso o validator não desce nos itens do slice (comportamento
	// padrão do go-playground/validator) e as tags dos itens nunca rodariam.
	Channels             []ChannelUpdateItem  `json:"channels" validate:"dive"`
	AIEnabled            *bool                `json:"ai_enabled"`
	AISystemPrompt       *string              `json:"ai_system_prompt"`
	AIContext            *string              `json:"ai_context"`
	AICampaigns          []CampaignUpdateItem `json:"ai_campaigns" validate:"dive"`
	TriageEnabled        *bool                `json:"triage_enabled"`
	TriageHandoffIntents []string             `json:"triage_handoff_intents"`
	DefaultUnitID        *string              `json:"default_unit_id"`
}
