// Package usecase implementa a ingestão de webhook: parse do payload por
// canal, resolução de identidade/unidade, e criação de conversation/message
// (docs/BACKEND-CONTRACT.md §8).
package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/webhook/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/sse"
	"github.com/google/uuid"
)

type Repository interface {
	GetCustomerIDByIdentity(ctx context.Context, channel entity.Channel, externalID string) (string, error)
	GetCustomerByPhone(ctx context.Context, phone string) (entity.Customer, error)
	CreateCustomer(ctx context.Context, c entity.Customer) error
	CreateCustomerIdentity(ctx context.Context, ci entity.CustomerIdentity) error
	GetUnitIDByPhoneNumberID(ctx context.Context, phoneNumberID string) (string, error)
	GetDefaultUnitID(ctx context.Context) (*string, error)
	GetTriageEnabled(ctx context.Context) (bool, error)
	FindConversation(ctx context.Context, customerID string, channel entity.Channel) (entity.Conversation, error)
	CreateConversation(ctx context.Context, c entity.Conversation) error
	UpdateConversationAfterMessage(ctx context.Context, id, preview string, at, windowExpiresAt time.Time) error
	CreateMessage(ctx context.Context, msg entity.Message) (bool, error)
}

// Triage é o subconjunto do usecase de triagem que a ingestão precisa —
// depender só da interface (não do pacote inteiro) evita acoplar os dois
// módulos além do necessário, mesma convenção de CustomerReader em
// conversation/usecase.
type Triage interface {
	RunTriage(ctx context.Context, conversationID string) error
}

// Engagement é o subconjunto do usecase de engagement que a ingestão
// precisa — mesma convenção de Triage acima.
type Engagement interface {
	IngestFromWebhook(ctx context.Context, e entity.SocialEngagement) error
}

type UseCase struct {
	repo       Repository
	triage     Triage
	engagement Engagement
	sseHub     *sse.Hub
}

func New(repo Repository, triage Triage, engagement Engagement, sseHub *sse.Hub) *UseCase {
	return &UseCase{repo: repo, triage: triage, engagement: engagement, sseHub: sseHub}
}

// IngestPayload faz o parse do payload cru (já com a assinatura HMAC
// validada pelo handler) e ingere cada mensagem. Erro de parse volta pro
// handler (payload realmente malformado); erro por-mensagem durante a
// ingestão fica só logado — uma mensagem ruim não deve derrubar as outras
// do mesmo lote.
func (uc *UseCase) IngestPayload(ctx context.Context, channel entity.Channel, rawBody []byte) error {
	var (
		messages []model.InboundMessage
		err      error
	)

	switch channel {
	case entity.ChannelWhatsApp:
		messages, err = parseWhatsApp(rawBody)
	case entity.ChannelInstagram, entity.ChannelFacebook:
		messages, err = parseMessaging(rawBody, channel)
	default:
		return fmt.Errorf("unknown channel %q", channel)
	}
	if err != nil {
		return err
	}

	for _, m := range messages {
		if ierr := uc.ingestOne(ctx, m); ierr != nil {
			log.Printf("webhook: failed to ingest message (channel=%s external_id=%s meta_message_id=%s): %v",
				m.Channel, m.ExternalID, m.MetaMessageID, ierr)
		}
	}

	// Engagements Meta-nativos (story reply/mention, comentário) só existem
	// em Instagram/Facebook — WhatsApp não tem posts/stories/comments
	// (docs/BACKEND-CONTRACT.md §5). O mesmo payload cru pode carregar
	// mensagens E eventos de engagement (a Meta manda tudo dentro de
	// "entry"), então ambos os parses rodam sobre o rawBody original.
	if channel == entity.ChannelInstagram || channel == entity.ChannelFacebook {
		uc.ingestEngagements(ctx, channel, rawBody)
	}

	return nil
}

func (uc *UseCase) ingestEngagements(ctx context.Context, channel entity.Channel, rawBody []byte) {
	storyEvents, err := parseStoryEngagements(rawBody, channel)
	if err != nil {
		log.Printf("webhook: failed to parse story engagements (channel=%s): %v", channel, err)
	}
	commentEvents, err := parseComments(rawBody, channel)
	if err != nil {
		log.Printf("webhook: failed to parse comment engagements (channel=%s): %v", channel, err)
	}

	events := append(storyEvents, commentEvents...)
	if len(events) == 0 {
		return
	}

	defaultUnitID, err := uc.repo.GetDefaultUnitID(ctx)
	if err != nil {
		log.Printf("webhook: failed to resolve default unit for engagement ingestion (channel=%s): %v", channel, err)
		return
	}
	if defaultUnitID == nil {
		log.Printf("webhook: no default_unit_id configured for channel=%s — dropping %d engagement event(s), configure via PUT /settings/meta", channel, len(events))
		return
	}

	for _, ev := range events {
		e := entity.SocialEngagement{
			ID:               uuid.Must(uuid.NewV7()).String(),
			Channel:          ev.Channel,
			Type:             ev.Type,
			UnitID:           *defaultUnitID,
			MediaID:          ev.MediaID,
			MediaURL:         ev.MediaURL,
			MediaCaption:     ev.MediaCaption,
			Body:             ev.Body,
			ExternalID:       ev.ExternalID,
			AuthorExternalID: ev.AuthorExternalID,
			CreatedAt:        ev.Timestamp,
		}
		if ierr := uc.engagement.IngestFromWebhook(ctx, e); ierr != nil {
			log.Printf("webhook: failed to ingest engagement (channel=%s type=%s external_id=%s): %v",
				ev.Channel, ev.Type, ev.ExternalID, ierr)
		}
	}
}

func (uc *UseCase) ingestOne(ctx context.Context, m model.InboundMessage) error {
	customerID, unitID, err := uc.resolveCustomerAndUnit(ctx, m)
	if err != nil {
		return err
	}

	conv, err := uc.repo.FindConversation(ctx, customerID, m.Channel)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			return err
		}

		// Toda conversa nova nasce em IA — regra vale pra todos os canais
		// (docs/PRODUCT-V2.md §6) — a menos que o kill-switch triage_enabled
		// esteja desligado, aí vai direto pra fila humana sem quebrar o
		// resto do fluxo (mesma seção do produto).
		mode := entity.ModeAITriage
		triageEnabled, terr := uc.repo.GetTriageEnabled(ctx)
		if terr != nil {
			return terr
		}
		if !triageEnabled {
			mode = entity.ModeHuman
		}

		conv = entity.Conversation{
			ID:            uuid.Must(uuid.NewV7()).String(),
			CustomerID:    customerID,
			Channel:       m.Channel,
			UnitID:        unitID,
			PipelineStage: entity.StageEmAtendimento,
			Mode:          mode,
			// Sem IA rodando ainda pra decidir intent x phone_gate no
			// primeiro contato — default seguro é not_needed; RunTriage
			// (fase 7) e o agente via POST /conversations/:id/phone
			// continuam podendo mudar isso depois.
			PhoneGate: entity.PhoneGateNotNeeded,
			CreatedAt: m.Timestamp,
			UpdatedAt: m.Timestamp,
		}
		if err := uc.repo.CreateConversation(ctx, conv); err != nil {
			return err
		}
	}

	msg := entity.Message{
		ID:             uuid.Must(uuid.NewV7()).String(),
		ConversationID: conv.ID,
		Direction:      entity.DirectionIn,
		SenderType:     entity.SenderContact,
		Kind:           entity.KindText,
		Channel:        m.Channel,
		Body:           m.Body,
		Status:         entity.MessageStatusDelivered,
		MetaMessageID:  &m.MetaMessageID,
		CreatedAt:      m.Timestamp,
	}
	created, err := uc.repo.CreateMessage(ctx, msg)
	if err != nil {
		return err
	}
	if !created {
		return nil // meta_message_id já visto — idempotência, não é erro
	}

	// Janela de atendimento de 24h da Meta — primeira vez que fica realmente
	// preenchida (a fase 4 já expunha o campo no contrato, mas nada o setava
	// ainda por falta de ingestão real de mensagem).
	windowExpiresAt := m.Timestamp.Add(24 * time.Hour)
	if err := uc.repo.UpdateConversationAfterMessage(ctx, conv.ID, preview(m.Body), m.Timestamp, windowExpiresAt); err != nil {
		return err
	}

	if uc.sseHub != nil {
		uc.sseHub.Publish(sse.Event{Name: "message.created", UnitID: unitID, Data: msg})
	}

	// RunTriage já checa sozinho ai_enabled e se a conversa ainda está em
	// ai_triage (fica não-op se não estiver) — erro aqui não deve derrubar a
	// ingestão da mensagem em si, só é logado.
	if uc.triage != nil {
		if err := uc.triage.RunTriage(ctx, conv.ID); err != nil {
			log.Printf("webhook: triage failed for conversation %s: %v", conv.ID, err)
		}
	}

	return nil
}

// resolveCustomerAndUnit implementa o fluxo de identidade de
// docs/BACKEND-CONTRACT.md §3: WhatsApp nasce identified (a Meta já entrega
// o telefone — nunca se pede nem confirma por OTP); Instagram/Facebook
// nascem anonymous. Também resolve a unidade dona (roteamento automático
// por phone_number_id no WhatsApp; default_unit_id nos centralizados).
func (uc *UseCase) resolveCustomerAndUnit(ctx context.Context, m model.InboundMessage) (customerID, unitID string, err error) {
	if m.Channel == entity.ChannelWhatsApp {
		unitID, err = uc.repo.GetUnitIDByPhoneNumberID(ctx, m.PhoneNumberID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return "", "", fmt.Errorf("no unit configured for whatsapp phone_number_id=%s — configure via PUT /settings/meta first", m.PhoneNumberID)
			}
			return "", "", err
		}
	} else {
		defaultUnitID, derr := uc.repo.GetDefaultUnitID(ctx)
		if derr != nil {
			return "", "", derr
		}
		if defaultUnitID == nil {
			return "", "", fmt.Errorf("no default_unit_id configured for centralized channel %s — configure via PUT /settings/meta first", m.Channel)
		}
		unitID = *defaultUnitID
	}

	customerID, err = uc.repo.GetCustomerIDByIdentity(ctx, m.Channel, m.ExternalID)
	if err == nil {
		return customerID, unitID, nil
	}
	if !apperrors.IsNotFound(err) {
		return "", "", err
	}

	// Identidade nunca vista — cria Customer (+merge por telefone, só pro
	// WhatsApp, que já entrega prova de posse) e CustomerIdentity.
	now := m.Timestamp
	if m.Channel == entity.ChannelWhatsApp {
		phone := "+" + m.ExternalID

		existing, perr := uc.repo.GetCustomerByPhone(ctx, phone)
		switch {
		case perr == nil:
			// Telefone já existia (ex.: merge anterior por OTP confirmado a
			// partir de IG/FB) — só anexa a identidade WhatsApp a ele.
			customerID = existing.ID
		case apperrors.IsNotFound(perr):
			customerID = uuid.Must(uuid.NewV7()).String()
			displayName := m.DisplayHandle
			if displayName == "" {
				displayName = phone
			}
			if cerr := uc.repo.CreateCustomer(ctx, entity.Customer{
				ID:             customerID,
				DisplayName:    displayName,
				Identification: entity.IdentificationIdentified,
				PrimaryPhone:   &phone,
				UnitID:         &unitID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}); cerr != nil {
				return "", "", cerr
			}
		default:
			return "", "", perr
		}

		verifiedAt := now
		if err := uc.repo.CreateCustomerIdentity(ctx, entity.CustomerIdentity{
			ID:         uuid.Must(uuid.NewV7()).String(),
			CustomerID: customerID,
			Channel:    m.Channel,
			ExternalID: m.ExternalID,
			PhoneE164:  &phone,
			VerifiedAt: &verifiedAt,
			CreatedAt:  now,
		}); err != nil {
			return "", "", err
		}
		return customerID, unitID, nil
	}

	customerID = uuid.Must(uuid.NewV7()).String()
	displayName := m.DisplayHandle
	if displayName == "" {
		displayName = "Cliente " + string(m.Channel)
	}
	if err := uc.repo.CreateCustomer(ctx, entity.Customer{
		ID:             customerID,
		DisplayName:    displayName,
		Identification: entity.IdentificationAnonymous,
		UnitID:         &unitID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		return "", "", err
	}

	var handle *string
	if m.DisplayHandle != "" {
		handle = &m.DisplayHandle
	}
	if err := uc.repo.CreateCustomerIdentity(ctx, entity.CustomerIdentity{
		ID:            uuid.Must(uuid.NewV7()).String(),
		CustomerID:    customerID,
		Channel:       m.Channel,
		ExternalID:    m.ExternalID,
		DisplayHandle: handle,
		CreatedAt:     now,
	}); err != nil {
		return "", "", err
	}
	return customerID, unitID, nil
}

func preview(body string) string {
	const maxLen = 140
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen]
}
