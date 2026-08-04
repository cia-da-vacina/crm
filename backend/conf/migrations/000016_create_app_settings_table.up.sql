-- Singleton (id fixo em 1, CHECK garante que nunca existe uma segunda linha)
-- — configurações globais de IA/triagem + a unidade padrão pra onde cai
-- conversa nova de canal centralizado (Instagram/Facebook — ver
-- meta_channel_configs). default_unit_id começa NULL: até o admin
-- configurar, o webhook de IG/FB não tem pra onde rotear (ver módulo webhook).
CREATE TABLE app_settings (
    id                     INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    ai_enabled             BOOLEAN NOT NULL DEFAULT false,
    ai_system_prompt       TEXT NOT NULL DEFAULT '',
    ai_context             TEXT NOT NULL DEFAULT '',
    triage_enabled         BOOLEAN NOT NULL DEFAULT true,
    triage_handoff_intents JSONB NOT NULL DEFAULT '[]',
    default_unit_id        UUID REFERENCES units(id),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO app_settings (id) VALUES (1);
