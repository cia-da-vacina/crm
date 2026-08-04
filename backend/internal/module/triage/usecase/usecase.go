// Package usecase implementa a triagem por IA: RunTriage roda quando uma
// mensagem chega numa conversa ainda em ai_triage (chamado pelo módulo
// webhook); Get serve GET /conversations/:id/triage a partir do que já foi
// persistido — a IA nunca é chamada durante uma leitura
// (docs/BACKEND-CONTRACT.md §4: "conteúdo 100% gerado pela IA... o frontend
// só exibe").
package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/domain/vo"
	"github.com/cia-da-vacina/crm/backend/internal/module/triage/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/meta"
	"github.com/cia-da-vacina/crm/backend/pkg/openai"
	"github.com/cia-da-vacina/crm/backend/pkg/sse"
	"github.com/google/uuid"
)

const defaultRecentMessagesLimit = 20

type Repository interface {
	GetConversation(ctx context.Context, id string) (entity.Conversation, error)
	UpdateConversation(ctx context.Context, c entity.Conversation) error
	GetCustomerIdentification(ctx context.Context, customerID string) (entity.CustomerIdentification, error)
	GetCustomerIdentityExternalID(ctx context.Context, customerID string, channel entity.Channel) (string, error)
	GetRecentMessages(ctx context.Context, conversationID string, limit int) ([]entity.Message, error)
	CreateMessage(ctx context.Context, msg entity.Message) error
	GetAppSettings(ctx context.Context) (entity.AppSettings, error)
	GetActiveCampaigns(ctx context.Context) ([]entity.AICampaign, error)
	GetSuggestedPopIDs(ctx context.Context, intent string) ([]string, error)
	GetActivePendingVerification(ctx context.Context, conversationID string) (entity.PhoneVerification, error)
}

type Access struct {
	Role    string
	UnitIDs []string
}

func (a Access) canAccessUnit(unitID string) bool {
	if a.Role == string(entity.RoleAdmin) {
		return true
	}
	for _, id := range a.UnitIDs {
		if id == unitID {
			return true
		}
	}
	return false
}

type UseCase struct {
	repo   Repository
	ai     openai.Client
	meta   *meta.Registry
	sseHub *sse.Hub
	model  string
}

func New(repo Repository, ai openai.Client, metaRegistry *meta.Registry, sseHub *sse.Hub, model string) *UseCase {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &UseCase{repo: repo, ai: ai, meta: metaRegistry, sseHub: sseHub, model: model}
}

// RunTriage é chamado pelo módulo webhook depois de ingerir uma mensagem
// inbound. Kill-switch: se ai_enabled=false ou a conversa já foi assumida
// por humano, não faz nada — silencioso, não é erro (a conversa só fica
// esperando claim manual).
func (uc *UseCase) RunTriage(ctx context.Context, conversationID string) error {
	settings, err := uc.repo.GetAppSettings(ctx)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	if !settings.AIEnabled {
		return nil
	}

	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	if conv.Mode != entity.ModeAITriage {
		return nil
	}

	identification, err := uc.repo.GetCustomerIdentification(ctx, conv.CustomerID)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}

	messages, err := uc.repo.GetRecentMessages(ctx, conversationID, defaultRecentMessagesLimit)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}

	campaigns, err := uc.repo.GetActiveCampaigns(ctx)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}

	prompt := buildPrompt(settings, conv, identification, messages, campaigns)

	result, err := uc.ai.Complete(ctx, openai.CompletionRequest{
		Model:        uc.model,
		Messages:     prompt,
		JSONResponse: true,
		Temperature:  0.3,
	})
	if err != nil {
		return fmt.Errorf("openai completion failed: %w", err)
	}

	parsed, err := parseTriageResponse(result.Content)
	if err != nil {
		return err
	}

	// Regra determinística por cima do julgamento do modelo — nunca confia
	// cegamente no que ele decidir pra WhatsApp ou pra quem já é identified
	// (ver docs/BACKEND-CONTRACT.md §3 e prompt.go).
	phoneGateRequired := parsed.PhoneGateRequired &&
		conv.Channel != entity.ChannelWhatsApp &&
		identification != entity.IdentificationIdentified

	phoneGate := conv.PhoneGate
	if phoneGateRequired && phoneGate == entity.PhoneGateNotNeeded {
		phoneGate = entity.PhoneGateRequired
	}

	readyForHandoff := parsed.ReadyForHandoff
	for _, handoffIntent := range settings.TriageHandoffIntents {
		if handoffIntent == parsed.Intent {
			readyForHandoff = true
			break
		}
	}

	now := time.Now()
	intent := parsed.Intent
	summary := parsed.Summary
	confidence := parsed.Confidence

	conv.Intent = &intent
	conv.AISummary = &summary
	if parsed.InternalNotes != "" {
		notes := parsed.InternalNotes
		conv.TriageNotes = &notes
	}
	conv.PhoneGate = phoneGate
	conv.CollectedFields = entity.JSONObject(parsed.CollectedFields)
	conv.TriageConfidence = &confidence
	conv.TriageReadyForHandoff = readyForHandoff
	conv.UpdatedAt = now

	if parsed.Reply != "" {
		if err := uc.sendReply(ctx, conv, parsed.Reply); err != nil {
			// Falha ao mandar a resposta não deve impedir de salvar a
			// triagem (intent/summary já são úteis pro agente mesmo sem a
			// IA ter conseguido responder) — só loga.
			log.Printf("triage: failed to send AI reply for conversation %s: %v", conv.ID, err)
		} else {
			conv.LastMessagePreview = preview(parsed.Reply)
			conv.LastMessageAt = &now
		}
	}

	if err := uc.repo.UpdateConversation(ctx, conv); err != nil {
		return apperrors.NewDatabaseError(err)
	}

	if uc.sseHub != nil {
		uc.sseHub.Publish(sse.Event{
			Name:   "conversation.triage_updated",
			UnitID: conv.UnitID,
			Data:   map[string]any{"conversation_id": conv.ID, "intent": intent, "ready_for_handoff": readyForHandoff},
		})
	}

	return nil
}

func (uc *UseCase) sendReply(ctx context.Context, conv entity.Conversation, reply string) error {
	externalID, err := uc.repo.GetCustomerIdentityExternalID(ctx, conv.CustomerID, conv.Channel)
	if err != nil {
		return err
	}

	sender, err := uc.meta.Sender(meta.ChannelType(conv.Channel))
	if err != nil {
		return err
	}

	result, err := sender.SendText(ctx, meta.SendTextInput{
		Recipient: meta.Recipient{Channel: meta.ChannelType(conv.Channel), ExternalID: externalID},
		Body:      reply,
	})
	if err != nil {
		return err
	}

	return uc.repo.CreateMessage(ctx, entity.Message{
		ID:             uuid.Must(uuid.NewV7()).String(),
		ConversationID: conv.ID,
		Direction:      entity.DirectionOut,
		SenderType:     entity.SenderAI,
		Kind:           entity.KindText,
		Channel:        conv.Channel,
		Body:           reply,
		Status:         entity.MessageStatusSent,
		MetaMessageID:  &result.MetaMessageID,
		CreatedAt:      time.Now(),
	})
}

// Get serve GET /conversations/:id/triage. 404 é o comportamento esperado
// quando a conversa já foi assumida por humano — o frontend trata isso
// silenciosamente, não é um erro real (docs/BACKEND-CONTRACT.md §4).
func (uc *UseCase) Get(ctx context.Context, conversationID string, access Access) (model.TriageSummary, error) {
	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.TriageSummary{}, apperrors.NewNotFoundError("triage")
		}
		return model.TriageSummary{}, apperrors.NewDatabaseError(err)
	}
	if !access.canAccessUnit(conv.UnitID) {
		return model.TriageSummary{}, apperrors.NewNotFoundError("triage")
	}
	if conv.Mode != entity.ModeAITriage {
		return model.TriageSummary{}, apperrors.NewNotFoundError("triage")
	}

	var suggestedPops []string
	if conv.Intent != nil {
		if ids, err := uc.repo.GetSuggestedPopIDs(ctx, *conv.Intent); err == nil {
			suggestedPops = ids
		}
	}

	var pendingPhoneMasked *string
	effectivePhoneGate := conv.PhoneGate
	if conv.PhoneGate == entity.PhoneGatePendingVerification {
		pv, err := uc.repo.GetActivePendingVerification(ctx, conversationID)
		switch {
		case err == nil && time.Now().Before(pv.ExpiresAt):
			masked := vo.MaskPhone(pv.PhoneE164)
			pendingPhoneMasked = &masked
		default:
			effectivePhoneGate = entity.PhoneGateRequired
		}
	}

	summary := ""
	if conv.AISummary != nil {
		summary = *conv.AISummary
	}

	return model.TriageSummary{
		ConversationID:     conv.ID,
		Intent:             conv.Intent,
		Confidence:         conv.TriageConfidence,
		Summary:            summary,
		SuggestedPops:      suggestedPops,
		ReadyForHandoff:    conv.TriageReadyForHandoff,
		PhoneGate:          string(effectivePhoneGate),
		PendingPhoneMasked: pendingPhoneMasked,
		CollectedFields:    conv.CollectedFields,
	}, nil
}

func preview(body string) string {
	const maxLen = 140
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen]
}
