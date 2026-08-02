# CRM Cia da Vacina — Frontend

Frontend do CRM de atendimento multicanal (WhatsApp, Instagram e Facebook Messenger) da Cia da Vacina: inbox com triagem por IA + handoff humano, fila de engagements de rede social e pipeline comercial por unidade.

Next.js 15 (App Router) + React 19 + TypeScript. O BFF fala com o backend via `API_URL`.

## Como rodar

1. Suba o backend (API em `:8080`).
2. Configure o frontend:

```bash
cp .env.example .env.local
npm install
npm run dev
```

```env
API_URL=http://localhost:8080/api/v1
COOKIE_SECURE=false
```

Abra [http://localhost:3000](http://localhost:3000).

Se o seed do backend ainda cria os usuários de demonstração, use:

| Usuário | Senha |
|---|---|
| `admin@ciadavacina.com.br` | `admin123` |
| `atendente@ciadavacina.com.br` | `agent123` |

## Arquitetura (resumo)

- **BFF de autenticação**: `POST /api/auth/login` autentica no backend e grava a sessão em cookies **httpOnly** (`cv_access`/`cv_refresh`) — o browser nunca vê o token.
- **Proxy autenticado**: toda chamada de domínio do browser vai para `/api/proxy/{path}`, que injeta `Authorization: Bearer` e repassa para `{API_URL}/{path}`.
- **`middleware.ts`**: guarda de rota no Edge — redireciona para `/login` sem sessão, e para `/inbox` se já autenticado.
- **PWA**: instalável (manifest + service worker), com fallback offline.
- **Design system**: componentes/tokens/ícones vêm dos pacotes internos `@cia-da-vacina/*`.

Detalhes em [`../docs/FRONTEND-ARCHITECTURE.md`](../docs/FRONTEND-ARCHITECTURE.md).

## Notas de segurança

- Nenhuma variável sensível é `NEXT_PUBLIC_*`. `API_URL` e `COOKIE_SECURE` são lidos apenas no servidor (`src/server/env.ts`).
- Tokens de canal Meta e chaves de IA nunca trafegam para o cliente — a UI só recebe `token_masked`.
- Sessão de usuário vive exclusivamente em cookies `httpOnly`/`SameSite=Lax`.

## Scripts

| Script | Descrição |
|---|---|
| `npm run dev` | Sobe o servidor de desenvolvimento (`next dev -H 0.0.0.0`). |
| `npm run build` | Build de produção. |
| `npm run start` | Sobe o build de produção (`next start -H 0.0.0.0`). |
| `npm run lint` | ESLint (`next lint`). |
| `npm run typecheck` | Checagem de tipos sem emitir (`tsc --noEmit`). |

## Documentação

Índice em [`docs/README.md`](../docs/README.md).

- [`docs/BACKEND-CONTRACT.md`](../docs/BACKEND-CONTRACT.md) — contrato de API
- [`docs/PRODUCT-V2.md`](../docs/PRODUCT-V2.md) — produto e jornadas
- [`docs/FRONTEND-ARCHITECTURE.md`](../docs/FRONTEND-ARCHITECTURE.md) — arquitetura do frontend
- [`docs/APPROVED-SCOPE.md`](../docs/APPROVED-SCOPE.md) — escopo aprovado
