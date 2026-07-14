# Sprint 1 — Kickoff

**Objetivo:** Fundação Auth/Units + contratos + shell FE com mocks.

## Entregue

| ID | Item | Status |
|----|------|--------|
| T-001 | Setup Go module, health, logging | Feito |
| T-002 | Schema SQL users/units + seed 5 unidades | Feito (migrations; runtime Memory) |
| T-003 | Auth JWT + RBAC middleware | Feito |
| T-004 | CRUD users/units + vínculos | Feito |
| T-005 | App shell Next.js + login + guards | Feito |
| T-006 | Layout, nav, unit switcher, tokens | Feito |
| T-007 | OpenAPI v0 + MSW mocks | Feito |

## Próximo (Sprint 2)

- T-010…T-014 WhatsApp webhook + domain messages + inbox API
- T-015…T-016 Inbox UI completo + composer
- Trocar Memory store por Postgres quando RDS/local estiver pronto

## Como validar

1. `go test ./...` em `backend/`
2. `go run ./cmd/api` → `POST /api/v1/auth/login`
3. `npm run dev` em `frontend/` com mocks → login → inbox
