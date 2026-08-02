# Arquitetura do Frontend — CRM Cia da Vacina

**Revisão:** 2026-08-02
**Stack:** Next.js 15 (App Router) · React 19 · TypeScript · styled-components · TanStack Query · design system próprio Cia da Vacina.

Documento curto de orientação para quem for mexer no frontend. Para o contrato de API consumido, ver [`BACKEND-CONTRACT.md`](./BACKEND-CONTRACT.md).

---

## Estrutura de pastas

```
frontend/src/
├── app/                      # App Router — rotas e route handlers
│   ├── (auth)/               # Grupo de rotas autenticadas (inbox, dashboard, users...)
│   ├── (unauth)/             # Grupo de rotas públicas (login)
│   ├── api/
│   │   ├── auth/             # BFF: login, logout, refresh, session, unit
│   │   └── proxy/[...path]/  # BFF genérico: repassa tudo para {API_URL}/{path}
│   ├── ~offline/             # Fallback de PWA offline
│   ├── layout.tsx            # Layout raiz (providers globais)
│   └── manifest.ts           # Web App Manifest (PWA)
├── domain/                   # Tipos de domínio puros (sem dependência de framework)
├── services/                 # Camada de acesso a dados — um arquivo por recurso
├── server/                   # Código server-only (env, cookies, fetch ao backend)
├── providers/                # Contextos React (auth, query, tema, navegação)
├── components/                # Componentes de app (nav, guards, modais)
├── lib/                      # Utilitários (constantes, erros, motion)
└── middleware.ts             # Guarda de rotas (Edge middleware)
```

Regra geral: **domain não importa de service, service não importa de components, server nunca é importado por código client**. `frontend/src/domain/index.ts` é o barrel público de tipos.

---

## Auth via BFF + cookies httpOnly

O frontend nunca guarda token de sessão em memória JS, `localStorage` ou `sessionStorage`. O fluxo:

1. `POST /api/auth/login` (route handler Next.js) chama o backend, recebe `access_token`/`refresh_token`, e os grava como cookies **httpOnly** (`cv_access`, `cv_refresh` — ver `server/cookies.ts`). Responde ao browser apenas `{ user: MeResponse }`.
2. Toda chamada de domínio do browser vai para `/api/proxy/{path}` (nunca direto ao backend). O route handler lê o cookie `cv_access`, injeta `Authorization: Bearer` e repassa ao backend (`server/backend.ts` / rota de proxy).
3. Em caso de `401` do backend, o proxy tenta uma renovação silenciosa via `/auth/refresh` (`server/refresh-session.ts`) e repete a chamada original uma vez.
4. `GET /api/auth/session` é a fonte de verdade de "quem está logado" para o app (`AuthProvider` em `providers/auth-provider.tsx`) — chamada no bootstrap e após login/logout.
5. `middleware.ts` faz o guard de rota no Edge: se não há cookie de sessão e a rota é protegida (`/inbox`, `/dashboard`, `/follow-ups`, `/pops`, `/users`, `/settings`, `/engagements`, `/customers`, `/campaigns`, `/units`), redireciona para `/login`; se já há sessão e o usuário tenta `/login`, redireciona para `/inbox`.
6. A unidade ativa (`cv_unit`) é outro cookie httpOnly, mas hoje é **preferência de cliente pura** — não há chamada ao backend para persistir a unidade ativa entre dispositivos (ver nota em `BACKEND-CONTRACT.md`).

---

## Camada de services

`frontend/src/services/*.ts` — um módulo por recurso (`auth`, `inbox`, `conversations`, `customers`, `engagements`, `followups`, `pops`, `users`, `units`, `dashboard`, `meta`, `loss-reasons`, `campaigns`), todos reexportados por `services/index.ts`.

- `services/http.ts` expõe `bffFetch`/`bffGet`/`bffPost`/`bffPut`/`bffPatch`/`bffDelete`: wrappers de `fetch` que **exigem** que o path comece com `/api/` (erro em runtime caso contrário) e sempre enviam `credentials: "same-origin"`.
- Nenhum service conhece a URL do backend — todos chamam `/api/proxy/...` ou `/api/auth/...`.
- Erros de rede/HTTP viram `ApiError` (`lib/errors.ts`), com `status`, `code` e `message` — consumidos pelas telas via `useMutation`/`useQuery` do TanStack Query.
- Paginação: `Paginated<T>` (offset/página) para listas de referência (units, users, customers); `CursorPage<T>` (cursor) para feeds de alto volume/tempo real (inbox, mensagens, engagements, follow-ups).

---

## Layouts auth/unauth

App Router usa dois grupos de rota:

- `(auth)/layout.tsx`: envolve as páginas autenticadas com `AuthGuard` (bloqueia render e redireciona para `/login` se não houver usuário — client-side, complementar ao `middleware.ts` no Edge) e `AppNav` (navegação/shell principal, seletor de unidade, menu de usuário).
- `(unauth)/layout.tsx`: shell mínimo, sem chrome de navegação — usado por `/login`.

Dupla camada de proteção (`middleware.ts` no Edge + `AuthGuard` no client) evita flash de conteúdo autenticado e cobre navegação client-side (`next/link`) que o middleware sozinho já intercepta, mas o guard reforça durante hidratação.

---

## PWA

Configurado via `@ducanh2912/next-pwa` em `next.config.ts`:
- Service worker gerado em build (`disable: true` em desenvolvimento).
- `fallbacks.document` aponta para `/~offline` quando não há rede.
- `app/manifest.ts` gera o Web App Manifest (ícones em `public/icons/`, `display: "standalone"`, cores de tema da marca, `lang: "pt-BR"`).
- `components/install-prompt.tsx` captura o evento `beforeinstallprompt` do navegador e mostra um banner de instalação customizado, com dispensa persistida em `sessionStorage` (chave `pwa-install-dismissed`, ver `lib/constants.ts`).

---

## Design system (pacotes)

Todo o UI visual vem dos pacotes do monorepo (`packages/*`), referenciados no `frontend/package.json` via workspace (`"*"`):

- `@cia-da-vacina/design-system` — componentes (`Button`, `Flex`, `Stack`, `Badge`, `AppShell`, `PageHeader`, `DataList`, `Toast`, etc.).
- `@cia-da-vacina/design-system-tokens` — tokens de tema claro/escuro (cores, espaçamento, radii, sombras) e `GlobalStyle`.
- `@cia-da-vacina/icon-system` — ícones (`BotIcon`, `HandshakeIcon`, `SendIcon`, etc.).
- `@cia-da-vacina/styled-system` — utilitários de styled-components/tema.

Todos declarados em `transpilePackages` no `next.config.ts` e o `styled-components` é forçado a um único caminho de resolução via alias de webpack, para evitar duas instâncias de `ThemeProvider` coexistindo entre o app e os pacotes. Tema claro/escuro: `ThemeModeProvider` (`providers/theme-provider.tsx`) + `CiaThemeProvider`.

---

## Variáveis de ambiente (server-only)

Nenhuma variável sensível é `NEXT_PUBLIC_*` — a única coisa exposta ao browser é o próprio bundle da UI, nunca URL de backend ou segredos.

| Variável | Onde é lida | Descrição |
|---|---|---|
| `API_URL` | `server/env.ts` | Base URL do backend REST, ex. `https://api.ciadavacina.com.br/api/v1`. Obrigatória em produção (o processo falha ao subir sem ela). Default de desenvolvimento: `http://localhost:8080/api/v1`. |
| `COOKIE_SECURE` | `server/env.ts` | `true`/`false` — se os cookies de sessão são marcados `Secure` (exige HTTPS). Default: `true` em produção, `false` fora dela. |

Ver `frontend/.env.example` para o template mínimo de desenvolvimento local. O frontend sempre fala com um backend real via BFF (`API_URL` + cookies httpOnly); não há modo mock.
