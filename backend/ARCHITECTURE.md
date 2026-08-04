# Arquitetura do Backend — CRM Cia da Vacina

**Status:** roadmap dos 10 passos **concluído** e testado contra Postgres real (ver §8). Três gaps conhecidos, todos pela mesma razão (sem credencial/ambiente real neste sandbox de dev): clients HTTP reais da Meta (§6), o client HTTP real da OpenAI só testado contra um servidor local, nunca contra `api.openai.com` de verdade (§7), e os payloads de webhook de comentário/story (§8) parseados conforme documentação pública da Meta, nunca vistos vindo de um app real. Em todos os três casos a lógica em volta (parse, roteamento, idempotência, regras de negócio) está genuinamente testada — só a chamada de rede/o formato exato do payload de origem não foram validados contra a Meta/OpenAI de verdade.
**Autor:** Felipe (Backend/Go), com apoio do Claude.
**Referência de padrões:** `neohabit/backend` (outro projeto da empresa) — handler → usecase → repository, wiring por módulo, Docker + Air.
**Fonte da verdade funcional:** [`../docs/BACKEND-CONTRACT.md`](../docs/BACKEND-CONTRACT.md), [`../docs/openapi.yaml`](../docs/openapi.yaml), [`../docs/PRODUCT-V2.md`](../docs/PRODUCT-V2.md), [`../docs/decisions.md`](../docs/decisions.md), [`../docs/adr/0001-simple-monolith.md`](../docs/adr/0001-simple-monolith.md).

Stack: Go 1.26 · chi (via wrapper) · sqlx + pgx · PostgreSQL · golang-migrate · Air (hot reload) · Docker.

---

## 1. Estrutura de pastas

Mesmo esqueleto de camadas do `neohabit/backend`, adaptado ao domínio do CRM. ✅ = já existe no repo; sem marca = planejado, ainda não construído.

```
backend/
├── cmd/
│   ├── server/main.go      # ✅ boot, middlewares globais, registro de rotas
│   ├── migrate/main.go     # ✅ golang-migrate (up/down/drop/force)
│   └── seed/main.go        # ✅ seeds (essential + demo)
├── internal/
│   ├── app/app.go           # ✅ struct App{DB, JWT, Meta, SSE} — DI raiz
│   ├── domain/
│   │   ├── entity/          # ✅ structs planas 1:1 com tabelas (db tags)
│   │   └── vo/               # ✅ Password (argon2id), Phone (E.164 + máscara)
│   └── module/
│       ├── middleware/       # ✅ RequireAuth, RequireRole, ClaimsFromContext
│       ├── auth/             # ✅ login/refresh/logout
│       ├── user/             # ✅ CRUD + PUT /users/:id/units
│       ├── unit/              # ✅ CRUD
│       ├── me/                # ✅ GET /me
│       ├── customer/         # ✅ Customer + CustomerIdentity (GETs; merge fica no conversation)
│       ├── conversation/     # ✅ inbox, detail, messages, claim, pipeline, phone/* (OTP+merge), SSE
│       ├── triage/           # ✅ RunTriage (chamado pelo webhook) + GET /conversations/:id/triage
│       ├── engagement/       # ✅ story_reply/mention, comentários — list/get/reply/dismiss/convert + ingestão via webhook
│       ├── lossreason/       # ✅ GET /loss-reasons
│       ├── followup/         # ✅ list + complete/cancel (criação automática vive em conversation/usecase)
│       ├── pop/              # ✅ CRUD, intent_tags como JSONB
│       ├── dashboard/        # ✅ GET /dashboard/summary
│       ├── metasettings/     # ✅ GET/PUT /settings/meta (tokens cifrados/mascarados, campanhas)
│       ├── auditlog/          # ✅ GET /audit-logs (admin-only) — leitura do que pkg/audit grava
│       └── webhook/          # ✅ GET/POST /webhooks/meta/{channel} (fora de /api/v1) — HMAC + ingestão real
├── pkg/
│   ├── http/        # ✅ Router (wrap chi), response envelope, ParseAndValidate — igual ao neohabit
│   ├── apperrors/    # ✅ ResponseError/SystemError + MapDBError — igual
│   ├── database/     # ✅ sqlx+pgx wrapper — igual
│   ├── env/, jwt/, validation/  # ✅ iguais (validation ganhou tag `e164`)
│   ├── cursor/        # ✅ NOVO, não previsto no plano original — paginação por cursor opaco (CursorPage<T>), usado por inbox e mensagens; será reusado por engagements/follow-ups
│   ├── sse/           # ✅ hub de broadcast + protocolo text/event-stream, com heartbeat
│   ├── meta/           # ✅ Sender/CommentResponder + MockClient + VerifySignature (HMAC webhook) — nenhuma chamada de rede real ainda (ver §6)
│   ├── crypto/        # ✅ AES-GCM pra token Meta em repouso, testado (round-trip + tamper detection)
│   ├── openai/         # ✅ Client/HTTPClient/MockClient — mesma dualidade de pkg/meta; HTTPClient testado de verdade contra httptest.Server
│   ├── audit/           # ✅ Logger.Log — grava audit_logs (append-only), chamado direto pelos usecases sensíveis (claim, pipeline_change, settings.meta.update, user.create/delete)
│   └── jobs/           # poller sobre tabela `jobs` — segue não sendo necessário (ver §2)
└── conf/migrations/ + Dockerfile(.dev) + docker-compose.yml + .air.toml + Makefile  # ✅
```

Cada módulo repete o padrão `handler → usecase → repository` (+ `model` quando o shape de saída difere da entity), com `module.go` fazendo a fiação (`New(a *app.App) *Module` + `Register(r *httppkg.Router)`), exatamente como `auth` e `unit` no neohabit. `handler` só faz parse/validate/chamada/resposta; toda regra de negócio (parede de privacidade, transições de pipeline, `reason_code` obrigatório, roteamento de canal Meta) fica isolada em `usecase`, testável sem HTTP nem banco real.

Não existe módulo `inbox` separado — `GET /inbox` é só mais um método de `conversation/usecase`, já que `InboxItem = ConversationSummary` no contrato.

---

## 2. Infra compartilhada (`pkg/`)

### Copiado ~1:1 do neohabit — ✅ feito

- `pkg/http` — `Router` (wrap de chi), `Success/Error/Handle`, `ParseAndValidate`/`ParseAndValidateOptional`.
- `pkg/apperrors` — `ResponseError` (regra de negócio, status certo) vs `SystemError` (500 + causa logada), incluindo `MapDBError` pra unique/FK violation do Postgres. Ganhou `NewConflictErrorMessage`, `NewTooManyRequestsError`, `NewBadGatewayError` no passo 4 (409/429/502 do fluxo de OTP).
- `pkg/database` — sqlx+pgx, `Transaction` helper.
- `pkg/env`, `pkg/validation` (validator/v10, + tag `e164`).
- `pkg/jwt` — **adaptado**: HS256 com secret único via env (não RS256 com par de chaves — ver §5).

### Novo, específico do domínio

- **`pkg/sse`** ✅ — hub de broadcast simples (fan-out pra todas as conexões abertas no processo) + heartbeat de 15s; a filtragem por unidade acontece no handler de cada módulo (`allowed func(unitID) bool`), não no hub. Alimenta `GET /inbox/stream` (endpoint próprio, não especificado literalmente no contrato — ver nota em §5).
- **`pkg/meta`** ✅ — `Sender`/`CommentResponder` (interface comum de envio) + `Registry` (resolve client por canal) + `MockClient` (loga o envio, sem chamada de rede). Já está **conectado** em `app.go` e em uso real por `conversation/usecase` (`POST .../messages` e o OTP passam pela interface de verdade). Troca por client HTTP real é só registrar outro `Sender` — nenhum usecase muda. Isso adianta parte do passo 6.
- **`pkg/cursor`** ✅ — não estava no plano original; ficou claro no passo 4 que `CursorPage<T>` (cursor opaco base64 de `timestamp|id`) é usado em quase todo endpoint de alto volume do contrato (inbox, mensagens, e adiante engagements/follow-ups), então virou pacote genérico em vez de lógica duplicada por módulo.
- **`pkg/crypto`** ✅ — AES-GCM (256 bits, chave via `APP_ENCRYPTION_KEY` hex) pra cifrar `channel_tokens` em repouso. `token_masked` é calculado uma vez no momento da rotação (a partir do plaintext, antes de descartá-lo) — o backend nunca decifra só pra exibir preview.
- **`pkg/jobs`** — tabela `jobs` + poller. Reconsiderado: TTL de OTP (`pending_verification` expirando) foi resolvido **sem job** — é lazy-checado na leitura e revertido na próxima escrita. Webhook também não precisou (ingestão síncrona é rápida o bastante — poucas queries — pra responder antes do timeout da Meta). `pkg/jobs` segue não sendo uma certeza do roadmap; só entraria se retry de envio Meta (mensagem falhou, precisa tentar de novo em background) virar necessidade real.
- **`pkg/audit`** ✅ — `Logger.Log(ctx, Entry)` grava em `audit_logs`, chamado direto pelos usecases (`conversation.Claim`/`UpdatePipeline`, `metasettings.Update`, `user.Create`/`Delete`) — não é um middleware genérico porque só o usecase sabe o que de fato mudou (ex.: de qual stage pra qual). Nunca retorna erro pro caller: falha ao auditar não derruba a ação de negócio já concluída, só fica logada. `internal/module/auditlog` expõe a leitura (`GET /audit-logs`, admin-only).

### Não entra no MVP

- MinIO/storage — mídia é V1.1, MVP é texto-first (`decisions.md`).
- Redis — proibido no MVP pelo ADR 0001.
- Docusaurus/telm (collector) — infra de documentação/observabilidade pesada pra um time de 2 devs; docs já vivem em markdown no repo.
- OpenTelemetry — **decisão confirmada: fica pra depois.** Log estruturado simples (stdlib + request-id) cobre o que o contrato pede agora.

---

## 3. Modelagem — tabelas (migrations)

**✅ Criadas (migrations 000001–000017):** `pgcrypto` (extensão, infra), `users`, `units`, `user_unit_relation`, `refresh_tokens`, `customers`, `customer_identities`, `loss_reasons` (seed fixo de `decisions.md` D-06 — antecipada do passo 5 pro passo 4 porque `PATCH .../pipeline` precisa validar `reason_code` contra ela), `conversations`, `messages`, `phone_verifications`, `follow_ups`, `pops` (`intent_tags` como JSONB, não `TEXT[]` — evita depender de scan de array do Postgres via pgx/sqlx), `meta_channel_configs` (WhatsApp por unidade, Instagram/Facebook globais — ver §5), `ai_campaigns`, `app_settings` (singleton — `id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1)`, guarda config de IA/triagem + `default_unit_id`), e 3 colunas novas em `conversations` (`collected_fields` JSONB, `triage_confidence`, `triage_ready_for_handoff`) — persistem o que `GET /conversations/:id/triage` devolve sem precisar re-chamar a IA a cada leitura.

**✅ Criada (migration 000018):** `social_engagements` (`channel`/`type`/`status` como `TEXT + CHECK`, `UNIQUE(channel, external_id)` pra idempotência de webhook, `customer_id`/`conversation_id` nullable com `ON DELETE SET NULL`) e `messages.reply_to_engagement_id` (adicionada só agora, depois que `social_engagements` existe — ver nota abaixo).

**✅ Criada (migration 000019):** `audit_logs` (append-only — nenhum código faz UPDATE/DELETE nela, só INSERT via `pkg/audit`; `actor_user_id`/`unit_id` nullable com `ON DELETE SET NULL`, `metadata` JSONB).

**Ainda não criadas:** `jobs` (se necessário — ver §2).

`Customer.identification` e `Conversation.phone_gate`/`pipeline_stage`/`mode` são `TEXT + CHECK`, não enum nativo do Postgres — mais simples de alterar via migration do que um `ALTER TYPE`.

`messages.reply_to_engagement_id` foi propositalmente deixado de fora da migration 000010 — só entrou na 000018, junto com `social_engagements`, pra cada migration referenciar só tabelas que já existem no momento em que roda.

---

## 4. Auth / RBAC

✅ Implementado e testado. Mesmo padrão de claims do neohabit (JWT carrega o que o usuário pode ver, evita bater no banco em toda request): `role` + `unit_ids` embutidos no access token, calculados no login/refresh. `middleware.RequireAuth` + `middleware.RequireRole(roles...)` cobrem autenticação/papel; escopo por unidade é decidido em cada usecase (`Access.canAccessUnit`), não por um middleware genérico de query param — na prática nenhum endpoint até agora precisou de um `RequireUnitAccess` central, já que o padrão é "admin vê tudo, resto filtra pelos `unit_ids` dos claims" aplicado dentro da regra de negócio de cada recurso.

Duração de token conforme `decisions.md`: access 15min, refresh 7 dias. Confirmado por teste: `POST /auth/logout` revoga **todas** as sessões do usuário (não uma específica) — é a única leitura possível já que o endpoint não recebe `refresh_token` no corpo (só Bearer), ver `auth/repository.RevokeAllUserRefreshTokens`.

---

## 5. Decisões confirmadas

| Decisão | Escolha | Motivo |
|---|---|---|
| Assinatura JWT | **HS256** (secret único via env) | Monólito único; o BFF só repassa o Bearer, não valida assinatura. |
| Telemetria (OTel) | **Deixar pra depois** | Log estruturado simples cobre o que o contrato pede agora. |
| `POST .../claim` em conversa já reivindicada | **Admin/manager/supervisor podem reassumir** (409 só pra agent comum) | Contrato marcava como "a definir com o time" — confirmado no passo 4. Cobre "atendente ausente, gerente realoca". |
| `PATCH .../pipeline` — ordem de transição | **Qualquer transição entre estágios válidos é permitida**, sem impor sequência | Contrato marcava como "a definir com o time" — confirmado no passo 4. Backend só valida pertinência ao enum + `reason_code` obrigatório em `nao_fechado`; frontend já não restringia ordem na UI. |
| `GET /inbox/stream` (SSE) | Endpoint próprio, path não especificado no `BACKEND-CONTRACT.md` | `decisions.md` só compromete com "SSE pra inbox" como tecnologia, sem definir o path. Escolhido `GET /inbox/stream`, escopado por unidade igual ao resto do módulo — ajustável sem custo se o frontend definir algo diferente. |
| Janela de follow-up (`due_at`) | **72h a partir da criação**, fixo | Contrato só diz "backend gera/agenda automaticamente", sem número. Default operacional de baixo risco, ajustável sem migração (`conversation/usecase.followUpDefaultWindow`). |
| Duplicar follow-up ao oscilar entre estágios | **Não duplica** — só cria se não houver um já aberto pra conversa | Índice único parcial (`status='open'`) + `ON CONFLICT DO NOTHING`. Evita spam se o agente mover a conversa pra frente/trás repetidamente (transições livres, ver decisão acima). |
| `units[]` do dashboard vs. filtro `unit_id?` | **Independentes** — `units[]` sempre lista todas as unidades acessíveis ao usuário, com métricas próprias, mesmo quando `unit_id` filtra os números agregados do topo | Leitura mais literal de "sempre as unidades acessíveis ao usuário" no contrato — permite comparar unidades mesmo filtrando o resumo geral numa só. |
| `meta_channel_configs` por unidade ou global? | **WhatsApp por unidade** (5 linhas, uma por unidade, D-01); **Instagram/Facebook globais** (1 linha cada, `unit_id` NULL) | Contrato marcava como "a definir pelo backend". Confirmado com o stakeholder: as 5 unidades têm WhatsApp próprio mas Instagram/Facebook são uma conta central da marca. `CHECK` na tabela garante essa regra no schema (`unit_id` obrigatório só quando `channel='whatsapp'`). |
| Conversa nova de canal centralizado (IG/FB) — qual unidade? | **`app_settings.default_unit_id`**, configurável via `PUT /settings/meta` | Consequência direta da decisão acima — sem `phone_number_id`, não há como rotear sozinho. Evitou tornar `conversations.unit_id` nullable (mudaria toda a lógica de escopo já construída nos passos 2–5) e evitou inventar um endpoint de "atribuir unidade" não previsto no contrato. Se `default_unit_id` não estiver configurado, a ingestão do webhook loga e descarta a mensagem (não cria dado incompleto) — só acontece antes do setup inicial. |
| `PUT /settings/meta` — `channel_tokens` (mapa global do contrato) | **Substituído por um campo `token` por item de `channels[]`**, incluindo `unit_id` quando aplicável | Consequência das duas decisões acima: um mapa `{whatsapp: "token"}` não identifica QUAL das 5 unidades rotacionar. Extensão documentada em `metasettings/model/model.go`. |
| Clients HTTP reais da Meta (WhatsApp Cloud API, Instagram Send API) | **Não escritos neste MVP de dev** — só a interface (`pkg/meta.Sender`) e o mock, já em uso real por `conversation`/`webhook` | Sem credencial de app Meta real neste ambiente, um client real seria código nunca exercitado de verdade — prefiro isso explícito a fingir que "funciona" sem ter rodado contra a API real nem uma vez. Trocar o mock por um client real é mecânico (implementar a interface, registrar em `app.go`) quando existir credencial de produção. |
| `ai_enabled` vs `triage_enabled` (dois campos separados no contrato) | **Papéis diferentes**: `triage_enabled=false` faz conversa nova nascer direto em `mode:"human"` (pula a IA por completo, `docs/PRODUCT-V2.md §6`); `ai_enabled=false` é o kill-switch de *chamar o modelo* — conversas continuam nascendo em `ai_triage`, só ficam esperando claim manual sem resposta automática | Contrato não detalha a diferença entre os dois; essa leitura segue literalmente o texto de `PRODUCT-V2.md` pro primeiro campo e usa o segundo como o "atrás de feature flag `ai.enabled`" de `decisions.md` D-02. Os dois testados via API de verdade (ligar/desligar cada um, mandar mensagem, checar `mode`/`intent`). |
| Onde a IA de triagem roda | **Síncrona dentro da ingestão do webhook** (`webhook/usecase` chama `triage.RunTriage` depois de criar a mensagem), não um endpoint separado nem fila | `GET /conversations/:id/triage` só lê o que já foi persistido — contrato diz "conteúdo 100% gerado pela IA... frontend só exibe", não "gerado on-demand a cada GET". Evita chamar a OpenAI a cada vez que alguém abre a tela de uma conversa. |
| `collected_fields`/`confidence`/`ready_for_handoff` de `TriageSummary` | **3 colunas novas em `conversations`** (não um objeto solto em memória) | Sem persistir, `GET .../triage` não teria o que devolver depois que a request de ingestão já terminou — são resultado do processamento assíncrono em relação à leitura, precisam sobreviver entre requests. |
| Client HTTP real da OpenAI | **Escrito e testado de verdade** (`httptest.Server`, não só mock) — diferente da decisão sobre os clients Meta acima | A API de chat da OpenAI é um único endpoint JSON estável e bem documentado (não 3 APIs distintas como a Meta) — dá pra testar a construção da request e o parse da resposta com confiança real, mesmo sem `OPENAI_API_KEY`. O que não foi validado é só "o modelo responde bem" (questão de prompt, não de código) e a rede até `api.openai.com` de fato. |
| `POST /engagements/:id/reply` de `post_comment`/`live_comment` — pública ou privada? | **Sempre `ReplyPrivate`** (Private Reply API/DM), nunca `ReplyPublic` | Contrato marcava como "a definir pelo backend". `story_reply`/`story_mention` só têm caminho de DM mesmo (não existe "resposta pública" a uma story); estendi o mesmo default pra `post_comment`/`live_comment` porque uma resposta pública sob o comentário pode expor dado de saúde do cliente (o que ele perguntou) pra qualquer visitante do post — DM é o default mais seguro em contexto de vacina/saúde. Decisão de baixo risco: trocar pra `ReplyPublic` é uma linha em `engagement/usecase.Reply`, não uma migration. |
| Shape dos payloads de webhook de comentário/story | **Best-effort a partir da documentação pública da Meta** (Instagram Messaging API pra `reply_to.story`/`story_mention`; Graph API `comments`/`live_comments`; Facebook Page `feed` com `item:"comment"`) | Mesma ressalva de §5 sobre os clients HTTP Meta: sem app real neste ambiente pra confirmar o shape exato campo-a-campo. Testado ponta a ponta com payloads sintéticos assinados via HMAC (`X-Hub-Signature-256` válido, mesma verificação que um payload real passaria) — a lógica de parse/roteamento/idempotência está exercitada de verdade, só o shape de origem é inferido. |
| Unidade de engagements Instagram/Facebook | **`app_settings.default_unit_id`**, mesma resolução usada pra mensagens de canal centralizado | Consequência direta da decisão do passo 6 (IG/FB são contas centrais da marca, não por unidade) — engagements herdam a mesma regra, sem reinventar roteamento. |
| `GET /audit-logs` — existe endpoint de leitura? | **Sim, criado** (`GET /audit-logs`, admin-only, cursor-paginado) | Não previsto no `openapi.yaml`/`BACKEND-CONTRACT.md` — o contrato só exige a escrita append-only (§9). Um log que ninguém consegue ler pela API não cumpre o propósito de auditoria; extensão de baixo risco (só leitura, admin-only, sem side-effect). |
| Rate limiting além de login/refresh | **Nenhum endpoint novo recebeu limiter dedicado** — o limiter global (`httprate.LimitByIP(500, 1min)`, já existia desde o passo 1) cobre o resto | `BACKEND-CONTRACT.md` §9 só recomenda limitar `login`/`refresh` especificamente (10/min, já existia) por serem as únicas rotas sem Bearer — todo o resto exige autenticação, então o limiter global genérico já é a mitigação correta contra abuso de conta comprometida/token vazado. |
| CORS travado em produção | **Nenhuma mudança de código necessária** — `pkg/http.CORS` já lê `CORS_ALLOWED_ORIGINS` do env desde o passo 1, default `*` só quando a env var não está setada | Revisado no passo 10 e confirmado que o mecanismo já existia certo: travar em produção é uma questão de configurar a env var no deploy (domínio do BFF Next.js), não de código. |

---

## 6. O que vale complexidade vs. o que fica simples

### Vale investir (diferencial do produto, contrato exige correção rígida) — ✅ feito no passo 4

- **Máquina de estados `phone_gate` + OTP WhatsApp + merge por `primary_phone`** — implementado e testado ponta a ponta, incluindo o caso de merge real (duas identidades, dois canais, um `Customer` final) e o self-heal de TTL/tentativas estouradas (reverte pra `required` sozinho, sem precisar de `pkg/jobs`).
- **Roteamento de mensagem por canal** — `conversation/usecase` já resolve o `Sender` certo via `pkg/meta.Registry` a partir do canal da conversa; janela de 24h/template fallback ainda não implementados (não havia necessidade ainda — mensagem de agente sempre livre no MVP atual).
- **Verificação HMAC de webhook + idempotência** — ✅ feito no passo 6, testado com assinatura válida/inválida/corpo adulterado (`pkg/meta.VerifySignature` + testes) e com reenvio do mesmo payload (mesmo `meta_message_id` — segunda vez não duplica mensagem, confirmado via teste real contra o Postgres).
- **Resolução de identidade + merge no webhook** — ✅ feito: WhatsApp nasce `identified` (com merge automático se o telefone já existir via um merge anterior por OTP); Instagram/Facebook nascem `anonymous`; conversa é reusada (não duplicada) quando o mesmo cliente manda várias mensagens seguidas no mesmo canal.

### Mantido deliberadamente simples

- **IA de triagem**: ✅ feito assim — uma chamada síncrona ao `gpt-4o-mini` (OpenAI) por mensagem recebida, atrás de `ai_enabled`, resposta em JSON estruturado (`response_format: json_object`) parseada direto, sem parser de linguagem natural nem regex. Nada de agent framework, memória vetorial, RAG ou orquestração multi-step — o escopo do produto é "saudação + intenção + resumo", não conversação contínua. Regras de negócio sensíveis (phone_gate por canal, handoff por intent configurado) ficam em código determinístico *depois* da resposta do modelo, nunca confiadas cegamente a ele — testado com 6 casos via `usecase_test.go` (mock de repositório + `openai.MockClient` configurável).
- **Jobs assíncronos**: só entra se algo realmente precisar (ver §2) — TTL de OTP já não precisou.
- **Dashboard**: ✅ feito assim — ~10 queries `COUNT`/`GROUP BY` diretas em sequência no usecase (uma por métrica/breakdown), sem camada de BI/materialized view nem mega-query única. Testado com dado real acumulado dos passos anteriores, números batendo.
- **Auditoria**: ✅ feito assim — `INSERT` append-only (`pkg/audit.Logger.Log`) chamado explicitamente nos 4 usecases que o contrato marca como sensíveis (`conversation.Claim`, `conversation.UpdatePipeline`, `metasettings.Update`, `user.Create`/`Delete`), não um middleware genérico — só o usecase sabe o metadata relevante de cada ação (ex.: de/pra qual stage). Testado ponta a ponta: cada ação gerou a entrada esperada em `GET /audit-logs`, com `actor_name` resolvido via join.
- **Engagements**: ✅ feito assim — módulo fino (`list`/`get`/`reply`/`dismiss`/`convert`) reusando `pkg/meta.CommentResponder` (mesmo `Registry` de conversations) e a mesma resolução de identidade/unidade centralizada do webhook (duplicada localmente por convenção de módulo autocontido, não importada de `webhook`). `Convert` reusa a lógica de criar `Customer`/`Conversation` já usada na ingestão de mensagem; conversa nasce em `mode:"human"` (não `ai_triage`) porque foi um agente que decidiu puxar aquele engagement pra fila, não um primeiro contato.
- Sem MinIO/mídia agora (V1.1), sem Redis/cache agora (ADR).

---

## 7. Docker / dev experience

✅ Tudo funcionando: `Dockerfile.dev` com Air, `Dockerfile` multi-stage terminando em `scratch`, `docker-compose.yml` com só `api` + `postgres` (nome de projeto explícito `cia-da-vacina-crm` pra não colidir com o volume do neohabit, que também tem uma pasta `backend/`), `Makefile` com os alvos de sempre, `.air.toml` apontando `./cmd/server`.

---

## 8. Ordem de implementação

1. ✅ **Bootstrap**: Docker + Air + Makefile + `app.go` + `/health` + migrate/seed rodando.
2. ✅ **Auth + Users + Units + Me** — fecha 100% do `openapi.yaml`.
3. ✅ **Customers + identidade** — `Customer`/`CustomerIdentity`, 3 GETs. (`phone_gate`/OTP foi replanejado pro passo 4 durante a execução — ver nota abaixo.)
4. ✅ **Inbox/Conversations/Messages + claim + pipeline + SSE + parede de privacidade completa** — absorveu o que seria o fim do passo 3 (`phone_gate`, OTP, merge) porque esses conceitos são definidos *por conversa* no contrato, e `Conversation` só passou a existir neste passo. Também antecipou parte do passo 5 (`loss_reasons`, pela validação de `reason_code`) e parte do passo 6 (`pkg/meta` conectado de verdade, não só o mock isolado).
5. ✅ **Follow-ups + POPs + Dashboard** — `GET /loss-reasons`, CRUD de `pops` (`intent_tags` filtrável via JSONB), `follow_ups` com criação automática religada em `conversation/usecase.UpdatePipeline` (ao entrar em `aguardando_fechamento`/`nao_fechado`, idempotente), `list`+`complete`+`cancel`, e `GET /dashboard/summary` inteiro (todos os campos do contrato, incluindo o breakdown por unidade).
6. ✅ **Meta settings + webhooks reais** — `GET`/`PUT /settings/meta` (tokens cifrados AES-GCM/mascarados, campanhas, canais por unidade/globais), `GET`/`POST /webhooks/meta/{channel}` com handshake, verificação HMAC, parse real de payload (WhatsApp Cloud API + Instagram/Messenger unificados), resolução de identidade/unidade/conversa, idempotência. Únicas peças fora: clients HTTP reais da Meta (sem credencial pra testar — ver §5/§6) e `GET /conversations/:id/triage` (fase 7).
7. ✅ **IA de triagem** — `pkg/openai` (client real testado via `httptest` + mock configurável), `triage/usecase.RunTriage` chamado pelo webhook após cada mensagem inbound (prompt monta system+contexto+histórico, pede JSON estruturado, aplica phone_gate/handoff determinístico por cima da resposta do modelo, manda a resposta da IA de volta via `pkg/meta.Sender`), `GET /conversations/:id/triage` servindo do que já foi persistido. Os dois kill-switches (`ai_enabled`, `triage_enabled`) testados via API de verdade.
8. ✅ **Engagements** (story/comment) — migration `social_engagements` + `messages.reply_to_engagement_id`; `GET /engagements`, `GET /engagements/:id`, `POST .../reply` (via `pkg/meta.CommentResponder.ReplyPrivate`), `POST .../dismiss`, `POST .../convert` (cria `Customer`/`Conversation` em `mode:"human"`, reusando identidade existente quando já houver); ingestão passiva via `webhook/usecase` (nova função `ingestEngagements`, parsers `parseStoryEngagements`/`parseComments`, interface `Engagement` construída via `engagement.NewUseCase(a)` mesma convenção de `Triage`); `dashboard.open_engagements` ligado a uma query real (`COUNT(*) WHERE status='open'`). Testado ponta a ponta contra Postgres real: ingestão de `post_comment` (Instagram `comments` e Facebook `feed`), `story_reply`, `story_mention`, idempotência (reenvio do mesmo `external_id` não duplica), `dismiss`/`reply`/`convert`, conflito em ação repetida (409), escopo por unidade (agente de outra unidade recebe 404, não vê o item na lista).
9. ~~OTP WhatsApp cross-canal~~ — **feito no passo 4**, item removido daqui.
10. ✅ **Auditoria + rate limiting + hardening final** — migration `audit_logs` (append-only) + `pkg/audit.Logger` chamado nos 4 pontos que o contrato marca como sensíveis (claim, pipeline_change, settings.meta.update, user.create/delete) + `GET /audit-logs` (admin-only, cursor-paginado, extensão de baixo risco — ver §5). Rate limiting e CORS revisados: ambos já estavam corretos desde o passo 1 (limiter dedicado em login/refresh + limiter global genérico cobrindo o resto; CORS já lia `CORS_ALLOWED_ORIGINS` do env) — nenhuma mudança de código necessária, só confirmação. Testado ponta a ponta contra Postgres real: cada ação sensível gerou a entrada de auditoria esperada, `GET /audit-logs` só acessível por admin (403 pra outros papéis, 401 sem token).

Essa ordem prioriza destravar o frontend cedo (login, inbox, pipeline) antes de entrar na parte mais cara (integração Meta real + IA) — e na prática o passo 4 acabou puxando pra frente mais peça de infra (`pkg/meta` real, `loss_reasons`) do que o previsto, porque fazia sentido testá-las juntas em vez de deixar pontas soltas. O passo 5 já veio mais limpo — encaixou exatamente no que sobrou planejado, sem puxar nada do passo 6 pra frente.

A partir do passo 6, a maior parte da "engenharia difícil" de identidade/privacidade/pipeline já estava resolvida e testada; o que faltava era sobretudo integração externa real (Meta, OpenAI) e módulos mais finos por cima da infra que já existe. Os passos 8 e 10 confirmaram essa leitura — nenhum dos dois precisou de peça de infra genuinamente nova além de `pkg/audit` (que por sua vez é só mais um wrapper fino sobre `*database.DB`, no mesmo espírito de `pkg/sse`/`pkg/crypto`). **Os 10 passos do roadmap original estão concluídos.** O que resta como trabalho futuro genuíno (não bloqueante pro MVP) é só: (1) clients HTTP reais da Meta quando existir credencial de produção, (2) validar o client OpenAI e os parsers de webhook de engagement contra tráfego real, (3) `pkg/jobs` se retry assíncrono de envio Meta virar necessidade.
