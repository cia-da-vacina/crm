CREATE TABLE pops (
    id          UUID PRIMARY KEY,
    title       TEXT NOT NULL,
    body        TEXT NOT NULL,
    -- JSONB em vez de TEXT[]: evita depender de scan de array do Postgres via
    -- pgx/sqlx (sem suporte direto a []string sem um tipo Scanner/Valuer
    -- customizado) — mesmo padrão que BoletoCosts usa no neohabit.
    intent_tags JSONB NOT NULL DEFAULT '[]',
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX pops_intent_tags_idx ON pops USING GIN (intent_tags);
