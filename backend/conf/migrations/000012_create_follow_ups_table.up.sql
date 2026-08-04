CREATE TABLE follow_ups (
    id              UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    customer_id     UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    unit_id         UUID NOT NULL REFERENCES units(id),
    -- Snapshot do estágio no momento da criação — não muda se a conversa
    -- evoluir depois; reflete o motivo que gerou o follow-up.
    pipeline_stage  TEXT NOT NULL,
    due_at          TIMESTAMPTZ NOT NULL,
    status          TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'done', 'canceled')),
    note            TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX follow_ups_unit_id_idx ON follow_ups(unit_id);
CREATE INDEX follow_ups_conversation_id_idx ON follow_ups(conversation_id);
-- Uma pendência aberta por conversa de cada vez — evita spam de follow-up
-- duplicado se o agente ficar oscilando entre aguardando_fechamento/nao_fechado.
CREATE UNIQUE INDEX follow_ups_open_per_conversation_idx ON follow_ups(conversation_id) WHERE status = 'open';
