-- Colunas que TriageSummary precisa persistir (fase 7 — sem isso não dá pra
-- responder GET /conversations/:id/triage sem re-chamar a IA a cada leitura,
-- o que o contrato não pede: "conteúdo 100% gerado pela IA... o frontend só
-- exibe", ou seja, é gerado quando a mensagem chega, não a cada GET).
ALTER TABLE conversations ADD COLUMN collected_fields JSONB NOT NULL DEFAULT '{}';
ALTER TABLE conversations ADD COLUMN triage_confidence DOUBLE PRECISION;
ALTER TABLE conversations ADD COLUMN triage_ready_for_handoff BOOLEAN NOT NULL DEFAULT false;
