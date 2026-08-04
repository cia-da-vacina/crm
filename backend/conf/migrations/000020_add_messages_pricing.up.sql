-- Núcleo de custo (docs/WHATSAPP-2026-ADAPTATION-PLAN.md, Frente A). A Meta
-- reporta a categoria de cobrança real por mensagem enviada no webhook de
-- status ("pricing.category"/"pricing.billable"/"pricing.pricing_model" —
-- ver seção 2.2 do plano), não um valor monetário. pricing_category/
-- pricing_billable/pricing_model guardam esse reporte quando confirmado via
-- webhook; cost_brl é sempre calculado localmente (categoria x
-- message_pricing_rates, migration seguinte), nunca vindo da Meta.
--
-- pricing_confirmed distingue as duas origens possíveis do valor: false =
-- estimado no momento do envio (kind=text -> "service"; kind=template ->
-- categoria do catálogo, ver message_templates) a partir do client mock, já
-- que não existe client HTTP real nem webhook de status real ainda
-- (backend/ARCHITECTURE.md §5/§6); true = reconciliado com o webhook de
-- status real quando esse client existir. Nenhuma migration futura deveria
-- precisar mudar esse desenho — só o valor de pricing_confirmed passa a
-- virar true na prática.
ALTER TABLE messages ADD COLUMN pricing_category TEXT
    CHECK (pricing_category IN (
        'marketing', 'utility', 'authentication', 'authentication_international',
        'service', 'free_entry_point', 'free_customer_service'
    ));
ALTER TABLE messages ADD COLUMN pricing_billable BOOLEAN;
ALTER TABLE messages ADD COLUMN pricing_model TEXT;
ALTER TABLE messages ADD COLUMN pricing_confirmed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE messages ADD COLUMN cost_brl NUMERIC(12, 4);

CREATE INDEX messages_pricing_category_idx ON messages(pricing_category) WHERE pricing_category IS NOT NULL;
