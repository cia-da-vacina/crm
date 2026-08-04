package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

const channelColumns = `id, channel, unit_id, enabled, account_id, display_name, phone_number_id, webhook_verified, token_ciphertext, token_masked, created_at, updated_at`

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetSettings(ctx context.Context) (entity.AppSettings, error) {
	var s entity.AppSettings
	err := r.db.GetContext(ctx, &s, `
		SELECT id, ai_enabled, ai_system_prompt, ai_context, triage_enabled, triage_handoff_intents, default_unit_id, updated_at
		FROM app_settings WHERE id = 1
	`)
	return s, err
}

func (r *Repository) UpdateSettings(ctx context.Context, s entity.AppSettings) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE app_settings SET
			ai_enabled = :ai_enabled, ai_system_prompt = :ai_system_prompt, ai_context = :ai_context,
			triage_enabled = :triage_enabled, triage_handoff_intents = :triage_handoff_intents,
			default_unit_id = :default_unit_id, updated_at = :updated_at
		WHERE id = 1
	`, s)
	return err
}

func (r *Repository) ListChannelConfigs(ctx context.Context) ([]entity.MetaChannelConfig, error) {
	var configs []entity.MetaChannelConfig
	err := r.db.SelectContext(ctx, &configs, `SELECT `+channelColumns+` FROM meta_channel_configs ORDER BY channel, unit_id`)
	return configs, err
}

// GetChannelConfig usa "IS NOT DISTINCT FROM" (não "=") pra unitID nil
// bater com unit_id IS NULL — é assim que instagram/facebook (centralizados)
// são identificados unicamente.
func (r *Repository) GetChannelConfig(ctx context.Context, channel entity.Channel, unitID *string) (entity.MetaChannelConfig, error) {
	var cfg entity.MetaChannelConfig
	err := r.db.GetContext(ctx, &cfg, `
		SELECT `+channelColumns+` FROM meta_channel_configs
		WHERE channel = $1 AND unit_id IS NOT DISTINCT FROM $2
	`, channel, unitID)
	return cfg, err
}

func (r *Repository) CreateChannelConfig(ctx context.Context, cfg entity.MetaChannelConfig) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO meta_channel_configs (id, channel, unit_id, enabled, account_id, display_name, phone_number_id, webhook_verified, token_ciphertext, token_masked, created_at, updated_at)
		VALUES (:id, :channel, :unit_id, :enabled, :account_id, :display_name, :phone_number_id, :webhook_verified, :token_ciphertext, :token_masked, :created_at, :updated_at)
	`, cfg)
	return err
}

func (r *Repository) UpdateChannelConfig(ctx context.Context, cfg entity.MetaChannelConfig) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE meta_channel_configs SET
			enabled = :enabled, account_id = :account_id, display_name = :display_name,
			phone_number_id = :phone_number_id, token_ciphertext = :token_ciphertext,
			token_masked = :token_masked, updated_at = :updated_at
		WHERE id = :id
	`, cfg)
	return err
}

func (r *Repository) ListCampaigns(ctx context.Context) ([]entity.AICampaign, error) {
	var campaigns []entity.AICampaign
	err := r.db.SelectContext(ctx, &campaigns, `
		SELECT id, title, description, starts_on, ends_on, active, created_at, updated_at
		FROM ai_campaigns ORDER BY starts_on DESC
	`)
	return campaigns, err
}

func (r *Repository) GetCampaign(ctx context.Context, id string) (entity.AICampaign, error) {
	var c entity.AICampaign
	err := r.db.GetContext(ctx, &c, `
		SELECT id, title, description, starts_on, ends_on, active, created_at, updated_at
		FROM ai_campaigns WHERE id = $1
	`, id)
	return c, err
}

func (r *Repository) CreateCampaign(ctx context.Context, c entity.AICampaign) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO ai_campaigns (id, title, description, starts_on, ends_on, active, created_at, updated_at)
		VALUES (:id, :title, :description, :starts_on, :ends_on, :active, :created_at, :updated_at)
	`, c)
	return err
}

func (r *Repository) UpdateCampaign(ctx context.Context, c entity.AICampaign) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE ai_campaigns SET title = :title, description = :description, starts_on = :starts_on,
		                        ends_on = :ends_on, active = :active, updated_at = :updated_at
		WHERE id = :id
	`, c)
	return err
}
