-- audit_logs é append-only por design (docs/BACKEND-CONTRACT.md §9: "ações
-- sensíveis... devem ser logadas de forma append-only") — nenhum código do
-- backend faz UPDATE/DELETE nela, só INSERT. actor_user_id fica nullable
-- porque um User pode ser removido depois (soft-delete via active:false na
-- prática, mas o schema não impede um DELETE físico futuro) sem que isso
-- apague o rastro da ação que ele tomou.
CREATE TABLE audit_logs (
    id            UUID PRIMARY KEY,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action        TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id   TEXT NOT NULL,
    unit_id       UUID REFERENCES units(id) ON DELETE SET NULL,
    metadata      JSONB NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX audit_logs_created_at_id_idx ON audit_logs(created_at DESC, id DESC);
CREATE INDEX audit_logs_actor_user_id_idx ON audit_logs(actor_user_id);
CREATE INDEX audit_logs_resource_idx ON audit_logs(resource_type, resource_id);
