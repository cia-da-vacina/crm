# CRM Cia da Vacina

CRM de atendimento multicanal (WhatsApp, Instagram e Facebook Messenger) com triagem por IA + handoff humano — frontend Next.js de produção (sem mocks) + backend Go.

O frontend **não inclui mais mocks (MSW)**: todo dado vem de um backend real através de um BFF de autenticação (cookies httpOnly, sem token exposto ao browser) e de um proxy autenticado (`/api/proxy/*`). Ver [`docs/FRONTEND-ARCHITECTURE.md`](./docs/FRONTEND-ARCHITECTURE.md) e [`docs/BACKEND-CONTRACT.md`](./docs/BACKEND-CONTRACT.md).

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
# defina API_URL apontando para o backend acima
npm install
npm run dev
```

Abra http://localhost:3000. É necessário o backend rodando com um seed de usuários — se o seed do backend ainda cria os usuários de demonstração (`admin@ciadavacina.com.br` / `admin123`, `atendente@ciadavacina.com.br` / `agent123`), essas credenciais continuam válidas em ambiente local; não são mais aplicáveis a uma "demo online" com mocks, pois o modo mock foi removido do frontend.

## Documentação

- [`docs/APPROVED-SCOPE.md`](./docs/APPROVED-SCOPE.md) — escopo aprovado e histórico de mudanças.
- [`docs/PRODUCT-V2.md`](./docs/PRODUCT-V2.md) — visão de produto, personas e jornadas.
- [`docs/BACKEND-CONTRACT.md`](./docs/BACKEND-CONTRACT.md) — contrato de API para o backend.
- [`docs/FRONTEND-ARCHITECTURE.md`](./docs/FRONTEND-ARCHITECTURE.md) — arquitetura do frontend.
- [`docs/openapi.yaml`](./docs/openapi.yaml) — especificação OpenAPI (Auth/Users/Units/Me).
- [`docs/spec.md`](./docs/spec.md) — especificação técnica original.
