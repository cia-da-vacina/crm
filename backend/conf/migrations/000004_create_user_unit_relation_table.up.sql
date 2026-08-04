CREATE TABLE user_unit_relation (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    unit_id    UUID NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_unit_relation_unique UNIQUE (user_id, unit_id)
);

CREATE INDEX user_unit_relation_unit_id_idx ON user_unit_relation(unit_id);
