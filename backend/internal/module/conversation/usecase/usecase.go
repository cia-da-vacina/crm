// Package usecase implementa as regras de negócio de inbox/conversas/
// mensagens/claim/pipeline e verificação de telefone (phone.go).
package usecase

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/domain/vo"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/repository"
	customermodel "github.com/cia-da-vacina/crm/backend/internal/module/customer/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/audit"
	"github.com/cia-da-vacina/crm/backend/pkg/cursor"
	"github.com/cia-da-vacina/crm/backend/pkg/meta"
	"github.com/cia-da-vacina/crm/backend/pkg/sse"
	"github.com/google/uuid"
)

// Repository agrega tudo que o usecase precisa do banco — conversations,
// messages, loss_reasons, phone_verifications e as mutações pontuais em
// customers/customer_identities que o merge por OTP exige (ver
// repository/customer.go).
type Repository interface {
	List(ctx context.Context, f model.InboxFilter) ([]repository.Row, bool, error)
	GetByID(ctx context.Context, id string) (repository.Row, error)
	GetConversation(ctx context.Context, id string) (entity.Conversation, error)
	UpdateConversation(ctx context.Context, c entity.Conversation) error
	ListMessages(ctx context.Context, conversationID string, before *time.Time, limit int) ([]entity.Message, bool, error)
	CreateMessage(ctx context.Context, msg entity.Message) error
	GetActiveLossReason(ctx context.Context, code string) (entity.LossReason, error)

	GetActivePendingVerification(ctx context.Context, conversationID string) (entity.PhoneVerification, error)
	GetActivePendingVerificationsBulk(ctx context.Context, conversationIDs []string) (map[string]entity.PhoneVerification, error)
	UpsertPendingVerification(ctx context.Context, pv entity.PhoneVerification) error
	UpdatePendingVerification(ctx context.Context, pv entity.PhoneVerification) error
	DeletePendingVerification(ctx context.Context, id string) error

	GetCustomerIdentityExternalID(ctx context.Context, customerID string, channel entity.Channel) (string, error)
	GetCustomerByPhone(ctx context.Context, phone string) (entity.Customer, error)
	MergeCustomerInto(ctx context.Context, sourceID, targetID string) error
	PromoteCustomerToIdentified(ctx context.Context, customerID, phone string) error
	SetCustomerIdentityVerified(ctx context.Context, customerID string, channel entity.Channel, phone string) error

	CreateFollowUpIfNotOpen(ctx context.Context, fu entity.FollowUp) error
}

// CustomerReader é o subconjunto do usecase de customer que este módulo
// precisa pra montar ConversationDetail.Customer (com identities embutido) —
// depender só da interface, não do pacote usecase inteiro, evita acoplar os
// dois módulos além do necessário.
type CustomerReader interface {
	Get(ctx context.Context, id string) (customermodel.Customer, error)
}

// PricingReader é o subconjunto do usecase de pricing que SendMessage
// precisa pra estimar Message.CostBRL no momento do envio (Frente A do plano
// de adaptação WhatsApp 2026) — mesma convenção de CustomerReader acima.
// Quando nil (ex.: testes que não passam pricing), SendMessage simplesmente
// não preenche o bloco de custo, sem quebrar o envio.
type PricingReader interface {
	GetRate(ctx context.Context, category entity.PricingCategory) (entity.MessagePricingRate, error)
}

// Access carrega quem está fazendo a chamada — extraído dos claims JWT pelo
// handler, nunca do corpo da request.
type Access struct {
	UserID  string
	Role    string
	UnitIDs []string
}

func (a Access) isAdmin() bool { return a.Role == string(entity.RoleAdmin) }

func (a Access) canReassign() bool {
	return a.Role == string(entity.RoleAdmin) || a.Role == string(entity.RoleManager) || a.Role == string(entity.RoleSupervisor)
}

func (a Access) canAccessUnit(unitID string) bool {
	if a.isAdmin() {
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
	repo     Repository
	customer CustomerReader
	sseHub   *sse.Hub
	meta     *meta.Registry
	audit    *audit.Logger
	pricing  PricingReader
}

func New(repo Repository, customerReader CustomerReader, sseHub *sse.Hub, metaRegistry *meta.Registry, auditLogger *audit.Logger, pricingReader PricingReader) *UseCase {
	return &UseCase{repo: repo, customer: customerReader, sseHub: sseHub, meta: metaRegistry, audit: auditLogger, pricing: pricingReader}
}

func (uc *UseCase) List(ctx context.Context, f model.InboxFilter) (model.CursorPage[model.ConversationSummary], error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 30
	}

	rows, hasMore, err := uc.repo.List(ctx, f)
	if err != nil {
		return model.CursorPage[model.ConversationSummary]{}, apperrors.NewDatabaseError(err)
	}

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	pending, err := uc.repo.GetActivePendingVerificationsBulk(ctx, ids)
	if err != nil {
		return model.CursorPage[model.ConversationSummary]{}, apperrors.NewDatabaseError(err)
	}

	now := time.Now()
	items := make([]model.ConversationSummary, len(rows))
	for i, row := range rows {
		items[i] = toSummary(row, pendingFor(pending, row.ID), now)
	}

	var nextCursor *string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		c := cursor.Encode(last.SortAt, last.ID)
		nextCursor = &c
	}

	return model.CursorPage[model.ConversationSummary]{Items: items, NextCursor: nextCursor}, nil
}

func (uc *UseCase) Get(ctx context.Context, id string, access Access) (model.ConversationDetail, error) {
	row, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.ConversationDetail{}, apperrors.NewNotFoundError("conversation")
		}
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}
	// 404 (não 403) pra fora do escopo de unidade — evita confirmar a
	// existência do recurso pra quem não deveria nem saber que ele existe.
	if !access.canAccessUnit(row.UnitID) {
		return model.ConversationDetail{}, apperrors.NewNotFoundError("conversation")
	}

	var pv *entity.PhoneVerification
	if row.PhoneGate == entity.PhoneGatePendingVerification {
		if found, err := uc.repo.GetActivePendingVerification(ctx, id); err == nil {
			pv = &found
		} else if !apperrors.IsNotFound(err) {
			return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
		}
	}

	customer, err := uc.customer.Get(ctx, row.CustomerID)
	if err != nil {
		return model.ConversationDetail{}, err
	}

	summary := toSummary(row, pv, time.Now())
	return model.ConversationDetail{
		ConversationSummary: summary,
		Customer:            customer,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}

func (uc *UseCase) ListMessages(ctx context.Context, conversationID, cursorRaw string, limit int, access Access) (model.CursorPage[model.Message], error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if err := uc.checkAccess(ctx, conversationID, access); err != nil {
		return model.CursorPage[model.Message]{}, err
	}

	beforeTime, _, err := cursor.Decode(cursorRaw)
	if err != nil {
		return model.CursorPage[model.Message]{}, apperrors.NewBadRequestError("invalid cursor")
	}
	var before *time.Time
	if cursorRaw != "" {
		before = &beforeTime
	}

	messages, hasMore, err := uc.repo.ListMessages(ctx, conversationID, before, limit)
	if err != nil {
		return model.CursorPage[model.Message]{}, apperrors.NewDatabaseError(err)
	}

	items := make([]model.Message, len(messages))
	for i, m := range messages {
		items[i] = toMessageModel(m)
	}

	var nextCursor *string
	if hasMore && len(messages) > 0 {
		last := messages[len(messages)-1]
		c := cursor.Encode(last.CreatedAt, last.ID)
		nextCursor = &c
	}

	return model.CursorPage[model.Message]{Items: items, NextCursor: nextCursor}, nil
}

// SendMessage só é permitido com a conversa em mode:"human" — enviar com
// mode:"ai_triage" é rejeitado com 409 independente do que a UI já bloqueia
// (docs/BACKEND-CONTRACT.md §4).
func (uc *UseCase) SendMessage(ctx context.Context, conversationID string, req model.SendMessageRequest, access Access) (model.Message, error) {
	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.Message{}, apperrors.NewNotFoundError("conversation")
		}
		return model.Message{}, apperrors.NewDatabaseError(err)
	}
	if !access.canAccessUnit(conv.UnitID) {
		return model.Message{}, apperrors.NewNotFoundError("conversation")
	}
	if conv.Mode != entity.ModeHuman {
		return model.Message{}, apperrors.NewConflictErrorMessage("conversation is not in human mode")
	}

	kind := entity.KindText
	if req.Kind != "" && req.Kind != string(entity.KindText) {
		return model.Message{}, apperrors.NewBadRequestError("only kind \"text\" is supported for now")
	}

	externalID, err := uc.repo.GetCustomerIdentityExternalID(ctx, conv.CustomerID, conv.Channel)
	if err != nil {
		return model.Message{}, apperrors.NewDatabaseError(err)
	}

	sender, err := uc.meta.Sender(toMetaChannel(conv.Channel))
	if err != nil {
		return model.Message{}, apperrors.NewBadGatewayError(err.Error())
	}
	result, err := sender.SendText(ctx, meta.SendTextInput{
		Recipient: meta.Recipient{Channel: toMetaChannel(conv.Channel), ExternalID: externalID},
		Body:      req.Body,
	})
	if err != nil {
		return model.Message{}, apperrors.NewBadGatewayError("failed to send message: " + err.Error())
	}

	now := time.Now()
	userID := access.UserID
	msg := entity.Message{
		ID:             uuid.Must(uuid.NewV7()).String(),
		ConversationID: conversationID,
		Direction:      entity.DirectionOut,
		SenderType:     entity.SenderAgent,
		SenderUserID:   &userID,
		Kind:           kind,
		Channel:        conv.Channel,
		Body:           req.Body,
		Status:         entity.MessageStatusSent,
		MetaMessageID:  &result.MetaMessageID,
		CreatedAt:      now,
	}
	// Texto livre 1:1 é sempre categoria "service" (guia WhatsApp 2026 —
	// nenhuma exceção de janela grátis se aplica aqui, já que CTWA/72h ainda
	// não é rastreado — ver docs/WHATSAPP-2026-ADAPTATION-PLAN.md Frente E).
	// pricing_confirmed fica false: é uma estimativa local no momento do
	// envio, não o que a Meta de fato cobrou (só confirmável via webhook de
	// status real, que ainda não existe client pra receber — Frente A nota).
	uc.applyEstimatedPricing(ctx, &msg, entity.PricingService)
	if err := uc.repo.CreateMessage(ctx, msg); err != nil {
		return model.Message{}, apperrors.NewDatabaseError(err)
	}

	conv.LastMessagePreview = preview(req.Body)
	conv.LastMessageAt = &now
	conv.UpdatedAt = now
	if err := uc.repo.UpdateConversation(ctx, conv); err != nil {
		return model.Message{}, apperrors.NewDatabaseError(err)
	}

	msgModel := toMessageModel(msg)
	uc.publish("message.created", conv.UnitID, msgModel)

	return msgModel, nil
}

// Claim é o handoff obrigatório IA -> humano. Agente comum só reivindica
// conversa livre (409 se já tem owner_id); admin/manager/supervisor também
// podem reatribuir uma conversa já reivindicada por outro (decisão de
// produto confirmada pro time — ver backend/ARCHITECTURE.md).
func (uc *UseCase) Claim(ctx context.Context, conversationID string, access Access) (model.ConversationDetail, error) {
	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.ConversationDetail{}, apperrors.NewNotFoundError("conversation")
		}
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}
	if !access.canAccessUnit(conv.UnitID) {
		return model.ConversationDetail{}, apperrors.NewNotFoundError("conversation")
	}

	if conv.OwnerID != nil && *conv.OwnerID != access.UserID && !access.canReassign() {
		return model.ConversationDetail{}, apperrors.NewConflictErrorMessage("conversation already claimed")
	}

	userID := access.UserID
	conv.OwnerID = &userID
	conv.Mode = entity.ModeHuman
	conv.UpdatedAt = time.Now()
	if err := uc.repo.UpdateConversation(ctx, conv); err != nil {
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}

	uc.publish("conversation.claimed", conv.UnitID, map[string]string{"conversation_id": conv.ID, "owner_id": userID})
	uc.logAudit(ctx, access.UserID, "conversation.claim", "conversation", conv.ID, conv.UnitID, map[string]any{"owner_id": userID})

	return uc.Get(ctx, conversationID, access)
}

// UpdatePipeline: qualquer transição entre estágios válidos é permitida —
// backend não impõe ordem (decisão de produto confirmada, mesma nota do
// Claim). reason_code é obrigatório e validado contra o catálogo quando
// stage é "nao_fechado" (docs/BACKEND-CONTRACT.md §4).
func (uc *UseCase) UpdatePipeline(ctx context.Context, conversationID string, req model.UpdatePipelineRequest, access Access) (model.ConversationDetail, error) {
	stage := entity.PipelineStage(req.Stage)
	if !entity.ValidPipelineStages[stage] {
		return model.ConversationDetail{}, apperrors.NewBadRequestError("invalid pipeline stage")
	}

	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.ConversationDetail{}, apperrors.NewNotFoundError("conversation")
		}
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}
	if !access.canAccessUnit(conv.UnitID) {
		return model.ConversationDetail{}, apperrors.NewNotFoundError("conversation")
	}

	var lossReasonLabel string
	if stage == entity.StageNaoFechado {
		if req.ReasonCode == "" {
			return model.ConversationDetail{}, apperrors.NewBadRequestError("reason_code is required for nao_fechado")
		}
		lossReason, err := uc.repo.GetActiveLossReason(ctx, req.ReasonCode)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return model.ConversationDetail{}, apperrors.NewBadRequestError("unknown reason_code")
			}
			return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
		}
		lossReasonLabel = lossReason.Label
		code := req.ReasonCode
		conv.LossReasonCode = &code
		if req.ReasonText != "" {
			text := req.ReasonText
			conv.LossReasonText = &text
		}
	} else {
		conv.LossReasonCode = nil
		conv.LossReasonText = nil
	}

	conv.PipelineStage = stage
	conv.UpdatedAt = time.Now()
	if err := uc.repo.UpdateConversation(ctx, conv); err != nil {
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}

	// Follow-up automático — docs/BACKEND-CONTRACT.md §6: "backend é
	// responsável por gerar/agendar follow-ups automaticamente (ex.: ao mover
	// para aguardando_fechamento ou nao_fechado)". CreateFollowUpIfNotOpen é
	// idempotente (índice único parcial + ON CONFLICT DO NOTHING), então
	// oscilar entre esses estágios não empilha follow-ups duplicados.
	if stage == entity.StageAguardandoFechamento || stage == entity.StageNaoFechado {
		if err := uc.createFollowUp(ctx, conv, stage, lossReasonLabel); err != nil {
			return model.ConversationDetail{}, err
		}
	}

	uc.publish("conversation.pipeline_changed", conv.UnitID, map[string]string{"conversation_id": conv.ID, "stage": string(stage)})
	uc.logAudit(ctx, access.UserID, "conversation.pipeline_change", "conversation", conv.ID, conv.UnitID, map[string]any{
		"stage": string(stage), "reason_code": req.ReasonCode,
	})

	return uc.Get(ctx, conversationID, access)
}

func (uc *UseCase) checkAccess(ctx context.Context, conversationID string, access Access) error {
	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("conversation")
		}
		return apperrors.NewDatabaseError(err)
	}
	if !access.canAccessUnit(conv.UnitID) {
		return apperrors.NewNotFoundError("conversation")
	}
	return nil
}

func (uc *UseCase) publish(name, unitID string, data any) {
	if uc.sseHub == nil {
		return
	}
	uc.sseHub.Publish(sse.Event{Name: name, UnitID: unitID, Data: data})
}

func (uc *UseCase) logAudit(ctx context.Context, actorUserID, action, resourceType, resourceID, unitID string, metadata map[string]any) {
	if uc.audit == nil {
		return
	}
	uc.audit.Log(ctx, audit.Entry{
		ActorUserID: &actorUserID, Action: action, ResourceType: resourceType,
		ResourceID: resourceID, UnitID: &unitID, Metadata: metadata,
	})
}

func pendingFor(m map[string]entity.PhoneVerification, conversationID string) *entity.PhoneVerification {
	if pv, ok := m[conversationID]; ok {
		return &pv
	}
	return nil
}

// toSummary deriva phone_gate/pending_phone_masked "de leitura" — se a
// pendência expirou (TTL estourado) mas ninguém ainda chamou resend/confirm
// pra persistir a reversão, a resposta já mostra "required" mesmo assim
// (ver docs/BACKEND-CONTRACT.md §3: "não deixar órfão em pending_verification
// indefinidamente"). A gravação efetiva no banco acontece na próxima
// mutação (initiate/resend/confirm), não aqui.
func toSummary(row repository.Row, pv *entity.PhoneVerification, now time.Time) model.ConversationSummary {
	effectiveGate := row.PhoneGate
	var maskedPhone *string

	if row.PhoneGate == entity.PhoneGatePendingVerification {
		if pv == nil || now.After(pv.ExpiresAt) {
			effectiveGate = entity.PhoneGateRequired
		} else {
			masked := vo.MaskPhone(pv.PhoneE164)
			maskedPhone = &masked
		}
	}

	status := "open"
	if row.PipelineStage == entity.StageFechado || row.PipelineStage == entity.StageNaoFechado {
		status = "closed"
	}

	return model.ConversationSummary{
		ID:                 row.ID,
		CustomerID:         row.CustomerID,
		CustomerName:       row.CustomerName,
		CustomerPhone:      row.CustomerPhone,
		Identification:     string(row.Identification),
		PhoneGate:          string(effectiveGate),
		PendingPhoneMasked: maskedPhone,
		Channel:            string(row.Channel),
		ChannelThreadID:    row.ChannelThreadID,
		UnitID:             row.UnitID,
		PipelineStage:      string(row.PipelineStage),
		Mode:               string(row.Mode),
		Status:             status,
		OwnerID:            row.OwnerID,
		Intent:             row.Intent,
		AISummary:          row.AISummary,
		TriageNotes:        row.TriageNotes,
		LastMessagePreview: row.LastMessagePreview,
		LastMessageAt:      row.LastMessageAt,
		WindowExpiresAt:    row.WindowExpiresAt,
		UnreadCount:        row.UnreadCount,
	}
}

func toMessageModel(m entity.Message) model.Message {
	msg := model.Message{
		ID:               m.ID,
		ConversationID:   m.ConversationID,
		Direction:        string(m.Direction),
		SenderType:       string(m.SenderType),
		Kind:             string(m.Kind),
		Channel:          string(m.Channel),
		Body:             m.Body,
		Status:           string(m.Status),
		MetaMessageID:    m.MetaMessageID,
		MediaURL:         m.MediaURL,
		MediaMimeType:    m.MediaMimeType,
		TemplateName:     m.TemplateName,
		PricingConfirmed: m.PricingConfirmed,
		CostBRL:          m.CostBRL,
		PricingBillable:  m.PricingBillable,
		CreatedAt:        m.CreatedAt,
	}
	if m.PricingCategory != nil {
		cat := string(*m.PricingCategory)
		msg.PricingCategory = &cat
	}
	return msg
}

func toMetaChannel(c entity.Channel) meta.ChannelType {
	return meta.ChannelType(c)
}

// applyEstimatedPricing preenche o bloco de custo de uma mensagem de saída a
// partir do rate card local (pkg pricing) — best-effort: se uc.pricing for
// nil (não injetado, ex.: algum teste que não precisa de custo) ou a
// categoria não estiver cadastrada, a mensagem é enviada e persistida do
// mesmo jeito, só sem PricingCategory/CostBRL preenchidos. Custo é sempre
// uma leitura auxiliar, nunca deve bloquear o envio de uma mensagem real.
func (uc *UseCase) applyEstimatedPricing(ctx context.Context, msg *entity.Message, category entity.PricingCategory) {
	if uc.pricing == nil {
		return
	}
	rate, err := uc.pricing.GetRate(ctx, category)
	if err != nil {
		return
	}
	cat := category
	billable := rate.Billable
	cost := rate.RateBRL
	msg.PricingCategory = &cat
	msg.PricingBillable = &billable
	msg.CostBRL = &cost
	msg.PricingConfirmed = false
}

func preview(body string) string {
	const maxLen = 140
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen]
}

// followUpDefaultWindow não é especificado numericamente pelo contrato (só
// "backend é responsável por gerar/agendar automaticamente") — é um default
// operacional, ajustável sem migração de schema, igual aos limites de OTP em
// phone.go.
const followUpDefaultWindow = 72 * time.Hour

func (uc *UseCase) createFollowUp(ctx context.Context, conv entity.Conversation, stage entity.PipelineStage, lossReasonLabel string) error {
	note := "Retomar contato — conversa movida para " + string(stage)
	if lossReasonLabel != "" {
		note += " (" + lossReasonLabel + ")"
	}

	fu := entity.FollowUp{
		ID:             uuid.Must(uuid.NewV7()).String(),
		ConversationID: conv.ID,
		CustomerID:     conv.CustomerID,
		UnitID:         conv.UnitID,
		PipelineStage:  stage,
		DueAt:          time.Now().Add(followUpDefaultWindow),
		Status:         entity.FollowUpOpen,
		Note:           note,
		CreatedAt:      time.Now(),
	}
	if err := uc.repo.CreateFollowUpIfNotOpen(ctx, fu); err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}
