CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    ip         TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT refresh_tokens_token_hash_unique UNIQUE (token_hash)
);

-- Índice parcial: só sessões ainda ativas importam pra lookup de login/refresh/logout.
CREATE INDEX refresh_tokens_user_id_active_idx ON refresh_tokens(user_id) WHERE revoked_at IS NULL;
