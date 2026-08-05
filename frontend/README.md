# CRM Cia da Vacina — Frontend

Frontend do CRM de atendimento multicanal (WhatsApp, Instagram e Facebook Messenger): inbox com triagem por IA + handoff humano, engagements, pipeline e agenda.

Next.js 15 (App Router) + React 19 + TypeScript. BFF com cookies httpOnly; dados via `USE_MOCKS` (fixtures) ou `API_URL` (backend real).

## Docker (raiz do monorepo)

```bash
cp .env.docker.example .env.docker
docker compose --env-file .env.docker up --build -d
```

App: http://localhost:3000  
Login: `admin@ciadavacina.com.br` / `admin123` (mocks)

## Local

```bash
cp .env.example .env.local   # USE_MOCKS=true
npm install                  # na raiz do monorepo
npm run dev
```

| Usuário | Senha |
|---|---|
| `admin@ciadavacina.com.br` | `admin123` |
| `atendente@ciadavacina.com.br` | `agent123` |

Para API real: `USE_MOCKS=false` + `API_URL=http://localhost:8080/api/v1`.

## Arquitetura (resumo)

- **BFF auth**: `POST /api/auth/login` → cookies httpOnly
- **Proxy**: `/api/proxy/{path}` → mock in-process ou `{API_URL}/{path}`
- **PWA** + design system workspace (`packages/*`)

Detalhes: [`docs/FRONTEND-ARCHITECTURE.md`](../docs/FRONTEND-ARCHITECTURE.md).

## Scripts

| Script | Descrição |
|---|---|
| `npm run dev` | Dev server |
| `npm run build` | Build produção |
| `npm run start` | Serve build |
| `npm run lint` | ESLint |
| `npm run typecheck` | `tsc --noEmit` |
