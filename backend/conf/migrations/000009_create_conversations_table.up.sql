CREATE TABLE conversations (
    id                    UUID PRIMARY KEY,
    customer_id           UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    channel               TEXT NOT NULL CHECK (channel IN ('whatsapp', 'instagram', 'facebook')),
    channel_thread_id     TEXT,
    unit_id               UUID NOT NULL REFERENCES units(id),
    pipeline_stage        TEXT NOT NULL DEFAULT 'em_atendimento'
        CHECK (pipeline_stage IN ('em_atendimento', 'em_negociacao', 'aguardando_fechamento', 'fechado', 'nao_fechado')),
    mode                  TEXT NOT NULL DEFAULT 'ai_triage' CHECK (mode IN ('ai_triage', 'human')),
    owner_id              UUID REFERENCES users(id),
    intent                TEXT,
    ai_summary            TEXT,
    triage_notes          TEXT,
    phone_gate            TEXT NOT NULL DEFAULT 'not_needed'
        CHECK (phone_gate IN ('not_needed', 'required', 'pending_verification', 'collected')),
    loss_reason_code      TEXT REFERENCES loss_reasons(code),
    loss_reason_text      TEXT,
    window_expires_at     TIMESTAMPTZ,
    last_message_preview  TEXT NOT NULL DEFAULT '',
    last_message_at       TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX conversations_unit_id_idx ON conversations(unit_id);
CREATE INDEX conversations_customer_id_idx ON conversations(customer_id);
CREATE INDEX conversations_owner_id_idx ON conversations(owner_id);
-- Cursor da /inbox ordena por (last_message_at, id) DESC — índice composto
-- cobre a query sem sort extra.
CREATE INDEX conversations_last_message_at_id_idx ON conversations(last_message_at DESC, id DESC);
