CREATE TABLE customers (
    id             UUID PRIMARY KEY,
    display_name   TEXT NOT NULL DEFAULT '',
    identification TEXT NOT NULL DEFAULT 'anonymous' CHECK (identification IN ('anonymous', 'identified')),
    primary_phone  TEXT,
    unit_id        UUID REFERENCES units(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- primary_phone (E.164) é a chave de negócio pra merge cross-canal — único só
-- quando presente, pois clientes anônimos não têm telefone (docs/BACKEND-CONTRACT.md §3).
CREATE UNIQUE INDEX customers_primary_phone_unique_idx ON customers(primary_phone) WHERE primary_phone IS NOT NULL;

CREATE INDEX customers_unit_id_idx ON customers(unit_id);
