-- Catálogo próprio de templates WhatsApp (docs/WHATSAPP-2026-ADAPTATION-PLAN.md,
-- Frente B — decisão D5: backend é fonte de verdade, não só um texto livre
-- em messages.template_name). variable_count valida a contagem de {{n}}
-- antes do envio evitar rejeição pela Meta (guia §4, checklist de aprovação);
-- approval_status reflete o status de aprovação no Meta Business Manager —
-- setado manualmente por enquanto (não há client HTTP real ainda pra
-- consultar isso via API, ver ARCHITECTURE.md §5).
CREATE TABLE message_templates (
    id               UUID PRIMARY KEY,
    name             TEXT NOT NULL,
    category         TEXT NOT NULL CHECK (category IN ('marketing', 'utility', 'authentication')),
    language_code    TEXT NOT NULL DEFAULT 'pt_BR',
    body             TEXT NOT NULL,
    variable_count   INT NOT NULL DEFAULT 0 CHECK (variable_count >= 0),
    approval_status  TEXT NOT NULL DEFAULT 'pending'
        CHECK (approval_status IN ('pending', 'approved', 'rejected')),
    active           BOOLEAN NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Um nome de template é único por idioma na conta Meta (é como a Cloud API
-- referencia o template no envio) — mesmo template em dois idiomas são duas
-- linhas.
CREATE UNIQUE INDEX message_templates_name_language_unique_idx ON message_templates(name, language_code);
CREATE INDEX message_templates_category_idx ON message_templates(category);
