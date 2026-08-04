-- WhatsApp é 1 número por unidade (docs/decisions.md D-01) — unit_id
-- obrigatório. Instagram/Facebook são centralizados numa conta única da
-- marca pra todas as 5 unidades (confirmado com o stakeholder) — unit_id
-- NULL, uma linha global por canal.
CREATE TABLE meta_channel_configs (
    id               UUID PRIMARY KEY,
    channel          TEXT NOT NULL CHECK (channel IN ('whatsapp', 'instagram', 'facebook')),
    unit_id          UUID REFERENCES units(id) ON DELETE CASCADE,
    enabled          BOOLEAN NOT NULL DEFAULT false,
    account_id       TEXT NOT NULL DEFAULT '',
    display_name     TEXT NOT NULL DEFAULT '',
    phone_number_id  TEXT,
    webhook_verified BOOLEAN NOT NULL DEFAULT false,
    -- Token cifrado (AES-GCM, pkg/crypto) — nunca texto plano. token_masked é
    -- calculado uma vez no momento da rotação (a partir do plaintext, antes
    -- de descartá-lo) pra nunca precisar decifrar só pra exibir um preview.
    token_ciphertext BYTEA,
    token_masked     TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT meta_channel_configs_whatsapp_needs_unit CHECK (
        (channel = 'whatsapp' AND unit_id IS NOT NULL) OR
        (channel != 'whatsapp' AND unit_id IS NULL)
    )
);

-- WhatsApp: uma config por (channel, unit). IG/FB: uma config global por
-- channel (unit_id NULL em todas, então o índice parcial garante só 1 linha).
CREATE UNIQUE INDEX meta_channel_configs_whatsapp_unique_idx ON meta_channel_configs(channel, unit_id) WHERE channel = 'whatsapp';
CREATE UNIQUE INDEX meta_channel_configs_global_unique_idx ON meta_channel_configs(channel) WHERE channel != 'whatsapp';

-- Lookup do webhook: dado um phone_number_id que chegou no payload da Meta,
-- descobrir direto qual unit_id/config é dono (roteamento automático D-01).
CREATE UNIQUE INDEX meta_channel_configs_phone_number_id_idx ON meta_channel_configs(phone_number_id) WHERE phone_number_id IS NOT NULL;
