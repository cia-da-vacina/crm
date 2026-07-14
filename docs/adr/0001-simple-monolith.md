# ADR 0001 — Monólito modular Go + Next.js + PostgreSQL

## Status

Aceito — 2026-07-14

## Contexto

Equipe de **dois desenvolvedores** (Felipe backend, Cristian frontend). Produto: CRM de atendimento WhatsApp com IA de triagem e pipeline comercial para 5 unidades.

Alternativas avaliadas: microserviços, BFF separado, Redis/Kafka desde o dia 1.

## Decisão

1. **Backend:** um único binário Go (monólito modular por pacotes de domínio).
2. **Frontend:** Next.js (App Router) consumindo a API Go via REST + SSE.
3. **Banco:** PostgreSQL como fonte da verdade.
4. **Sem Redis/Kafka no MVP:** retries e follow-ups via tabela `jobs` + goroutines no mesmo processo.
5. **Tempo real:** SSE (não WebSocket) para reduzir complexidade operacional.
6. **Deploy:** API + web separados em processo, um Postgres gerenciado.

## Consequências

### Positivas
- Menos ops, menos falhas de rede internas, entrega mais rápida.
- Transações e consistência mais simples (Postgres).
- Onboarding e debug fáceis para time pequeno.

### Negativas
- Escala vertical primeiro; extrair serviços só se um módulo saturar (ex.: WhatsApp webhook).
- SSE exige cuidado com proxies/timeouts (heartbeat).

## Não faremos no MVP

- Multi-tenant SaaS genérico
- Message broker
- Cache distribuído obrigatório
- WebSocket próprio
