# CRM Cia da Vacina

CRM de atendimento multicanal (WhatsApp, Instagram e Facebook Messenger) com triagem por IA + handoff humano — frontend Next.js + backend Go.

O frontend fala com o backend via BFF (cookies httpOnly + proxy `/api/proxy/*`). Ver [`docs/FRONTEND-ARCHITECTURE.md`](./docs/FRONTEND-ARCHITECTURE.md) e [`docs/BACKEND-CONTRACT.md`](./docs/BACKEND-CONTRACT.md).

## Repositório

https://github.com/cia-da-vacina/crm

## Status do backend

A pasta [`backend/`](./backend/) está **sem implementação** de propósito — só um README apontando para o contrato. O Felipe implementa a API Go a partir de [`docs/BACKEND-CONTRACT.md`](./docs/BACKEND-CONTRACT.md) e [`docs/openapi.yaml`](./docs/openapi.yaml).

## Docker (recomendado)

Requer [Docker Desktop](https://www.docker.com/products/docker-desktop/) (ou Engine + Compose v2).

```bash
cp .env.docker.example .env.docker
docker compose --env-file .env.docker up --build
# ou: npm run docker:up   (depois de criar .env.docker)
```

- App: http://localhost:3000
- `API_URL` no `.env.docker` aponta para a API (default: `http://host.docker.internal:8080/api/v1` — API no host).
- Sem backend no ar, a UI sobe; login/dados falham até a API existir.
- Parar: `docker compose down` / `npm run docker:down`

Imagem: `frontend/Dockerfile` (multi-stage, Next.js `output: "standalone"`, contexto = raiz do monorepo por causa dos workspaces do design system).

## Como rodar o frontend (sem Docker)

```bash
# na raiz do monorepo (workspaces)
cp frontend/.env.example frontend/.env.local
# API_URL=http://localhost:8080/api/v1  (quando a API existir)
npm install
npm run dev
```

Abra http://localhost:3000.

## Documentação

Índice em [`docs/README.md`](./docs/README.md).

- [`docs/APPROVED-SCOPE.md`](./docs/APPROVED-SCOPE.md) — escopo aprovado
- [`docs/PRODUCT-V2.md`](./docs/PRODUCT-V2.md) — produto e jornadas
- [`docs/BACKEND-CONTRACT.md`](./docs/BACKEND-CONTRACT.md) — contrato de API
- [`docs/FRONTEND-ARCHITECTURE.md`](./docs/FRONTEND-ARCHITECTURE.md) — arquitetura do frontend
- [`docs/openapi.yaml`](./docs/openapi.yaml) — OpenAPI (Auth/Users/Units/Me)
- [`docs/decisions.md`](./docs/decisions.md) — defaults operacionais
- [`docs/marketing/`](./docs/marketing/) — deck, proposta e materiais comerciais
