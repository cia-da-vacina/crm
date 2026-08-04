CREATE TABLE units (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    code       TEXT NOT NULL,
    timezone   TEXT NOT NULL DEFAULT 'America/Sao_Paulo',
    active     BOOLEAN NOT NULL DEFAULT true,
    address    TEXT NOT NULL,
    city       TEXT NOT NULL,
    district   TEXT,
    complement TEXT,
    reference  TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT units_code_unique UNIQUE (code)
);
