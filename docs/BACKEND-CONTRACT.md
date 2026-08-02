# Contrato de API — Backend CRM Cia da Vacina

**Status:** contrato de referência para implementação do backend (Go).
**Revisão:** 2026-08-02 — reflete o frontend de produção (sem mocks), multicanal (WhatsApp + Instagram + Facebook Messenger).
**Fonte da verdade complementar:** [`openapi.yaml`](./openapi.yaml) (Auth/Users/Units/Me formalizados em OpenAPI). Este documento cobre o restante do domínio (Inbox, Conversas, Engagements, Follow-ups, POPs, Dashboard, Meta settings, Webhooks) na mesma convenção, servindo de checklist funcional para o time de backend. Não substitui o OpenAPI — os dois devem ser mantidos em paralelo.

Todos os paths abaixo são **relativos a `/api/v1`** (`API_URL` no backend, ver [`FRONTEND-ARCHITECTURE.md`](./FRONTEND-ARCHITECTURE.md)). Os tipos citados (`Customer`, `ConversationDetail`, `Message` etc.) são os definidos em `frontend/src/domain/*.ts` — use-os como referência conceitual de shape, não como schema literal de banco.

---

## Índice

0. [Mapeamento de proxy](#0-mapeamento-de-proxy)
1. [Auth](#1-auth)
2. [Units / Users](#2-units--users)
3. [Customers (CRM id + identidades por canal)](#3-customers-crm-id--identidades-por-canal)
4. [Inbox / Conversations / Messages / Claim / Pipeline / Triage](#4-inbox--conversations--messages--claim--pipeline--triage)
5. [Engagements (story_reply, story_mention, post_comment, live_comment)](#5-engagements-story_reply-story_mention-post_comment-live_comment)
6. [Follow-ups, POPs, Loss reasons, Dashboard](#6-follow-ups-pops-loss-reasons-dashboard)
7. [Meta settings (multicanal)](#7-meta-settings-multicanal)
8. [Webhooks Meta (backend-only)](#8-webhooks-meta-backend-only)
9. [Requisitos de segurança](#9-requisitos-de-segurança)
10. [O que o frontend nunca processa](#10-o-que-o-frontend-nunca-processa)

---

## 0. Mapeamento de proxy

O browser **nunca** conhece `API_URL` nem envia tokens Meta/JWT diretamente. Todo tráfego de domínio passa por um único route handler Next.js que atua como proxy autenticado (BFF):

```
Browser → GET/POST/PUT/PATCH/DELETE /api/proxy/{path}
        → Next.js route handler (server-side, injeta Authorization: Bearer <access_token> do cookie httpOnly)
        → Backend GET/POST/PUT/PATCH/DELETE {API_URL}/{path}
```

- `{path}` é repassado **verbatim**, incluindo query string (`?unit_id=...`).
- Se o backend responder `401`, o proxy tenta **uma única vez** renovar a sessão via `POST /auth/refresh` (usando o refresh token do cookie) e repete a chamada original com o novo access token. Se o refresh falhar, o `401` original é repassado ao browser.
- Corpo e status HTTP do backend são espelhados 1:1 para o browser (exceto headers sensíveis).
- Rotas de autenticação (`/api/auth/login`, `/api/auth/logout`, `/api/auth/refresh`, `/api/auth/session`, `/api/auth/unit`) **não passam pelo proxy genérico** — são route handlers dedicados que manipulam cookies httpOnly diretamente (ver seção 1).

Exemplo: o service de frontend chama `GET /api/proxy/inbox?unit_id=u1`, que o proxy traduz para `GET {API_URL}/inbox?unit_id=u1` no backend.

---

## 1. Auth

Sessão gerenciada via BFF: o backend emite JWT access/refresh, mas o browser **nunca** os vê — eles ficam em cookies `httpOnly` setados pelos route handlers Next.js (`cv_access`, `cv_refresh`).

### `POST /auth/login`
- **Auth:** nenhuma (público).
- **Request:** `{ email: string, password: string }`
- **Response `200`:** `{ access_token, refresh_token, expires_in, user: User }` — este shape (`BackendLoginResponse`) só é consumido no servidor Next.js; o browser recebe de volta apenas `{ user: MeResponse }`.
- **Erros:** `401` credenciais inválidas; `403` usuário inativo (`active: false`).
- **Notas:** logo após o login, o BFF chama `GET /me` com o token recém-emitido para obter `units` (não presente no payload de login) antes de responder ao browser.

### `POST /auth/logout`
- **Auth:** `Bearer <access_token>`.
- **Request:** vazio.
- **Response:** `204`.
- **Notas:** deve invalidar/revogar o refresh token no backend (denylist ou versão de sessão). O BFF chama este endpoint best-effort — cookies são limpos no browser independentemente do resultado.

### `POST /auth/refresh`
- **Auth:** nenhuma (o refresh token vai no corpo, não como Bearer).
- **Request:** `{ refresh_token: string }`
- **Response `200`:** mesmo shape de `BackendLoginResponse` (novo par access/refresh — rotação obrigatória do refresh token).
- **Erros:** `401` refresh token expirado/revogado → BFF trata como sessão expirada e força novo login.

### `GET /me`
- **Auth:** `Bearer <access_token>`.
- **Response `200`:** `MeResponse` = `User` (`id, email, name, role, active, unit_ids?`) + `units: Unit[]` (apenas as unidades que o usuário pode operar).
- **Notas:** é a fonte de verdade de sessão para o frontend — chamado pelo BFF em `GET /api/auth/session` a cada bootstrap de app e após refresh.

**Endpoint de conveniência do BFF (não existe no backend):** `POST /api/auth/unit` apenas grava a unidade ativa em um cookie httpOnly (`cv_unit`) no Next.js — **hoje não há chamada ao backend para persistir isso**. Se o time de backend quiser persistir a última unidade ativa por usuário (ex.: multi-dispositivo), será necessário expor um endpoint novo (`PUT /me/active-unit` ou similar) — atualmente é apenas uma preferência de cliente.

---

## 2. Units / Users

### `GET /units`
- **Auth:** Bearer.
- **Response:** `Paginated<Unit>` — `Unit { id, name, code, timezone, active, address, city, district?, complement?, reference? }`.
- **Notas:** admin recebe todas; demais papéis só as unidades em `unit_ids`. Campos de endereço alimentam a tela de gestão e o contexto operacional.

### `GET /units/:id`
- **Auth:** Bearer.
- **Response:** `Unit`.

### `POST /units`
- **Auth:** Bearer (admin).
- **Request:** `CreateUnitPayload { name, code, city, address, timezone?, active?, district?, complement?, reference? }`.
- **Response:** `201 Unit`.

### `PATCH /units/:id`
- **Auth:** Bearer (admin).
- **Request:** `UpdateUnitPayload` (parcial de `CreateUnitPayload`).
- **Response:** `Unit`.
- **Notas:** a UI de **Unidades** (`/units`) é a superfície principal de CRUD e de visualização de quem tem acesso ao CRM de cada unidade (via `User.unit_ids` / `PUT /users/:id/units`).

### `GET /users`
- **Auth:** Bearer (admin/manager).
- **Response:** `Paginated<User>`.

### `GET /users/:id`
- **Auth:** Bearer.
- **Response:** `User`.

### `POST /users`
- **Auth:** Bearer (admin).
- **Request:** `{ email, password, name, role: UserRole, unit_ids: string[] }`
- **Response `201`:** `User`.
- **Notas:** `role` ∈ `admin | manager | supervisor | agent`. Backend deve hashear a senha (bcrypt/argon2) e nunca retorná-la.

### `PATCH /users/:id`
- **Auth:** Bearer (admin, ou o próprio usuário para campos limitados).
- **Request:** `{ name?, role?, active?, password? }`
- **Response:** `User`.

### `DELETE /users/:id`
- **Auth:** Bearer (admin).
- **Response:** `204`. Preferir soft-delete (`active: false`) a exclusão física, por auditoria.

### `PUT /users/:id/units`
- **Auth:** Bearer (admin).
- **Request:** `{ unit_ids: string[] }`
- **Response:** `204`.
- **Notas:** substitui integralmente o vínculo usuário×unidade (não é incremental).

---

## 3. Customers (CRM id + telefone como chave + identidades por canal)

**Decisão de arquitetura crítica:** a Meta **não fornece um identificador unificado entre canais**. Cada canal tem seu próprio identificador nativo:

| Canal | Identificador nativo | Telefone na entrada |
|---|---|---|
| WhatsApp | `wa_id` | **Sim** — Meta entrega o número (E.164). Cliente nasce `identified`. |
| Instagram | `ig_scoped_id` (IGSID) | **Não** — cliente nasce `anonymous`. |
| Facebook Messenger | `fb_psid` (PSID) | **Não** — cliente nasce `anonymous`. |

### Modelo

- `Customer.id` (UUID) — entidade canônica no CRM.
- `Customer.primary_phone` (E.164, **único quando presente**) — **chave de negócio** para merge cross-canal.
- `Customer.identification`: `anonymous` | `identified`.
- `CustomerIdentity` — vínculo N:1 entre `external_id` do canal e um `Customer`.

### Parede de privacidade (backend deve enforcing)

| `anonymous` (sem telefone) | `identified` (com `primary_phone`) |
|---|---|
| Dúvidas, preços, campanhas, chat leve | Histórico unificado entre canais |
| Sem acesso a ficha/histórico cross-canal | Agendamento e dados cadastrais |
| Sem merge automático | Follow-ups ligados à ficha |

O frontend **não decide** o que está atrás da parede — só respeita flags (`identification`, `phone_gate`) retornadas pela API. Endpoints que expõem dados gated devem retornar `403` com `code: "phone_required"` se o cliente ainda for anônimo.

### Fluxo de telefone por canal

1. **WhatsApp (webhook):** criar/atualizar `Customer` já `identified` com `primary_phone` = número Meta; criar `CustomerIdentity` whatsapp. **Nunca pedir telefone nem OTP** — a própria entrega pela Meta já prova posse do número naquele canal.
2. **Instagram / Messenger (webhook):** criar `Customer` `anonymous` (`primary_phone: null`) + `CustomerIdentity` do canal. Conversas leves seguem sem telefone.
3. **Quando a triagem/intenção exige cadastro** (`phone_gate: "required"`):
   1. A IA pede o número no canal atual (IG/FB).
   2. Backend valida formato E.164 e inicia verificação — **ainda não** seta `primary_phone`, **ainda não** promove para `identified`, **ainda não** faz merge.
   3. `phone_gate → pending_verification`; envia **OTP / confirmação via WhatsApp** para o número informado (template de autenticação Meta aprovado).
   4. Só após o cliente confirmar no WhatsApp (código ou resposta válida): seta `primary_phone`, promove `identified`, faz **merge** se já existir Customer canônico com esse telefone, e `phone_gate → collected`.

**Por que OTP no WhatsApp:** sem prova de posse, qualquer pessoa no Instagram poderia digitar o telefone de outra e atravessar a parede (histórico, agendamento, merge). O WhatsApp é o fator de posse do número.

### `phone_gate` (por conversa)

Valores: `not_needed` | `required` | `pending_verification` | `collected`.

| Estado | Significado |
|---|---|
| `not_needed` | Intenção leve — não perguntar |
| `required` | Precisa de telefone; ainda não informado |
| `pending_verification` | Número recebido; aguardando confirmação no WhatsApp |
| `collected` | Telefone confirmado (ou veio da Meta no WA) |

- Definido **somente no backend** (regras de intenção × canal × estado do Customer).
- Exposto em `ConversationSummary` e `TriageSummary` (`pending_phone_masked` quando aplicável).
- Exemplos: `precos`/`duvidas` → `not_needed`; `agendar` em IG → `required` → `pending_verification` → `collected`; WA → `collected`/`not_needed`.

### Verificação WhatsApp (2FA de posse)

- Usar **template de autenticação** (OTP) da WhatsApp Cloud API — mensagem business-initiated permitida para esse fim após aprovação do template.
- Código de curta validade (ex.: 5–10 min), com limite de reenvios e lockout básico.
- Confirmação pode ser: (a) usuário informa o código de volta no IG/FB / endpoint de confirm; ou (b) backend correlaciona webhook WhatsApp inbound do mesmo número respondendo ao template (preferência de produto: código digitado no fluxo atual é mais simples de UX cross-canal).
- **Falhas:** número inválido / sem WhatsApp → voltar `phone_gate` para `required` com mensagem clara; não deixar órfão em `pending_verification` indefinidamente (TTL).
- Enquanto `pending_verification` ou `required`, recursos gated continuam bloqueados (`403 phone_required` / `403 phone_verification_pending`).

### `GET /customers`
- **Auth:** Bearer.
- **Query:** `q?` (busca por nome/telefone), `unit_id?`, `identification?`.
- **Response:** `Paginated<Customer>` — `Customer { id, display_name, identification, primary_phone, unit_id?, identities: CustomerIdentity[], created_at, updated_at }`.
- **Notas:** listagens padrão de “clientes do CRM” podem filtrar só `identified`; anônimos existem para threads em aberto mas não são ficha comercial completa.

### `GET /customers/:id`
- **Auth:** Bearer.
- **Response:** `Customer` (com `identities` embutido).
- **Erros:** `403 phone_required` se a política da role/tela exigir identificado e o recurso for anônimo (opcional por endpoint).

### `GET /customers/:id/identities`
- **Auth:** Bearer.
- **Response:** `CustomerIdentity[]` — `{ id, customer_id, channel, external_id, display_handle?, phone_e164?, verified_at?, created_at }`.

### `POST /conversations/:id/phone` (iniciar verificação)
- **Auth:** Bearer (agente) **ou** uso interno pela IA de triagem no backend (mesmo shape).
- **Request:** `{ phone_e164: string }` (E.164, ex. `+5511999998888`).
- **Response `202`:** `ConversationDetail` com `phone_gate: "pending_verification"`, `pending_phone_masked`, `identification` ainda `"anonymous"`.
- **Efeitos:** valida E.164; persiste pendência; dispara template OTP WhatsApp para o número; **não** mergeia e **não** seta `primary_phone`.
- **Erros:** `400` formato inválido; `429` rate limit de reenvio; `502` falha ao enviar template Meta.
- **Notas:** no WhatsApp este endpoint quase não é usado (já identificado). Em IG/FB é o início da atravessia da parede.

### `POST /conversations/:id/phone/confirm` (confirmar OTP)
- **Auth:** Bearer **ou** chamada interna da triagem quando o cliente cola o código no chat IG/FB.
- **Request:** `{ code: string }`.
- **Response `200`:** `ConversationDetail` com `identification: "identified"`, `phone_gate: "collected"`, `customer` mergeado se aplicável.
- **Erros:** `400` código inválido/expirado; após N tentativas, invalidar pendência e voltar para `required`.
- **Notas:** só aqui ocorre promoção + merge por `primary_phone`.

### `POST /conversations/:id/phone/resend`
- **Auth:** Bearer / interno.
- **Request:** vazio (usa pendência atual).
- **Response `202`:** mesmo shape de pendência; novo OTP.
- **Erros:** `409` se não houver pendência; `429` se excedeu reenvios.

### Backend-interno (sem endpoint de frontend hoje)
- **Criação na entrada do webhook:** ver fluxo acima — WA identificado; IG/FB anônimo.
- **Merge por telefone:** só após OTP confirmado. 100% backend. Frontend nunca recebe `external_id` bruto para matching.
- **Revisão de merge sugerido por agente:** futuro (`GET /customers/merge-candidates`, `POST /customers/:id/merge`) — fora do escopo UI atual.

---

## 4. Inbox / Conversations / Messages / Claim / Pipeline / Triage

Uma `ConversationDetail` representa **uma thread de um canal com uma identidade de cliente** — um `Customer` pode ter várias conversas abertas simultaneamente em canais diferentes (ex.: uma no WhatsApp, outra no Instagram).

### `GET /inbox`
- **Auth:** Bearer.
- **Query:** `unit_id?`, `stage?: PipelineStage`, `channel?: ChannelType`, `mode?: ConversationMode`, `cursor?`, `limit?`.
- **Response:** `CursorPage<InboxItem>` (`InboxItem` = `ConversationSummary`) — paginação por cursor (`next_cursor: string | null`), não por página, pois é uma lista de alto volume/tempo real.
- **Campos de `ConversationSummary`:** `id, customer_id, customer_name, customer_phone?, identification, phone_gate, pending_phone_masked?, channel, channel_thread_id?, unit_id, pipeline_stage, mode, status, owner_id, intent, ai_summary, triage_notes?, last_message_preview, last_message_at, window_expires_at?, unread_count?`.
- **Notas:** `identification` / `phone_gate` / `pending_phone_masked` alimentam a parede de privacidade na UI. `window_expires_at` é a janela de 24h da Meta — aviso no frontend; bloqueio de envio fora da janela é do backend.

### `GET /conversations/:id`
- **Auth:** Bearer.
- **Response:** `ConversationDetail` = `ConversationSummary` + `{ customer: Customer, created_at, updated_at }`.

### `GET /conversations/:id/messages`
- **Auth:** Bearer.
- **Query:** `cursor?`, `limit?`.
- **Response:** `CursorPage<Message>` — `Message { id, conversation_id, direction: "in"|"out", sender_type: "contact"|"agent"|"ai"|"system", kind: MessageKind, channel, body, status: MessageStatus, meta_message_id?, reply_to_engagement_id?, media_url?, media_mime_type?, template_name?, created_at }`.
- **Notas:** `meta_message_id` é o id nativo da Meta (`wamid.*` para WA, `mid.*` para IG/FB) — necessário para reconciliar status de entrega/leitura vindos por webhook. `reply_to_engagement_id` liga a mensagem a um `SocialEngagement` de origem (ex.: resposta enviada a partir de uma resposta de story).

### `POST /conversations/:id/messages`
- **Auth:** Bearer. Requer que a conversa esteja em `mode: "human"` (ver claim abaixo) — enviar com `mode: "ai_triage"` deve ser rejeitado pelo backend (`409`), independente do que a UI já bloqueia.
- **Request:** `{ body: string, kind?: MessageKind }` (default `kind: "text"`).
- **Response `201`:** `Message` criada.
- **Notas:** backend decide qual API da Meta chamar (Cloud API para WA, Send API para IG/Messenger) conforme `channel` da conversa; decide se precisa usar template (fora da janela de 24h) — frontend não escolhe isso.

### `POST /conversations/:id/claim`
- **Auth:** Bearer.
- **Request:** vazio (usuário autenticado vira `owner_id`).
- **Response:** `ConversationDetail` atualizada, com `mode: "human"` e `owner_id` preenchido.
- **Erros:** `409` se já houver outro `owner_id` (a menos que o solicitante seja supervisor/admin fazendo reassign — política de negócio a definir com o time).
- **Notas:** este é o **handoff obrigatório** IA → humano. Ver regra "Triagem antes do humano" em [`PRODUCT-V2.md`](./PRODUCT-V2.md).

### `PATCH /conversations/:id/pipeline`
- **Auth:** Bearer.
- **Request:** `{ stage: PipelineStage, reason_code?: string, reason_text?: string }`
- **Response:** `ConversationDetail` atualizada.
- **Regra de negócio (backend deve validar, frontend só espelha em UX):** `reason_code` é **obrigatório** quando `stage === "nao_fechado"` e deve existir no catálogo retornado por `GET /loss-reasons`. Rejeitar com `400` se ausente/inválido.
- **Estágios válidos (`PipelineStage`):** `em_atendimento → em_negociacao → aguardando_fechamento → { fechado | nao_fechado }`. Definir no backend se transições "para trás" são permitidas (o frontend hoje não restringe a ordem na UI).

### `GET /conversations/:id/triage`
- **Auth:** Bearer.
- **Response:** `TriageSummary { conversation_id, intent, confidence?, summary, suggested_pops?, ready_for_handoff, phone_gate, pending_phone_masked?, collected_fields? }`.
- **Erros:** `404` é esperado e tratado silenciosamente pelo frontend quando a conversa já foi assumida por humano (não há mais triagem ativa) — não é um estado de erro real, apenas "não aplicável".
- **Notas:** conteúdo 100% gerado pela IA de triagem no backend (ver seção 10). O frontend só exibe. Quando `phone_gate === "required"`, a IA solicita telefone e chama `POST .../phone`. Quando `pending_verification`, orienta o cliente a informar o código recebido no WhatsApp e chama `POST .../phone/confirm`. Só então ações gated / `ready_for_handoff` pleno para intenções que exigem ficha.

---

## 5. Engagements (`story_reply`, `story_mention`, `post_comment`, `live_comment`)

Interações Meta-nativas **fora** do fluxo normal de mensagens 1:1: resposta a story, menção em story, comentário em post/live. Chegam por webhooks distintos dos de mensagens (ver seção 8) e são trabalhadas em fila própria antes de, opcionalmente, virarem uma `ConversationDetail` completa.

### `GET /engagements`
- **Auth:** Bearer.
- **Query:** `unit_id?`, `channel?: ChannelType`, `type?: EngagementType`, `status?: EngagementStatus`, `cursor?`, `limit?`.
- **Response:** `CursorPage<SocialEngagement>`.
- **Campos de `SocialEngagement`:** `id, customer_id?, customer_name?, channel, type, status, unit_id, media_id?, media_url?, media_caption?, body, external_id, author_external_id, conversation_id?, created_at, replied_at?`.
- **Tipos (`EngagementType`):** `story_reply | story_mention | post_comment | live_comment | private_reply`. `private_reply` é o registro da resposta privada enviada pelo agente a um comentário (nem sempre listado como item de fila próprio).
- **Status (`EngagementStatus`):** `open | replied | dismissed | converted_to_conversation`.
- **Notas:** `customer_id`/`customer_name` só existem quando `author_external_id` já foi resolvido para um `Customer` (mesma lógica de identidade da seção 3); pode chegar `null` para autor ainda não vinculado.

### `GET /engagements/:id`
- **Auth:** Bearer.
- **Response:** `SocialEngagement`.

### `POST /engagements/:id/reply`
- **Auth:** Bearer.
- **Request:** `{ body: string }`
- **Response:** `SocialEngagement` atualizado (`status: "replied"`, `replied_at` preenchido).
- **Notas:** backend decide a API Meta correta por tipo — resposta a comentário pode ser pública (reply ao comentário) **ou** privada (Private Reply API do Instagram/Facebook), e resposta a story reply/mention vai pelo Send API como DM. Essa escolha é de negócio/backend, não da UI.

### `POST /engagements/:id/dismiss`
- **Auth:** Bearer.
- **Response:** `SocialEngagement` com `status: "dismissed"`.

### `POST /engagements/:id/convert`
- **Auth:** Bearer.
- **Response `201`:** `ConversationDetail` recém-criada, com `conversation_id` também gravado de volta no `SocialEngagement` (`status: "converted_to_conversation"`).
- **Notas:** promove a interação para uma thread completa de conversa (ex.: um comentário vira um atendimento de pipeline). O backend deve garantir que a `Customer`/`CustomerIdentity` do autor já exista ou seja criada nesse momento.

---

## 6. Follow-ups, POPs, Loss reasons, Dashboard

### `GET /followups`
- **Auth:** Bearer.
- **Query:** `unit_id?`, `status?: FollowUpStatus`, `stage?: PipelineStage`, `cursor?`, `limit?`.
- **Response:** `CursorPage<FollowUp>` — `FollowUp { id, conversation_id, customer_id, customer_name, customer_phone?, unit_id, pipeline_stage, due_at, status, note, created_at, completed_at? }`.
- **Notas:** backend é responsável por gerar/agendar follow-ups automaticamente (ex.: ao mover para `aguardando_fechamento` ou `nao_fechado`) — não há endpoint de criação manual hoje no frontend.

### `POST /followups/:id/complete`
- **Auth:** Bearer.
- **Response:** `FollowUp` com `status: "done"`, `completed_at` preenchido.

### `POST /followups/:id/cancel`
- **Auth:** Bearer.
- **Response:** `FollowUp` com `status: "canceled"`.

### `GET /pops`
- **Auth:** Bearer.
- **Query:** `intent?: Intent`.
- **Response:** `Pop[]` (lista simples, sem paginação — volume esperado é baixo). `Pop { id, title, body, intent_tags: Intent[], active, created_at?, updated_at? }`.

### `GET /pops/:id`
- **Auth:** Bearer.
- **Response:** `Pop`.

### `POST /pops`
- **Auth:** Bearer (admin/manager).
- **Request:** `{ title, body, intent_tags: Intent[], active? }`
- **Response `201`:** `Pop`.

### `PATCH /pops/:id`
- **Auth:** Bearer (admin/manager).
- **Request:** mesmo shape de criação, parcial.
- **Response:** `Pop`.

### `DELETE /pops/:id`
- **Auth:** Bearer (admin/manager).
- **Response:** `204`.

### `GET /loss-reasons`
- **Auth:** Bearer.
- **Response:** `LossReason[]` — `{ code, label }`. Catálogo administrável (ver seed em `docs/decisions.md`), consumido pelo modal de pipeline ao mover para `nao_fechado`.

### `GET /dashboard/summary`
- **Auth:** Bearer.
- **Query:** `unit_id?` (omitido = visão consolidada de todas as unidades que o usuário acessa).
- **Response:** `DashboardSummary` — snapshot operacional derivado do fluxo (conversas, mensagens, follow-ups, engagements). Sem receita/ticket/período.
  - Contagens base: `open_conversations`, `by_stage: Record<PipelineStage,number>`, `by_channel: Record<ChannelType,number>`, `closed`, `not_closed`, `decided` (= closed + not_closed), `conversion_rate`, `ai_triage`, `human`.
  - Riscos / fila: `unclaimed` (abertas sem `owner_id`), `awaiting_reply` (abertas cuja última mensagem é `direction: "in"`), `stale_open` (abertas com `last_message_at` > 24h), `awaiting_phone` (`phone_gate` in `required | pending_verification`), `window_expiring` (`window_expires_at` nas próximas 4h), `awaiting_followup`, `overdue_followups` (follow-ups `open` com `due_at` no passado), `open_engagements`.
  - Demanda / canal: `by_intent: Record<Intent,number>` (só abertas; `intent` nulo conta em `outro`), `closed_by_channel`, `not_closed_by_channel`.
  - `units: DashboardUnitSummary[]` — `{ unit_id, unit_name, open, closed, not_closed, conversion_rate, unclaimed, awaiting_followup }` (sempre as unidades acessíveis ao usuário; métricas no escopo da query).
- **Notas:** `awaiting_reply` exige olhar a última mensagem da conversa. Não há comparativo temporal nem valores financeiros neste endpoint.

---

## 7. Meta settings (multicanal)

Configuração de conexão com a Meta por canal, por unidade (ou global, a definir pelo backend conforme modelo de WABA/Page). **Tokens de acesso nunca trafegam para o browser** — o cliente só vê `token_masked`.

### `GET /settings/meta`
- **Auth:** Bearer (admin).
- **Response:** `MetaSettings { channels: MetaChannelConfig[], ai_enabled, ai_system_prompt, ai_context, ai_campaigns: AICampaign[], triage_enabled, triage_handoff_intents: Intent[] }`.
- **`MetaChannelConfig`:** `{ channel: ChannelType, enabled, account_id, display_name, phone_number_id?, webhook_verified, token_masked }`.
  - `account_id` é o WABA id (WhatsApp), IG Business Account id (Instagram) ou Page id (Facebook).
  - `phone_number_id` só existe para `channel: "whatsapp"`.
  - `token_masked` é um preview tipo `EAAG...9f2a` — **nunca** o token completo.
- **`AICampaign`:** `{ id, title, description, starts_on, ends_on, active }`: janelas de contexto/campanha que alimentam o prompt da IA de triagem (ex.: campanha de gripe em maio).
- **UI Agenda:** a superfície principal de CRUD de `ai_campaigns` é a rota autenticada `/campaigns` (calendário mensal). `PUT /settings/meta` continua aceitando `ai_campaigns` (usado pela Agenda e por integrações). A tela Canais Meta não edita mais a lista; só resume e aponta para a Agenda.

### `PUT /settings/meta`
- **Auth:** Bearer (admin).
- **Request:** `UpdateMetaSettingsPayload` — parcial de `MetaSettings` exceto `channels`, mais:
  - `channels?: Array<Partial<Omit<MetaChannelConfig,"token_masked">> & { channel: ChannelType }>` — atualiza config por canal sem poder setar `token_masked` diretamente.
  - `channel_tokens?: Partial<Record<ChannelType,string>>` — **único jeito de rotacionar token de um canal**. Write-only: o backend recebe o token bruto, criptografa e persiste (ver seção 9); nunca o ecoa de volta.
- **Response:** `MetaSettings` fresco (sempre buscar do zero após salvar — o frontend nunca reaproveita `channel_tokens` enviado como estado local).

---

## 8. Webhooks Meta (backend-only)

**Nenhum destes endpoints é chamado pelo frontend, direta ou indiretamente via proxy.** Existem apenas para a Meta enviar eventos ao backend. Documentados aqui para alinhar expectativas de dado que eventualmente chega como `Message`/`SocialEngagement` na API do item 4/5.

Convenção sugerida de path (fora de `/api/v1`, ex. `/webhooks/meta/*`), cada um precisa de:
- **Verificação (`GET`)**: handshake `hub.mode=subscribe&hub.verify_token=...&hub.challenge=...` — responder `hub.challenge` em texto puro se `verify_token` bater com o configurado.
- **Recebimento (`POST`)**: validar assinatura `X-Hub-Signature-256` (HMAC-SHA256 com o App Secret) **antes** de processar qualquer payload — payload não assinado corretamente deve ser rejeitado com `403` e nunca persistido.

| Webhook | Canal | Eventos típicos | Produz |
|---|---|---|---|
| WhatsApp messages | `whatsapp` | mensagem recebida, status de entrega/leitura | `Message` (direction `in`), atualização de `MessageStatus` em mensagens `out` |
| Instagram messages | `instagram` | DM recebida, `reply_to.story` (resposta a story) | `Message` (DM normal) ou `SocialEngagement type: story_reply` |
| Instagram referrals | `instagram` | `messaging_referrals` (menção em story) | `SocialEngagement type: story_mention` |
| Facebook Messenger | `facebook` | DM recebida no Messenger da Page | `Message` |
| Comentários (IG/Page) | `instagram` \| `facebook` | comentário em post ou live | `SocialEngagement type: post_comment \| live_comment` |

Responsabilidades do backend em todo webhook:
1. Validar assinatura HMAC.
2. Resolver/criar `Customer` + `CustomerIdentity` a partir do `external_id` do canal (seção 3).
3. Idempotência — a Meta reenvia webhooks; usar `meta_message_id`/`external_id` como chave de deduplicação.
4. Rotear para IA de triagem (se `mode: "ai_triage"` e `triage_enabled: true`) ou manter em fila humana.
5. Responder `200` rapidamente (processamento pesado deve ser assíncrono — a Meta corta/reenvia em caso de timeout).

---

## 9. Requisitos de segurança

- **Tokens Meta**: nunca retornados em nenhuma resposta JSON além de `token_masked`. Armazenar criptografados em repouso (ex. AES-GCM com chave em KMS/secret manager), nunca em texto plano no banco.
- **Chaves OpenAI/LLM**: vivem exclusivamente em variáveis de ambiente do backend. Nenhum endpoint deve retorná-las, logá-las ou aceitá-las como input de cliente.
- **JWT via BFF**: o backend emite JWT normalmente (`Authorization: Bearer`), mas quem fala com o backend é sempre o servidor Next.js — o browser nunca tem o token em `localStorage`/`sessionStorage`/JS acessível. Cookies `cv_access`/`cv_refresh` são `httpOnly`, `SameSite=Lax`, e `Secure` em produção.
- **CORS**: como o browser só fala com o próprio domínio do frontend (rotas `/api/*` same-origin), o backend **não precisa** aceitar requisições CORS de origens de browser. Recomenda-se restringir CORS do backend a nenhuma origem de navegador (allowlist vazia ou apenas o servidor do BFF, se aplicável) — qualquer chamada direta ao backend a partir de um browser deve ser rejeitada.
- **Assinatura de webhook**: obrigatória em 100% dos endpoints da seção 8, sem exceção, mesmo em ambiente de desenvolvimento.
- **Rate limiting / abuse**: recomenda-se limitar `POST /auth/login` e `POST /auth/refresh` por IP/usuário para mitigar brute-force, já que essas rotas são as únicas sem Bearer.
- **Auditoria**: ações sensíveis (`claim`, mudança de `pipeline_stage`, alteração de `settings/meta`, criação/remoção de usuário) devem ser logadas de forma append-only (ver `docs/decisions.md`).

---

## 10. O que o frontend nunca processa

Por design, os itens abaixo são responsabilidade **exclusiva** do backend. O frontend não tem código, lógica ou acesso a dados para nenhum deles:

- **IA de triagem** — geração de `TriageSummary` (intenção, confiança, resumo, POPs sugeridos, `collected_fields`). O frontend só exibe o resultado.
- **Roteamento de mensagens** — decidir qual API da Meta chamar (Cloud API, Send API, Private Reply API), gestão de janela de 24h, fallback de template.
- **Matching/linking de identidade** — resolver `external_id` para `Customer.id`; OTP WhatsApp antes de setar `primary_phone`; merge só após confirmação; enforcement da parede (`403 phone_required` / `phone_verification_pending`).
- **Validação de assinatura de webhook** — HMAC de payloads da Meta.
- **Criptografia/gestão de tokens** — cifrar, rotacionar e decifrar tokens de canal; nunca chegam ao browser além do `token_masked`.
- **Validação de regras de negócio do pipeline** — ex.: `reason_code` obrigatório em `nao_fechado`, transições de estágio permitidas, bloqueio de envio com conversa em `ai_triage`. O frontend **pode espelhar essas mesmas regras apenas como validação de UX** (desabilitar botão, mostrar campo obrigatório) para dar feedback rápido ao usuário — mas o backend deve reforçá-las de forma independente, pois é a única fonte de verdade.
