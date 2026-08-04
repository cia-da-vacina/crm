CREATE TABLE customer_identities (
    id             UUID PRIMARY KEY,
    customer_id    UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    channel        TEXT NOT NULL CHECK (channel IN ('whatsapp', 'instagram', 'facebook')),
    external_id    TEXT NOT NULL,
    display_handle TEXT,
    phone_e164     TEXT,
    verified_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Chave natural pra resolver external_id -> Customer na ingestão de
    -- webhook (idempotência) — nunca duas identidades do mesmo canal com o
    -- mesmo id nativo (wa_id/IGSID/PSID) podem existir separadamente.
    CONSTRAINT customer_identities_channel_external_id_unique UNIQUE (channel, external_id)
);

CREATE INDEX customer_identities_customer_id_idx ON customer_identities(customer_id);
