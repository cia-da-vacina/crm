# CRM Cia da Vacina

CRM de atendimento multicanal (WhatsApp, Instagram e Facebook Messenger) com triagem por IA + handoff humano — frontend Next.js + backend Go.

O frontend fala com o backend via BFF (cookies httpOnly + proxy `/api/proxy/*`). Ver [`docs/FRONTEND-ARCHITECTURE.md`](./docs/FRONTEND-ARCHITECTURE.md) e [`docs/BACKEND-CONTRACT.md`](./docs/BACKEND-CONTRACT.md).

## Repositório

https://github.com/cia-da-vacina/crm

## Docker (recomendado)

Requer [Docker Desktop](https://www.docker.com/products/docker-desktop/) (ou Engine + Compose v2).

```bash
cp .env.docker.example .env.docker
docker compose --env-file .env.docker up --build -d
# ou: npm run docker:up
```

- App: http://localhost:3000
- Por padrão `USE_MOCKS=true` — UI completa **sem** API Go
- Login demo: `admin@ciadavacina.com.br` / `admin123`
- Parar: `docker compose down` / `npm run docker:down`

Quando a API do Felipe estiver pronta, no `.env.docker`:

```env
USE_MOCKS=false
API_URL=http://backend:8080/api/v1
```

## Como rodar o frontend (sem Docker)

```bash
cp frontend/.env.example frontend/.env.local   # USE_MOCKS=true
npm install
npm run dev
```

## Documentação

Índice em [`docs/README.md`](./docs/README.md).
