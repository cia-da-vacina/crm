// Package usecase implementa a caixa de entrada de engagements Meta-nativos
// (story reply/mention, comentário de post/live) — docs/BACKEND-CONTRACT.md
// §5. Diferente de conversation, engagements não têm mensagens: são itens de
// uma fila que um agente triagem manualmente (responder, dispensar ou
// converter numa conversa de verdade).
package usecase

import (
	"context"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/repository"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/cursor"
	"github.com/cia-da-vacina/crm/backend/pkg/meta"
	"github.com/cia-da-vacina/crm/backend/pkg/sse"
	"github.com/google/uuid"
)

type Repository interface {
	List(ctx context.Context, f model.ListFilter) ([]repository.Row, bool, error)
	GetByID(ctx context.Context, id string) (repository.Row, error)
	GetEntity(ctx context.Context, id string) (entity.SocialEngagement, error)
	UpdateEngagement(ctx context.Context, e entity.SocialEngagement) error
	CreateEngagement(ctx context.Context, e entity.SocialEngagement) (bool, error)
	GetCustomerIDByIdentity(ctx context.Context, channel entity.Channel, externalID string) (string, error)
	CreateCustomer(ctx context.Context, c entity.Customer) error
	CreateCustomerIdentity(ctx context.Context, ci entity.CustomerIdentity) error
	FindConversation(ctx context.Context, customerID string, channel entity.Channel) (entity.Conversation, error)
	CreateConversation(ctx context.Context, c entity.Conversation) error
}

// Access replica o shape usado nos outros módulos — cada um define o seu,
// mantendo-os independentes (ver backend/ARCHITECTURE.md).
type Access struct {
	Role    string
	UnitIDs []string
}

func (a Access) isAdmin() bool { return a.Role == string(entity.RoleAdmin) }

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
	repo   Repository
	meta   *meta.Registry
	sseHub *sse.Hub
}

func New(repo Repository, metaRegistry *meta.Registry, sseHub *sse.Hub) *UseCase {
	return &UseCase{repo: repo, meta: metaRegistry, sseHub: sseHub}
}

// IngestFromWebhook cria o engagement a partir de um evento vindo da
// ingestão de webhook (webhook/usecase chama isso via a interface
// Engagement, mesma convenção de Triage.RunTriage). CustomerID fica
// preenchido se a identidade do autor já existir — sem criar Customer novo
// aqui: isso só acontece em Convert, quando um agente decide puxar o
// engagement pro fluxo de conversa de verdade.
func (uc *UseCase) IngestFromWebhook(ctx context.Context, e entity.SocialEngagement) error {
	e.Status = entity.EngagementOpen

	customerID, err := uc.repo.GetCustomerIDByIdentity(ctx, e.Channel, e.AuthorExternalID)
	if err == nil {
		e.CustomerID = &customerID
	} else if !apperrors.IsNotFound(err) {
		return err
	}

	created, err := uc.repo.CreateEngagement(ctx, e)
	if err != nil {
		return err
	}
	if !created {
		return nil // external_id já visto — idempotência, não é erro
	}

	if uc.sseHub != nil {
		row, err := uc.repo.GetByID(ctx, e.ID)
		if err == nil {
			uc.sseHub.Publish(sse.Event{Name: "engagement.created", UnitID: e.UnitID, Data: toModel(row)})
		}
	}
	return nil
}

func (uc *UseCase) List(ctx context.Context, f model.ListFilter) (model.CursorPage, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 30
	}

	rows, hasMore, err := uc.repo.List(ctx, f)
	if err != nil {
		return model.CursorPage{}, apperrors.NewDatabaseError(err)
	}

	items := make([]model.SocialEngagement, len(rows))
	for i, row := range rows {
		items[i] = toModel(row)
	}

	var nextCursor *string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		c := cursor.Encode(last.CreatedAt, last.ID)
		nextCursor = &c
	}

	return model.CursorPage{Items: items, NextCursor: nextCursor}, nil
}

func (uc *UseCase) Get(ctx context.Context, id string, access Access) (model.SocialEngagement, error) {
	row, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.SocialEngagement{}, apperrors.NewNotFoundError("engagement")
		}
		return model.SocialEngagement{}, apperrors.NewDatabaseError(err)
	}
	if !access.canAccessUnit(row.UnitID) {
		return model.SocialEngagement{}, apperrors.NewNotFoundError("engagement")
	}
	return toModel(row), nil
}

// Dismiss marca o engagement como tratado sem responder — ex.: comentário
// irrelevante/spam. Não chama a API da Meta.
func (uc *UseCase) Dismiss(ctx context.Context, id string, access Access) (model.SocialEngagement, error) {
	e, err := uc.loadForMutation(ctx, id, access)
	if err != nil {
		return model.SocialEngagement{}, err
	}
	if e.Status != entity.EngagementOpen {
		return model.SocialEngagement{}, apperrors.NewConflictErrorMessage("engagement já foi tratado")
	}

	e.Status = entity.EngagementDismissed
	if err := uc.repo.UpdateEngagement(ctx, e); err != nil {
		return model.SocialEngagement{}, apperrors.NewDatabaseError(err)
	}

	row, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return model.SocialEngagement{}, apperrors.NewDatabaseError(err)
	}
	return toModel(row), nil
}

// Reply responde o engagement através da Meta. post_comment/live_comment
// respondem via Private Reply por padrão (não ReplyPublic) — o contrato
// deixa a escolha pública/privada "a definir pelo backend"
// (docs/BACKEND-CONTRACT.md §5); a existência de um tipo dedicado
// private_reply e o fato de um comentário público poder expor dados do
// cliente (evitar em contexto de saúde) tornam a DM o default mais seguro.
// story_reply/story_mention SÓ existem como DM (não há "público" pra story),
// então ReplyPrivate é o único caminho ali de qualquer forma. Fica registrado
// em backend/ARCHITECTURE.md como decisão de baixo risco, não bloqueante.
func (uc *UseCase) Reply(ctx context.Context, id string, req model.ReplyRequest, access Access) (model.SocialEngagement, error) {
	e, err := uc.loadForMutation(ctx, id, access)
	if err != nil {
		return model.SocialEngagement{}, err
	}
	if e.Status != entity.EngagementOpen {
		return model.SocialEngagement{}, apperrors.NewConflictErrorMessage("engagement já foi tratado")
	}

	responder, err := uc.meta.CommentResponder(meta.ChannelType(e.Channel))
	if err != nil {
		return model.SocialEngagement{}, apperrors.NewBadGatewayError(err.Error())
	}

	if _, err := responder.ReplyPrivate(ctx, meta.ReplyCommentInput{
		Channel:           meta.ChannelType(e.Channel),
		CommentExternalID: e.ExternalID,
		Body:              req.Body,
	}); err != nil {
		return model.SocialEngagement{}, apperrors.NewBadGatewayError(err.Error())
	}

	now := time.Now()
	e.Status = entity.EngagementReplied
	e.RepliedAt = &now
	if err := uc.repo.UpdateEngagement(ctx, e); err != nil {
		return model.SocialEngagement{}, apperrors.NewDatabaseError(err)
	}

	row, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return model.SocialEngagement{}, apperrors.NewDatabaseError(err)
	}
	return toModel(row), nil
}

// Convert promove o engagement pra uma conversa 1:1 de verdade — o agente
// decidiu que vale a pena atender esse contato pelo fluxo normal. Reusa
// Customer/Conversation existentes do mesmo autor+canal quando possível, em
// vez de sempre criar do zero. A conversa nasce em mode:"human" (não
// ai_triage): foi um agente que decidiu puxar essa conversa pra fila,
// diferente do primeiro contato via mensagem — não faz sentido devolver pra
// IA algo que um humano já escolheu tratar.
func (uc *UseCase) Convert(ctx context.Context, id string, access Access) (model.SocialEngagement, error) {
	e, err := uc.loadForMutation(ctx, id, access)
	if err != nil {
		return model.SocialEngagement{}, err
	}
	if e.Status == entity.EngagementConverted {
		return model.SocialEngagement{}, apperrors.NewConflictErrorMessage("engagement já foi convertido")
	}

	customerID, err := uc.resolveCustomer(ctx, e)
	if err != nil {
		return model.SocialEngagement{}, apperrors.NewDatabaseError(err)
	}

	conv, err := uc.repo.FindConversation(ctx, customerID, e.Channel)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			return model.SocialEngagement{}, apperrors.NewDatabaseError(err)
		}
		now := time.Now()
		conv = entity.Conversation{
			ID:            uuid.Must(uuid.NewV7()).String(),
			CustomerID:    customerID,
			Channel:       e.Channel,
			UnitID:        e.UnitID,
			PipelineStage: entity.StageEmAtendimento,
			Mode:          entity.ModeHuman,
			PhoneGate:     entity.PhoneGateNotNeeded,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := uc.repo.CreateConversation(ctx, conv); err != nil {
			return model.SocialEngagement{}, apperrors.NewDatabaseError(err)
		}
	}

	now := time.Now()
	e.CustomerID = &customerID
	e.ConversationID = &conv.ID
	e.Status = entity.EngagementConverted
	e.RepliedAt = &now
	if err := uc.repo.UpdateEngagement(ctx, e); err != nil {
		return model.SocialEngagement{}, apperrors.NewDatabaseError(err)
	}

	row, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return model.SocialEngagement{}, apperrors.NewDatabaseError(err)
	}
	return toModel(row), nil
}

func (uc *UseCase) resolveCustomer(ctx context.Context, e entity.SocialEngagement) (string, error) {
	if e.CustomerID != nil {
		return *e.CustomerID, nil
	}

	customerID, err := uc.repo.GetCustomerIDByIdentity(ctx, e.Channel, e.AuthorExternalID)
	if err == nil {
		return customerID, nil
	}
	if !apperrors.IsNotFound(err) {
		return "", err
	}

	now := time.Now()
	customerID = uuid.Must(uuid.NewV7()).String()
	if err := uc.repo.CreateCustomer(ctx, entity.Customer{
		ID:             customerID,
		DisplayName:    "Cliente " + string(e.Channel),
		Identification: entity.IdentificationAnonymous,
		UnitID:         &e.UnitID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		return "", err
	}
	if err := uc.repo.CreateCustomerIdentity(ctx, entity.CustomerIdentity{
		ID:         uuid.Must(uuid.NewV7()).String(),
		CustomerID: customerID,
		Channel:    e.Channel,
		ExternalID: e.AuthorExternalID,
		CreatedAt:  now,
	}); err != nil {
		return "", err
	}
	return customerID, nil
}

func (uc *UseCase) loadForMutation(ctx context.Context, id string, access Access) (entity.SocialEngagement, error) {
	e, err := uc.repo.GetEntity(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return entity.SocialEngagement{}, apperrors.NewNotFoundError("engagement")
		}
		return entity.SocialEngagement{}, apperrors.NewDatabaseError(err)
	}
	if !access.canAccessUnit(e.UnitID) {
		return entity.SocialEngagement{}, apperrors.NewNotFoundError("engagement")
	}
	return e, nil
}

func toModel(row repository.Row) model.SocialEngagement {
	return model.SocialEngagement{
		ID:               row.ID,
		CustomerID:       row.CustomerID,
		CustomerName:     row.CustomerName,
		Channel:          string(row.Channel),
		Type:             string(row.Type),
		Status:           string(row.Status),
		UnitID:           row.UnitID,
		MediaID:          row.MediaID,
		MediaURL:         row.MediaURL,
		MediaCaption:     row.MediaCaption,
		Body:             row.Body,
		ExternalID:       row.ExternalID,
		AuthorExternalID: row.AuthorExternalID,
		ConversationID:   row.ConversationID,
		CreatedAt:        row.CreatedAt,
		RepliedAt:        row.RepliedAt,
	}
}
