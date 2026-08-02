# CRM Cia da Vacina

CRM de atendimento multicanal (WhatsApp, Instagram e Facebook Messenger) com triagem por IA + handoff humano — frontend Next.js + backend Go.

O frontend fala com o backend via BFF (cookies httpOnly + proxy `/api/proxy/*`). Ver [`docs/FRONTEND-ARCHITECTURE.md`](./docs/FRONTEND-ARCHITECTURE.md) e [`docs/BACKEND-CONTRACT.md`](./docs/BACKEND-CONTRACT.md).

## Repositório

https://github.com/cia-da-vacina/crm

## Como rodar localmente

```bash
# Terminal 1 — Backend
cd backend
go run ./cmd/api

# Terminal 2 — Frontend
cd frontend
cp .env.example .env.local
# API_URL=http://localhost:8080/api/v1
npm install
npm run dev
```

Abra http://localhost:3000. Com o seed do backend, use `admin@ciadavacina.com.br` / `admin123` ou `atendente@ciadavacina.com.br` / `agent123`.

## Documentação

Índice em [`docs/README.md`](./docs/README.md).

- [`docs/APPROVED-SCOPE.md`](./docs/APPROVED-SCOPE.md) — escopo aprovado
- [`docs/PRODUCT-V2.md`](./docs/PRODUCT-V2.md) — produto e jornadas
- [`docs/BACKEND-CONTRACT.md`](./docs/BACKEND-CONTRACT.md) — contrato de API
- [`docs/FRONTEND-ARCHITECTURE.md`](./docs/FRONTEND-ARCHITECTURE.md) — arquitetura do frontend
- [`docs/openapi.yaml`](./docs/openapi.yaml) — OpenAPI (Auth/Users/Units/Me)
- [`docs/decisions.md`](./docs/decisions.md) — defaults operacionais
- [`docs/marketing/`](./docs/marketing/) — deck, proposta e materiais comerciais
