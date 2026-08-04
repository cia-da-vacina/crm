CREATE TABLE phone_verifications (
    id              UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    phone_e164      TEXT NOT NULL,
    code_hash       TEXT NOT NULL,
    attempts        INT NOT NULL DEFAULT 0,
    resend_count    INT NOT NULL DEFAULT 0,
    expires_at      TIMESTAMPTZ NOT NULL,
    confirmed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Uma pendência ativa (não confirmada) por conversa de cada vez — resend
-- reusa a mesma linha em vez de criar outra.
CREATE UNIQUE INDEX phone_verifications_active_per_conversation_idx
    ON phone_verifications(conversation_id) WHERE confirmed_at IS NULL;
