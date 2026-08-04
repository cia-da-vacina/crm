package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/metasettings/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/audit"
	"github.com/cia-da-vacina/crm/backend/pkg/crypto"
	"github.com/google/uuid"
)

type Repository interface {
	GetSettings(ctx context.Context) (entity.AppSettings, error)
	UpdateSettings(ctx context.Context, s entity.AppSettings) error
	ListChannelConfigs(ctx context.Context) ([]entity.MetaChannelConfig, error)
	GetChannelConfig(ctx context.Context, channel entity.Channel, unitID *string) (entity.MetaChannelConfig, error)
	CreateChannelConfig(ctx context.Context, cfg entity.MetaChannelConfig) error
	UpdateChannelConfig(ctx context.Context, cfg entity.MetaChannelConfig) error
	ListCampaigns(ctx context.Context) ([]entity.AICampaign, error)
	GetCampaign(ctx context.Context, id string) (entity.AICampaign, error)
	CreateCampaign(ctx context.Context, c entity.AICampaign) error
	UpdateCampaign(ctx context.Context, c entity.AICampaign) error
}

type UseCase struct {
	repo   Repository
	cipher *crypto.Cipher
	audit  *audit.Logger
}

func New(repo Repository, cipher *crypto.Cipher, auditLogger *audit.Logger) *UseCase {
	return &UseCase{repo: repo, cipher: cipher, audit: auditLogger}
}

func (uc *UseCase) Get(ctx context.Context) (model.Settings, error) {
	settings, err := uc.repo.GetSettings(ctx)
	if err != nil {
		return model.Settings{}, apperrors.NewDatabaseError(err)
	}
	configs, err := uc.repo.ListChannelConfigs(ctx)
	if err != nil {
		return model.Settings{}, apperrors.NewDatabaseError(err)
	}
	campaigns, err := uc.repo.ListCampaigns(ctx)
	if err != nil {
		return model.Settings{}, apperrors.NewDatabaseError(err)
	}
	return toModel(settings, configs, campaigns), nil
}

// Update sempre busca do zero pra devolver ao final — docs/BACKEND-CONTRACT.md
// §7: "o frontend nunca reaproveita channel_tokens enviado como estado local".
func (uc *UseCase) Update(ctx context.Context, req model.UpdateSettingsRequest, actorUserID string) (model.Settings, error) {
	settings, err := uc.repo.GetSettings(ctx)
	if err != nil {
		return model.Settings{}, apperrors.NewDatabaseError(err)
	}

	if req.AIEnabled != nil {
		settings.AIEnabled = *req.AIEnabled
	}
	if req.AISystemPrompt != nil {
		settings.AISystemPrompt = *req.AISystemPrompt
	}
	if req.AIContext != nil {
		settings.AIContext = *req.AIContext
	}
	if req.TriageEnabled != nil {
		settings.TriageEnabled = *req.TriageEnabled
	}
	if req.TriageHandoffIntents != nil {
		settings.TriageHandoffIntents = entity.IntentTags(req.TriageHandoffIntents)
	}
	if req.DefaultUnitID != nil {
		settings.DefaultUnitID = req.DefaultUnitID
	}
	settings.UpdatedAt = time.Now()

	if err := uc.repo.UpdateSettings(ctx, settings); err != nil {
		return model.Settings{}, apperrors.MapDBError(err, map[string]string{
			"app_settings_default_unit_id_fkey": "unit",
		})
	}

	for _, item := range req.Channels {
		if err := uc.upsertChannel(ctx, item); err != nil {
			return model.Settings{}, err
		}
	}

	for _, item := range req.AICampaigns {
		if err := uc.upsertCampaign(ctx, item); err != nil {
			return model.Settings{}, err
		}
	}

	if uc.audit != nil {
		uc.audit.Log(ctx, audit.Entry{
			ActorUserID: &actorUserID, Action: "settings.meta.update",
			ResourceType: "app_settings", ResourceID: "1",
			Metadata: map[string]any{"channels_updated": len(req.Channels), "campaigns_updated": len(req.AICampaigns)},
		})
	}

	return uc.Get(ctx)
}

// upsertChannel valida a regra de negócio confirmada com o stakeholder:
// whatsapp é sempre por unidade, instagram/facebook são sempre centralizados
// (ver model.go e backend/ARCHITECTURE.md §5).
func (uc *UseCase) upsertChannel(ctx context.Context, item model.ChannelUpdateItem) error {
	channel := entity.Channel(item.Channel)

	if channel == entity.ChannelWhatsApp && item.UnitID == nil {
		return apperrors.NewBadRequestError("whatsapp channel requires unit_id")
	}
	if channel != entity.ChannelWhatsApp && item.UnitID != nil {
		return apperrors.NewBadRequestError(string(channel) + " is a centralized channel and must not set unit_id")
	}

	now := time.Now()
	existing, err := uc.repo.GetChannelConfig(ctx, channel, item.UnitID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			return apperrors.NewDatabaseError(err)
		}

		cfg := entity.MetaChannelConfig{
			ID:        uuid.Must(uuid.NewV7()).String(),
			Channel:   channel,
			UnitID:    item.UnitID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		applyChannelUpdate(&cfg, item)
		if item.Token != nil {
			if err := uc.setToken(&cfg, *item.Token); err != nil {
				return apperrors.NewInternalError(err)
			}
		}
		if err := uc.repo.CreateChannelConfig(ctx, cfg); err != nil {
			return apperrors.MapDBError(err, map[string]string{
				"meta_channel_configs_unit_id_fkey": "unit",
			})
		}
		return nil
	}

	applyChannelUpdate(&existing, item)
	existing.UpdatedAt = now
	if item.Token != nil {
		if err := uc.setToken(&existing, *item.Token); err != nil {
			return apperrors.NewInternalError(err)
		}
	}
	if err := uc.repo.UpdateChannelConfig(ctx, existing); err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// upsertCampaign: ID ausente cria; presente atualiza (404 se não existir —
// evita criar silenciosamente com um id que o client mandou por engano).
func (uc *UseCase) upsertCampaign(ctx context.Context, item model.CampaignUpdateItem) error {
	startsOn, err := time.Parse("2006-01-02", item.StartsOn)
	if err != nil {
		return apperrors.NewBadRequestError("starts_on must be YYYY-MM-DD")
	}
	endsOn, err := time.Parse("2006-01-02", item.EndsOn)
	if err != nil {
		return apperrors.NewBadRequestError("ends_on must be YYYY-MM-DD")
	}

	active := true
	if item.Active != nil {
		active = *item.Active
	}
	now := time.Now()

	if item.ID == nil {
		c := entity.AICampaign{
			ID:          uuid.Must(uuid.NewV7()).String(),
			Title:       item.Title,
			Description: item.Description,
			StartsOn:    startsOn,
			EndsOn:      endsOn,
			Active:      active,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := uc.repo.CreateCampaign(ctx, c); err != nil {
			return apperrors.NewDatabaseError(err)
		}
		return nil
	}

	existing, err := uc.repo.GetCampaign(ctx, *item.ID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("campaign")
		}
		return apperrors.NewDatabaseError(err)
	}

	existing.Title = item.Title
	existing.Description = item.Description
	existing.StartsOn = startsOn
	existing.EndsOn = endsOn
	existing.Active = active
	existing.UpdatedAt = now
	if err := uc.repo.UpdateCampaign(ctx, existing); err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func applyChannelUpdate(cfg *entity.MetaChannelConfig, item model.ChannelUpdateItem) {
	if item.Enabled != nil {
		cfg.Enabled = *item.Enabled
	}
	if item.AccountID != nil {
		cfg.AccountID = *item.AccountID
	}
	if item.DisplayName != nil {
		cfg.DisplayName = *item.DisplayName
	}
	if item.PhoneNumberID != nil {
		cfg.PhoneNumberID = item.PhoneNumberID
	}
}

func (uc *UseCase) setToken(cfg *entity.MetaChannelConfig, token string) error {
	ciphertext, err := uc.cipher.Encrypt(token)
	if err != nil {
		return err
	}
	masked := maskToken(token)
	cfg.TokenCiphertext = ciphertext
	cfg.TokenMasked = &masked
	return nil
}

// maskToken segue o formato do exemplo do contrato ("EAAG...9f2a").
func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func toModel(s entity.AppSettings, configs []entity.MetaChannelConfig, campaigns []entity.AICampaign) model.Settings {
	channels := make([]model.ChannelConfig, len(configs))
	for i, c := range configs {
		channels[i] = model.ChannelConfig{
			Channel:         string(c.Channel),
			UnitID:          c.UnitID,
			Enabled:         c.Enabled,
			AccountID:       c.AccountID,
			DisplayName:     c.DisplayName,
			PhoneNumberID:   c.PhoneNumberID,
			WebhookVerified: c.WebhookVerified,
			TokenMasked:     c.TokenMasked,
		}
	}

	campaignModels := make([]model.Campaign, len(campaigns))
	for i, c := range campaigns {
		campaignModels[i] = model.Campaign{
			ID:          c.ID,
			Title:       c.Title,
			Description: c.Description,
			StartsOn:    c.StartsOn,
			EndsOn:      c.EndsOn,
			Active:      c.Active,
		}
	}

	tags := []string(s.TriageHandoffIntents)
	if tags == nil {
		tags = []string{}
	}

	return model.Settings{
		Channels:             channels,
		AIEnabled:            s.AIEnabled,
		AISystemPrompt:       s.AISystemPrompt,
		AIContext:            s.AIContext,
		AICampaigns:          campaignModels,
		TriageEnabled:        s.TriageEnabled,
		TriageHandoffIntents: tags,
		DefaultUnitID:        s.DefaultUnitID,
	}
}
