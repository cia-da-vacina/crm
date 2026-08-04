-- Tabela de preço local usada pra converter pricing_category (o que a Meta
-- reporta ou o que o backend estima) num valor em BRL — a Meta não devolve
-- o valor monetário cobrado, só a categoria (ver migration 000020 e
-- docs/WHATSAPP-2026-ADAPTATION-PLAN.md §2.2). Os valores default abaixo são
-- as ESTIMATIVAS do rate card público mais recente da Meta usadas em
-- WhatsApp_API_Optimization_Guide.md (raiz do repo) — o próprio guia avisa
-- que os valores definitivos costumam sair só perto de 1º/set/2026, por
-- isso a tabela é editável via API (GET/PATCH /settings/pricing-rates,
-- admin-only) em vez de constante no código.
CREATE TABLE message_pricing_rates (
    category    TEXT PRIMARY KEY
        CHECK (category IN (
            'marketing', 'utility', 'authentication', 'authentication_international',
            'service', 'free_entry_point', 'free_customer_service'
        )),
    rate_brl    NUMERIC(10, 4) NOT NULL,
    billable    BOOLEAN NOT NULL DEFAULT true,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO message_pricing_rates (category, rate_brl, billable) VALUES
    ('marketing',                     0.3100, true),
    ('utility',                       0.0068, true),
    ('authentication',                0.0068, true),
    -- authentication_international: o guia menciona que "pode custar mais"
    -- sem dar valor — usa o mesmo da authentication doméstica até a Cia da
    -- Vacina confirmar a subcategoria configurada na conta (ver guia §1
    -- Regra 5). Ajustável sem migration via PATCH.
    ('authentication_international',  0.0068, true),
    ('service',                       0.0350, true),
    ('free_entry_point',              0.0000, false),
    ('free_customer_service',         0.0000, false);
