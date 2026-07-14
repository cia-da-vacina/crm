# CRM Cia da Vacina — Especificação Técnica (Spec-Kit)

> **Status:** APROVADO para implementação (Sprint 1+)  
> **Data de aprovação:** 2026-07-14  
> **Canal MVP:** WhatsApp Meta Cloud API  
> **IA MVP:** Triagem inicial + handoff humano  
> **Equipe:** Felipe (Backend/Go) · Cristian (Frontend/Next.js)

---

# CRM Cia da Vacina â€” EspecificaÃ§Ã£o TÃ©cnica (Spec-Kit)

**DecisÃµes fechadas com o stakeholder**
- Canal MVP: **somente WhatsApp** via **Meta Cloud API**
- IA MVP: **triagem inicial** (saudaÃ§Ã£o, identificaÃ§Ã£o de necessidade, roteamento) + **handoff humano obrigatÃ³rio**
- Equipe: Felipe (backend/Go/DB/APIs) + Cristian (frontend/Next.js)
- Unidade de negÃ³cio: **5 unidades** Cia da Vacina

**Regra:** tudo nÃ£o citado pelo usuÃ¡rio e nÃ£o decidido acima estÃ¡ marcado como **SUPOSIÃ‡ÃƒO**.

---

# ETAPA 1 â€” COMPREENSÃƒO

## Objetivos de negÃ³cio
1. Centralizar atendimento WhatsApp das 5 unidades em uma Ãºnica plataforma.
2. Padronizar respostas com base nos POPs da Cia da Vacina.
3. Acelerar o primeiro contato via IA de triagem, sem perder humanizaÃ§Ã£o.
4. Dar visibilidade do funil comercial (etapa do cliente).
5. Reduzir passividade pÃ³s-atendimento (follow-up de â€œnÃ£o fechadoâ€ / â€œaguardando fechamentoâ€).
6. Medir produtividade e conversÃ£o por unidade.
7. Aumentar taxa de conversÃ£o dos atendimentos.

## Problemas que o sistema resolve
1. Conversas dispersas no WhatsApp (sem controle central).
2. Atendimento nÃ£o padronizado entre profissionais/unidades.
3. Falta de visÃ£o de status comercial do lead.
4. Perda de oportunidades sem follow-up.
5. Dificuldade de comparar desempenho entre unidades.
6. Retrabalho e atraso na identificaÃ§Ã£o da necessidade do cliente.
7. AssunÃ§Ã£o manual â€œno escuroâ€ sem histÃ³rico consolidado.

## Atores do sistema
- **Cliente** (usuÃ¡rio externo no WhatsApp)
- **IA de Triagem** (bot inicial)
- **Atendente** (responde e move pipeline)
- **Supervisor** (fila, redistribuiÃ§Ã£o, qualidade)
- **Gerente de Unidade** (visÃ£o da unidade + conversÃ£o)
- **Administrador** (usuÃ¡rios, unidades, configs, POPs, integraÃ§Ãµes)
- **Sistema** (webhooks, jobs de follow-up, auditoria)

### SUPOSIÃ‡ÃƒO (papÃ©is)
Roles RBAC acima; Stakeholder nÃ£o listou hierarquia â€” modelado para operaÃ§Ã£o multiunidade tÃ­pica.

## Requisitos Funcionais

| ID | Nome | DescriÃ§Ã£o | Prioridade | DependÃªncias |
|---|---|---|---|---|
| RF-001 | AutenticaÃ§Ã£o | Login seguro com sessÃ£o/JWT e logout | Must | â€” |
| RF-002 | GestÃ£o de usuÃ¡rios | CRUD usuÃ¡rios, logout forÃ§ado, ativar/desativar | Must | RF-001 |
| RF-003 | PermissÃµes RBAC | PapÃ©is: Admin, Gerente, Supervisor, Atendente | Must | RF-002 |
| RF-004 | Unidades | Cadastro/gestÃ£o das 5 unidades | Must | RF-002 |
| RF-005 | VÃ­nculo usuÃ¡rioÃ—unidade | UsuÃ¡rio opera em 1+ unidades (escopo de dados) | Must | RF-003, RF-004 |
| RF-006 | Inbox WhatsApp | Lista conversas ativas/histÃ³rico por unidade/fila | Must | RF-010, RF-005 |
| RF-007 | Thread de mensagens | Enviar/receber texto; ordenaÃ§Ã£o cronolÃ³gica; status entrega | Must | RF-006 |
| RF-008 | Webhook WhatsApp | Receber eventos Meta Cloud API (mensagem, status) | Must | â€” |
| RF-009 | Envio WhatsApp | Enviar mensagens via Graph API com sessÃ£o/template | Must | RF-008 |
| RF-010 | Contatos/Clientes | Identificar contato por telefone; enriquecer perfil | Must | RF-008 |
| RF-011 | IA triagem | SaudaÃ§Ã£o + classificaÃ§Ã£o de intenÃ§Ã£o + resumo + roteamento | Must | RF-007, RF-012 |
| RF-012 | Handoff humano | Atendente assume; IA para; histÃ³rico preservado | Must | RF-011, RF-006 |
| RF-013 | AtribuiÃ§Ã£o | Auto/manual assign para atendente/unidade | Must | RF-005, RF-012 |
| RF-014 | Pipeline comercial | Etapas: Em atendimento; Em negociaÃ§Ã£o; Aguardando fechamento; Fechado; NÃ£o fechado | Must | RF-006 |
| RF-015 | TransiÃ§Ã£o de etapa | Mover conversa/oportunidade com motivo (obrigatÃ³rio em NÃ£o fechado) | Must | RF-014 |
| RF-016 | Motivo de nÃ£o conversÃ£o | CatÃ¡logo + texto livre; base para estratÃ©gia | Must | RF-015 |
| RF-017 | Follow-up | Lista â€œretomarâ€ (Aguardando fechamento / NÃ£o fechado) + lembretes | Must | RF-014, RF-016 |
| RF-018 | POPs / scripts | Biblioteca de procedimentos e respostas padrÃ£o por intenÃ§Ã£o | Must | RF-003 |
| RF-019 | Inserir POP na conversa | Atendente aplica snippet/POP no composer | Should | RF-018, RF-007 |
| RF-020 | Dashboard unidade | Atendimentos abertos, por etapa, conversÃ£o, SLA simples | Must | RF-014 |
| RF-021 | Dashboard consolidado | VisÃ£o Admin/Gerente das 5 unidades | Must | RF-020 |
| RF-022 | Produtividade | Mensagens/atendimentos por atendente/unidade (perÃ­odo) | Should | RF-007, RF-013 |
| RF-023 | Auditoria | Log de aÃ§Ãµes crÃ­ticas (handoff, etapa, assign, envio) | Must | RF-001 |
| RF-024 | ConfiguraÃ§Ãµes | Tokens Meta, WABA, nÃºmero, flags IA, horÃ¡rios | Must | RF-003 |
| RF-025 | Busca | Buscar conversas/contatos por nome/telefone | Should | RF-006, RF-010 |
| RF-026 | NotificaÃ§Ãµes in-app | Novo msg / menÃ§Ã£o / lembrete follow-up | Should | RF-007 |
| RF-027 | Templates WhatsApp | Envio fora da janela 24h via templates aprovados | Must | RF-009 |
| RF-028 | Encerrar conversa | Resolver/arquivar mantendo histÃ³rico | Should | RF-014 |

### Fora do MVP (explicitamente)
- Instagram/Facebook/outros canais
- IA conversacional contÃ­nua
- Discador, e-mail, SMS
- BI avanÃ§ado / exportaÃ§Ãµes complexas

## Requisitos NÃ£o Funcionais
- **Performance:** Inbox inicial < 2s (p95) com 5 unidades; envio msg < 1s API local (excluindo Meta).
- **SeguranÃ§a:** HTTPS, secrets em env/vault, least privilege, isolamento por unidade.
- **Escalabilidade:** 5 unidades, dezenas de atendentes; crescimento orgÃ¢nico sem redesign (1 app Go + Postgres).
- **Disponibilidade:** alvo **99%** comercial; webhook retry da Meta tratado idempotentemente.
- **Auditoria:** aÃ§Ãµes sensÃ­veis imutÃ¡veis (append-only).
- **LGPD:** base legal atendimento; retenÃ§Ã£o configurÃ¡vel; acesso por papel; nÃ£o treinar modelo externo com PII sem consentimento â€” **SUPOSIÃ‡ÃƒO:** polÃ­tica de retenÃ§Ã£o 24 meses.
- **Observabilidade:** logs estruturados + health + mÃ©tricas bÃ¡sicas (requests, errors, webhook lag).
- **Logs:** correlaÃ§Ã£o `request_id` / `conversation_id`.
- **Backup:** Postgres diÃ¡rio + retenÃ§Ã£o 7â€“30 dias â€” **SUPOSIÃ‡ÃƒO** hosting Cloud (ex.: Managed Postgres).

## RestriÃ§Ãµes
1. Apenas 2 desenvolvedores (Felipe + Cristian).
2. Stack: Go + Next.js + React + TypeScript.
3. Canal MVP = WhatsApp Meta Cloud API apenas.
4. IA = triagem + handoff (nÃ£o substitui atendente).
5. Evitar overengineering (sem microserviÃ§os, sem Kafka no MVP).
6. POPs da Cia da Vacina devem orientar atendimento (conteÃºdo a fornecer pelo negÃ³cio).

## SuposiÃ§Ãµes
1. **SUPOSIÃ‡ÃƒO:** 1 nÃºmero WhatsApp Business por unidade OU um hub com roteamento â€” *default adoptado:* **1 WABA com N nÃºmeros (1 por unidade)** se disponÃ­vel; senÃ£o 1 nÃºmero + tag de unidade na triagem.
2. **SUPOSIÃ‡ÃƒO:** AutenticaÃ§Ã£o email/senha (sem SSO corporativo).
3. **SUPOSIÃ‡ÃƒO:** LLM via API (OpenAI/Azure OpenAI/Anthropic) com prompts internos; fallback para menu de intenÃ§Ãµes se LLM falhar.
4. **SUPOSIÃ‡ÃƒO:** Tempo real via **SSE** (mais simples que WebSocket) para inbox.
5. **SUPOSIÃ‡ÃƒO:** Storage de mÃ­dia WhatsApp em object storage (S3-compatible) sÃ³ se mÃ­dia for Must; **MVP texto-first**, mÃ­dia fase V1.1.
6. **SUPOSIÃ‡ÃƒO:** HorÃ¡rio comercial brasileiro (America/Sao_Paulo).
7. **SUPOSIÃ‡ÃƒO:** Deploy Ãºnico (API + web) em cloud com Postgres gerenciado.

## DÃºvidas em aberto (nÃ£o bloqueiam o plano; mudam estimativa)
1. 1 nÃºmero por unidade ou nÃºmero Ãºnico?
2. Qual provedor LLM e se hÃ¡ DPA/LGPD jÃ¡ aprovado?
3. ConteÃºdo oficial dos POPs (quando disponÃ­vel)?
4. SLA de primeira resposta alvo?
5. Precisa integraÃ§Ã£o com agenda/prontuÃ¡rio/sistema de vacinas existente?
6. Quem define catÃ¡logo de motivos de nÃ£o conversÃ£o?
7. Volume estimado msgs/dia?
8. Hospedagem preferida (AWS, GCP, Azure, VPS)?

---

# ETAPA 2 â€” ENGENHARIA REVERSA (MÃ“DULOS)

| MÃ³dulo | Objetivo | Funcionalidades | DependÃªncias | Prioridade | Impacto |
|---|---|---|---|---|---|
| Auth | Acesso seguro | Login, JWT, refresh, logout | â€” | P0 | Bloqueia tudo |
| Users & RBAC | Controle acesso | CRUD user, roles, escopo unidade | Auth | P0 | SeguranÃ§a multiunidade |
| Units | Multiunidade | CRUD 5 unidades | Auth | P0 | Isolamento de dados |
| WhatsApp Gateway | Entrada/saÃ­da msgs | Webhook, send, templates, statuses | Config | P0 | CoraÃ§Ã£o do produto |
| Contacts | Identidade do cliente | Upsert por phone, perfil | Gateway | P0 | HistÃ³rico Ãºnico |
| Conversations/Inbox | Atendimento | Lista, thread, assign, status | Contacts | P0 | OperaÃ§Ã£o diÃ¡ria |
| AI Triage | Primeiro contato inteligente | Prompt POP-aware, intent, roteamento | Conversations | P0 | Agilidade + padrÃ£o |
| Handoff | HumanizaÃ§Ã£o | Assume, pause IA, ownership | AI + Inbox | P0 | Qualidade |
| Pipeline CRM | Funil comercial | Etapas, motivos, histÃ³rico | Conversations | P0 | ConversÃ£o |
| Follow-up | Retomada | Fila, lembretes | Pipeline | P1 | Reduz passividade |
| POPs | PadronizaÃ§Ã£o | Cadastro, busca, insert | Auth | P1 | Qualidade |
| Dashboard | GestÃ£o | KPIs unidade/consolidado | Pipeline | P1 | Visibilidade gerencial |
| Audit | Compliance | Event log | Auth | P1 | LGPD/rastreio |
| Config | OperaÃ§Ã£o | Meta tokens, IA flags | Auth | P0 | Go-live |
| Notifications | UX tempo real | SSE eventos | Inbox | P1 | Agilidade |

---

# ETAPA 3 â€” USER STORIES (principais)

### US-01 Login
**Como** Atendente **quero** autenticar **para** acessar o inbox com seguranÃ§a.  
**ACEITE:** credenciais vÃ¡lidas â†’ token; invÃ¡lidas â†’ erro; usuÃ¡rio inativo bloqueado.  
**Fluxo:** form â†’ API login â†’ cookie/token â†’ redirect `/inbox`.  
**Alt:** senha errada; usuÃ¡rio desativado.  
**Erros:** 401/403.  
**RN:** senha hashed (bcrypt/argon2); sessÃ£o com expiraÃ§Ã£o.

### US-02 Receber mensagem WhatsApp
**Como** Sistema **quero** ingressar mensagem Meta **para** abrir/atualizar conversa.  
**ACEITE:** webhook verificado; idempotÃªncia por `wamid`; cria contato/conversa; dispara IA se estado=triagem.  
**RN:** isolamento por nÃºmero/unidade.

### US-03 Triagem IA
**Como** Cliente **quero** ser saudado e direcionado **para** falar com quem resolve.  
**ACEITE:** IA responde saudaÃ§Ã£o; classifica intenÃ§Ã£o (ex.: agendar, preÃ§os, dÃºvidas, reclamaÃ§Ã£o â€” **SUPOSIÃ‡ÃƒO catÃ¡logo**); sugere unidade; cria resumo; coloca em fila humana.  
**Alt:** falha LLM â†’ mensagem fallback + fila humano.  
**RN:** mÃ¡x N turnos IA (**SUPOSIÃ‡ÃƒO:** 3) ou handoff automÃ¡tico se intent claro / pedido humano.

### US-04 Assumir conversa
**Como** Atendente **quero** assumir **para** continuar atendimento humanizado.  
**ACEITE:** ownership muda; IA desligada; evento auditado; POP sugerido por intent.

### US-05 Enviar mensagem
**Como** Atendente **quero** responder no WhatsApp **para** avanÃ§ar a venda.  
**ACEITE:** dentro 24h free-form; fora 24h sÃ³ template; status sent/delivered/read refletidos.

### US-06 Mover pipeline
**Como** Atendente **quero** mudar etapa **para** refletir negociaÃ§Ã£o.  
**ACEITE:** etapas vÃ¡lidas; â€œNÃ£o fechadoâ€ exige motivo; histÃ³rico de transiÃ§Ã£o.

### US-07 Follow-up
**Como** Supervisor **quero** lista de clientes em aberto/nÃ£o fechado **para** retomar e converter.  
**ACEITE:** filtros por unidade/etapa/motivo; registrar tentativa de retorno.

### US-08 Dashboard
**Como** Gerente **quero** ver desempenho da unidade **para** gerir conversÃ£o.  
**ACEITE:** contadores por etapa; % fechado; atendimentos abertos; perÃ­odo filtrÃ¡vel.

### US-09 POPs
**Como** Atendente **quero** inserir script POP **para** padronizar.  
**ACEITE:** busca por tag/intent; insert no composer editÃ¡vel antes do envio.

### US-10 Admin config Meta
**Como** Admin **quero** configurar WABA/token/nÃºmero **para** operaÃ§Ã£o.  
**ACEITE:** secrets mascarados; teste de webhook; e2e send test.

*(Demais stories derivadas no backlog â€” Etapa 9)*

---

# ETAPA 4 â€” CASOS DE USO (amostra canÃ´nica + padrÃ£o)

### CU-03 Triagem IA
- **PrÃ©:** webhook ok; conversa `state=ai_triage`; prompt/POP carregados.
- **Principal:** msg cliente â†’ IA â†’ reply WhatsApp â†’ atualiza intent/resumo â†’ se critÃ©rios handoff â†’ fila.
- **Alt:** timeout LLM; conteÃºdo ofensivo; cliente pede humano.
- **PÃ³s:** conversa com intent; aguardando atendente OU ainda em triagem.

### CU-04 Handoff
- **PrÃ©:** conversa em fila/triage; atendente autenticado com escopo unidade.
- **Principal:** claim â†’ ownership â†’ stop IA â†’ notifica UI.
- **Alt:** jÃ¡ atribuÃ­da a outro (conflito â†’ 409).
- **PÃ³s:** `owner_id` set; `mode=human`.

### CU-06 Pipeline
- **PrÃ©:** conversa sob ownership ou papel supervisor.
- **Principal:** seleciona etapa â†’ valida â†’ pede motivo se NÃ£o fechado â†’ grava transiÃ§Ã£o.
- **PÃ³s:** etapa atual + audit trail.

*(Aplicar o mesmo esqueleto PrÃ©/Principal/Alt/PÃ³s a US-01â€¦US-10 na implementaÃ§Ã£o.)*

---

# ETAPA 5 â€” MODELO DE DADOS

**Entidades principais (Postgres)**

- **users:** id, email, password_hash, name, role(`admin|manager|supervisor|agent`), active, created_at
- **units:** id, name, code, timezone, active
- **user_units:** user_id, unit_id (PK composta)
- **whatsapp_numbers:** id, unit_id, phone_number_id, display_phone, waba_id, active
- **contacts:** id, wa_id, phone_e164, name, unit_id?, created_at; **UNIQUE(wa_id)**
- **conversations:** id, contact_id, unit_id, status(`open|pending|resolved`), mode(`ai_triage|human`), owner_id?, pipeline_stage, intent?, ai_summary?, last_message_at, window_expires_at
- **messages:** id, conversation_id, direction(`in|out`), sender_type(`contact|agent|ai|system`), body, wa_message_id UNIQUE, status(`accepted|sent|delivered|read|failed`), created_at
- **pipeline_transitions:** id, conversation_id, from_stage, to_stage, reason_code?, reason_text?, by_user_id, created_at
- **loss_reasons:** id, code, label, active
- **pops:** id, title, body, intent_tags[], unit_id nullable, active
- **followups:** id, conversation_id, due_at, status(`open|done|canceled`), note, created_by
- **ai_runs:** id, conversation_id, model, prompt_version, input_ref, output_json, latency_ms, error?
- **audit_logs:** id, actor_user_id?, action, entity_type, entity_id, payload_json, ip?, created_at
- **message_templates:** id, name, language, status, body_preview, meta_template_name
- **configs:** key, value_encrypted?, updated_by, updated_at

**Ãndices:** `conversations(unit_id, status, last_message_at DESC)`; `messages(conversation_id, created_at)`; `contacts(phone_e164)`; `followups(status, due_at)`; `audit_logs(created_at)`.

**RestriÃ§Ãµes:** FKs; enum stages exatamente as 5 etapas; motivo obrigatÃ³rio se `to_stage=nao_fechado`.

```mermaid
flowchart LR
  units --> whatsapp_numbers
  users --> user_units
  units --> user_units
  contacts --> conversations
  units --> conversations
  users --> conversations
  conversations --> messages
  conversations --> pipeline_transitions
  conversations --> followups
  pops -.-> conversations
```

---

# ETAPA 6 â€” APIs (contrato inicial)

Base: `/api/v1` â€” Auth Bearer JWT â€” erros `{code,message,details}`

| MÃ©todo | Endpoint | PermissÃ£o | Notas |
|---|---|---|---|
| POST | `/auth/login` | public | â†’ tokens |
| POST | `/auth/logout` | auth | |
| GET | `/me` | auth | perfil + units |
| CRUD | `/users` | admin | |
| CRUD | `/units` | admin | |
| GET | `/inbox?unit_id&stage&owner` | agent+ | lista conversas |
| GET | `/conversations/{id}` | scoped | detalhe |
| GET | `/conversations/{id}/messages` | scoped | paginado cursor |
| POST | `/conversations/{id}/messages` | owner/supervisor | envia WhatsApp |
| POST | `/conversations/{id}/claim` | agent+ | handoff |
| POST | `/conversations/{id}/assign` | supervisor+ | |
| PATCH | `/conversations/{id}/pipeline` | agent+ | body stage+reason |
| GET/POST | `/followups` | agent+ | |
| CRUD | `/pops` | admin/supervisor write | |
| GET | `/dashboard/summary` | manager+ | KPIs |
| GET | `/sse/events` | auth | tempo real |
| GET/POST | `/webhooks/whatsapp` | Meta | verify + ingress |
| GET/PUT | `/settings/whatsapp` | admin | |
| GET | `/loss-reasons` | auth | |

**Erros HTTP:** 400 validaÃ§Ã£o; 401/403; 404; 409 claim conflito; 422 regra negÃ³cio (fora janela sem template); 429; 502 Meta upstream.

---

# ETAPA 7 â€” ARQUITETURA (simples, 2 devs)

```mermaid
flowchart TB
  WA[Meta WhatsApp Cloud API] -->|webhook| API[Go API monolith]
  Web[Next.js App] -->|REST + SSE| API
  API --> PG[(PostgreSQL)]
  API -->|send messages| WA
  API -->|triage prompts| LLM[LLM Provider]
  API --> Obj[S3 opcional midia V1.1]
```

**DecisÃµes**
- **Go monÃ³lito modular** (package por domÃ­nio): 1 deploy, 1 DB â€” adequado a 2 pessoas.
- **Next.js App Router**: BFF mÃ­nimo; preferir chamar API Go direto (cookies/JWT).
- **Postgres:** fonte da verdade; sem sharding.
- **Cache:** sem Redis no MVP; cache em memÃ³ria sÃ³ para JWKS/config TTL curto se precisar.
- **Mensageria:** **nÃ£o** no MVP â€” webhook processado sÃ­ncrono + fila DB (`jobs` table) se retry simples.
- **Tempo real:** **SSE** (menos ops que WS).
- **Workers:** goroutine/cron no mesmo binÃ¡rio para follow-ups e retries Meta.

**Por quÃª:** menor ops, menos falhas de integraÃ§Ã£o interna, entrega mais rÃ¡pida.

---

# ETAPA 8 â€” ANÃLISE TRIPLA (por mÃ³dulo â€” sÃ­ntese)

Escala: Valor / Complexidade / Tempo / BenefÃ­cio / ManutenÃ§Ã£o / Escala / Risco (0â€“10) â†’ veredito.

| MÃ³dulo | A Negativos | B Positivos | Notas mÃ©dias | Veredito |
|---|---|---|---|---|
| Auth/RBAC/Units | Escopo multiunidade errado = vazamento | Base sÃ³lida reuso | V9 C5 T5 B9 M8 E7 R5 | **Vale Muito** |
| WhatsApp Gateway | API Meta volÃ¡til; templates; 24h window | Canal Ãºnico real do negÃ³cio | V10 C8 T8 B10 M6 E7 R8 | **Vale Muito** (atenÃ§Ã£o risco) |
| Inbox/Msgs | Tempo real + estados | Core do produto | V10 C7 T7 B10 M7 E7 R6 | **Vale Muito** |
| AI Triage | Custo LLM; alucinaÃ§Ã£o; LGPD | Padroniza + agiliza entrada | V8 C7 T6 B8 M6 E7 R7 | **Vale a Pena** (MVP limitado) |
| Handoff | ConcorrÃªncia claim | Humaniza; reduz frustraÃ§Ã£o | V9 C4 T3 B9 M8 E8 R3 | **Vale Muito** |
| Pipeline+Motivos | Disciplina de uso humano | Controle conversÃ£o | V10 C4 T4 B10 M8 E8 R3 | **Vale Muito** |
| Follow-up | Spam/WhatsApp policy | Combate passividade | V8 C4 T4 B8 M7 E7 R4 | **Vale a Pena** |
| POPs | ConteÃºdo depende do negÃ³cio | PadronizaÃ§Ã£o baixa tech | V7 C2 T3 B7 M9 E8 R2 | **Vale a Pena** |
| Dashboard | Pode virar BI infinito | GestÃ£o unidades | V8 C4 T4 B8 M7 E6 R3 | **Vale a Pena** (KPIs mÃ­nimos) |
| Audit/Config | FÃ¡cil descuidar | Compliance + go-live | V7 C3 T3 B7 M8 E7 R3 | **Vale a Pena** |
| NotificaÃ§Ãµes SSE | ConexÃµes idle | Produtividade inbox | V7 C4 T3 B7 M7 E6 R4 | **Vale a Pena** |

**Corte Tech Lead:** sem omnichannel, sem microserviÃ§os, sem IA contÃ­nua, dashboard sÃ³ 5â€“7 KPIs.

---

# ETAPA 9 â€” BACKLOG (tarefas 4â€“12h)

ConvenÃ§Ã£o: **BE** Felipe / **FE** Cristian | Estimativa em horas | P0â†’P2

### FundaÃ§Ã£o
- T-001 BE Setup Go module, config, health, logging estruturado â€” 6h P0
- T-002 BE Postgres schema base (users, units, user_units) + migrate tool â€” 8h P0
- T-003 BE Auth login/JWT/refresh/middleware RBAC â€” 10h P0
- T-004 BE CRUD users/units + assign units â€” 10h P0
- T-005 FE App shell Next.js, auth pages, guard routes â€” 10h P0
- T-006 FE Layout app (nav, unit switcher) + design tokens â€” 8h P0
- T-007 Ambos OpenAPI/contrato v0 + mocks MSW â€” 8h P0

### WhatsApp + DomÃ­nio conversas
- T-010 BE Webhook verify + ingest idempotente â€” 10h P0
- T-011 BE Models contacts/conversations/messages + repos â€” 10h P0
- T-012 BE Send text Graph API + status updates â€” 10h P0
- T-013 BE Templates send (fora janela) â€” 8h P0
- T-014 BE Inbox query filters + pagination â€” 8h P0
- T-015 FE Inbox list + conversation thread UI â€” 12h P0
- T-016 FE Composer send + estados loading/error/window 24h â€” 10h P0
- T-017 BE Settings WhatsApp admin API â€” 6h P0
- T-018 FE Settings WhatsApp (mascarar secrets) â€” 6h P0

### IA + Handoff
- T-020 BE AI triage service (prompt, intent, summary, max turns) â€” 12h P0
- T-021 BE Fallback sem LLM + flags config â€” 6h P0
- T-022 BE Claim/assign/handoff + conflitos 409 â€” 8h P0
- T-023 FE Banner modo IA/humano + botÃ£o Assumir â€” 6h P0
- T-024 FE Painel intent/resumo IA â€” 6h P0

### Pipeline + Follow-up
- T-030 BE Pipeline patch + transitions + loss reasons â€” 8h P0
- T-031 FE Stage selector + motivo modal â€” 8h P0
- T-032 BE Follow-ups CRUD + job due â€” 8h P1
- T-033 FE Fila follow-up + filtros â€” 8h P1

### POPs + Dashboard + Audit + Realtime
- T-040 BE/FE POPs CRUD + insert composer â€” 8h+8h P1
- T-041 BE Dashboard summary SQL â€” 8h P1
- T-042 FE Dashboard unidade + consolidado â€” 10h P1
- T-043 BE Audit log writer + list admin â€” 6h P1
- T-044 BE SSE events + FE subscribe â€” 8h+8h P1
- T-045 BE/FE Busca contatos/conversas â€” 6h+6h P2
- T-046 QA e2e checklist WhatsApp sandbox â€” 8h P0 (compartilhado)
- T-047 Hardening LGPD/retention + backups doc â€” 6h P1

---

# ETAPA 10 â€” DIVISÃƒO DA EQUIPE

## Backend (Felipe) â€” pacotes
- **Models/Repos:** users, units, contacts, conversations, messages, pipeline, pops, followups, audit, configs
- **Repositories:** SQLC ou pgx
- **Services/Use cases:** Auth, WhatsAppIngress, WhatsAppEgress, AITriage, ClaimConversation, UpdatePipeline, Dashboard, FollowUpScheduler
- **Middlewares:** auth, rbac, unit-scope, request-id, recover
- **Endpoints:** lista Etapa 6
- **ValidaÃ§Ãµes:** go-playground/validator; regras janela 24h; motivo obrigatÃ³rio
- **IntegraÃ§Ãµes:** Meta Graph, LLM
- **Workers:** retry send; follow-up due notifier
- **Testes:** unit services + integration webhook idempotency
- **Docs:** OpenAPI
- **Tempo individual agregado:** ~**160â€“190h**

## Frontend (Cristian)
- **PÃ¡ginas:** login, inbox, conversation, follow-ups, dashboard, pops, users/admin, settings
- **Layouts:** auth / app / admin
- **Componentes:** ConversationList, Thread, Composer, StageBadge, ClaimButton, PopPicker, KpiCards, UnitSwitcher
- **Hooks:** useInbox, useConversation, useSSE, useAuth
- **State:** server-state TanStack Query; UI state mÃ­nimo (Zustand opcional sÃ³ filtros)
- **APIs:** client tipado do OpenAPI
- **Forms/validaÃ§Ã£o:** React Hook Form + Zod
- **Loading/erros/empty/responsive:** obrigatÃ³rio em inbox/dashboard
- **Testes:** Testing Library fluxos claim/send/pipeline
- **Tempo individual agregado:** ~**130â€“160h**

---

# ETAPA 11 â€” PARALELISMO

- **Front pode comeÃ§ar?** Sim, **Dia 1** com mocks apÃ³s T-007.
- **Mocks?** Sim (MSW) para inbox, messages, claim, pipeline, dashboard.
- **Contratos primeiros:** Auth, Inbox list, Messages, Claim, Pipeline patch, SSE event shape, WhatsApp settings.
- **Endpoints que desbloqueiam FE:** `login`, `GET /inbox`, `GET messages`, `POST messages`, `POST claim`, `PATCH pipeline`.
- **Paralelo:** FE shell+inbox mock || BE schema+auth+webhook.
- **Bloqueios:** go-live WhatsApp real bloqueia validaÃ§Ã£o e2e; IA precisa de chave LLM; POPs conteÃºdo negÃ³cio.

---

# ETAPA 12 â€” ÃRVORE DE DEPENDÃŠNCIAS

```mermaid
flowchart TB
  Auth --> UsersRBAC
  Auth --> Units
  UsersRBAC --> UnitScope
  Units --> WhatsAppConfig
  WhatsAppConfig --> WebhookGateway
  WebhookGateway --> Contacts
  Contacts --> Conversations
  Conversations --> Messages
  Messages --> AITriage
  AITriage --> Handoff
  Handoff --> Pipeline
  Pipeline --> FollowUp
  Conversations --> POPs
  Pipeline --> Dashboard
  Auth --> Audit
  Conversations --> SSE
```

---

# ETAPA 13 â€” ROADMAP (ordem anti-retrabalho)

1. Contratos OpenAPI + Auth + Units + RBAC  
2. WhatsApp webhook/send (sandbox) + Contacts/Conversations/Messages  
3. Inbox FE conectado  
4. Claim/Handoff  
5. Pipeline + motivos  
6. IA triagem (com kill-switch)  
7. Templates 24h  
8. SSE  
9. Follow-up  
10. POPs  
11. Dashboard  
12. Audit + hardening + go-live checklist  

**Por quÃª:** UI e domÃ­nio de conversa estabilizam cedo; IA/dashboard depois para nÃ£o invalidar contratos.

---

# ETAPA 14 â€” CRONOGRAMA

### Tabela resumo (horas)

| Funcionalidade | BE | FE | Deps | Complex. | h BE | h FE | Paralelo | Total calendar* |
|---|---|---|---|---|---|---|---|---|
| FundaÃ§Ã£o Auth/Units | Felipe | Cristian | â€” | M | 34 | 26 | Sim | ~1.5 sem |
| WhatsApp+Inbox | Felipe | Cristian | FundaÃ§Ã£o | A | 54 | 28 | Parcial | ~2.5 sem |
| Handoff+Pipeline | Felipe | Cristian | Inbox | M | 16 | 14 | Sim | ~1 sem |
| IA Triagem | Felipe | Cristian | Msg+Handoff | A | 18 | 12 | Parcial | ~1 sem |
| Templates+Settings | Felipe | Cristian | WhatsApp | M | 14 | 6 | Sim | ~0.5 sem |
| Follow-up | Felipe | Cristian | Pipeline | B | 8 | 8 | Sim | ~0.5 sem |
| POPs | Felipe | Cristian | Inbox | B | 8 | 8 | Sim | ~0.5 sem |
| Dashboard+SSE+Audit | Felipe | Cristian | Pipeline | M | 22 | 26 | Sim | ~1.5 sem |
| QA/Hardening | ambos | ambos | tudo P0 | M | 10 | 10 | Sim | ~0.5 sem |

\*Total calendar assume ~6h produtivas/dia/dev.

**Agregado:** BE ~**184h** | FE ~**138h** | Bruto ~**322h**  
**Paralelo 2 devs:** ~**190â€“210h** de calendÃ¡rio de equipe â‰ˆ **8â€“10 semanas** (5 dias/sem Ã— ~6h) com folga de risco Meta/LLM.

### Por semanas (MVP)
- S1: contratos, auth, units, shell FE  
- S2â€“S3: webhook, messages, inbox real  
- S4: claim + pipeline + motivos  
- S5: IA triagem + fallback  
- S6: templates + settings + SSE  
- S7: follow-up + POPs  
- S8: dashboard + audit + hardening  
- S9â€“S10: buffer integraÃ§Ã£o Meta, UAT, ajustes POP/negÃ³cio  

### Sprints (2 semanas)
- Sprint 1: FundaÃ§Ã£o + inÃ­cio WhatsApp  
- Sprint 2: Inbox + send/receive estÃ¡veis  
- Sprint 3: Handoff + Pipeline + IA  
- Sprint 4: Templates + Follow-up + POPs + Dashboard  
- Sprint 5 (buffer): UAT + go-live  

### Marcos
- M1: Login + multiunidade  
- M2: WhatsApp e2e sandbox  
- M3: OperaÃ§Ã£o humana completa (inbox+pipeline)  
- M4: IA triagem com kill-switch  
- M5: GestÃ£o (dashboard+follow-up)  
- M6: ProduÃ§Ã£o

---

# ETAPA 15 â€” MVP

**IndispensÃ¡vel (MVP)**  
Auth/RBAC/Units, WhatsApp in/out, Inbox, IA triagem+handoff, Pipeline+motivos, Templates 24h, Dashboard mÃ­nimo, Audit bÃ¡sico, Settings Meta.

**Pode esperar (pÃ³s-MVP imediato)**  
MÃ­dia rica, busca avanÃ§ada, notificaÃ§Ãµes push mobile, produtividade detalhada.

**V2**  
Instagram/Messenger, IA assistindo sugestÃµes durante humano, automaÃ§Ãµes de campanha, BI export, SSO, app mobile, CSAT.

**Maior valor rÃ¡pido**  
Inbox+pipeline+motivos (mesmo sem IA) jÃ¡ elimina passividade; IA Ã© acelerador do primeiro contato.

---

# ETAPA 16 â€” ANÃLISE CRÃTICA

- **Duplicados:** â€œcontrole conversasâ€ e â€œpipelineâ€ â€” unificar em Conversation+stage (sem entidade Deal separada no MVP) â†’ **simplicidade**.
- **DesnecessÃ¡rio agora:** omnichannel, microserviÃ§os, Redis, Kafka, app nativo.
- **InconsistÃªncia potencial:** 1 nÃºmero vs 5 â€” precisa fechar antes do go-live.
- **Riscos:** aprovaÃ§Ã£o templates Meta; polÃ­tica 24h; custo LLM; adoÃ§Ã£o do time no preenchimento de etapas.
- **Gargalo:** Felipe no Gateway Meta (caminho crÃ­tico).
- **SimplificaÃ§Ã£o:** 1 tabela `conversations.pipeline_stage` em vez de CRM deals paralelo.
- **Reuso:** mesmo componente de filtros unidade no Inbox/Dashboard/Follow-up.

---

# ETAPA 17 â€” TECH LEAD REVIEW

- **Removeria:** canais sociais; IA contÃ­nua; BI rico; microserviÃ§os; Redis obrigatÃ³rio.
- **Simplificaria:** Deal=Conversation; SSE nÃ£o WS; jobs em tabela Postgres.
- **Faria diferente:** fechar contrato OpenAPI no dia 1; WhatsApp sandbox na semana 2; IA atrÃ¡s de feature flag.
- **Overengineering:** multi-tenant genÃ©rico SaaS â€” aqui Ã© single-tenant Cia da Vacina + 5 units.
- **Risco atraso:** Meta (templates/webhook), conteÃºdo POP, indefiniÃ§Ã£o de nÃºmeros.
- **Felipe bloqueado:** credenciais WABA; provider LLM; regras de roteamento unidade.
- **Cristian bloqueado:** contrato inbox/messages; decision SSE payload.
- **-30% tempo:** cortar dashboard elaborado + POPs CRUD rico (comeÃ§am markdown seed) + adiar follow-up job para lista manual + IA fase logo apÃ³s handoff humano estÃ¡vel.
- **Qualidade sem alongar:** contract tests; checklist e2e WhatsApp; feature flags; code review cruzado sÃ³ em Gateway/Auth/Pipeline.

---

# ETAPA 18 â€” CONCLUSÃƒO

1. **Resumo executivo:** CRM interno para centralizar WhatsApp das 5 unidades, triagem por IA, handoff humano, funil comercial e gestÃ£o de conversÃ£o â€” stack simples Go + Next.js + Postgres.  
2. **Escopo final MVP:** auth multiunidade, WhatsApp Meta, inbox, IA triagem, handoff, pipeline 5 etapas + motivos, templates, follow-up bÃ¡sico, POPs, dashboard mÃ­nimo, audit.  
3. **Arquitetura:** monÃ³lito Go + Next.js + Postgres + SSE + LLM API; sem broker.  
4. **Backlog:** tarefas T-001â€¦T-047 (4â€“12h).  
5. **Roadmap:** fundaÃ§Ã£o â†’ WhatsApp â†’ inbox â†’ handoff/pipeline â†’ IA â†’ templates/SSE â†’ follow-up/POPs â†’ dashboard â†’ go-live.  
6. **Cronograma:** 8â€“10 semanas com buffer.  
7. **Estimativa BE:** ~184h  
8. **Estimativa FE:** ~138h  
9. **Estimativa total esforÃ§o:** ~322h  
10. **Em dias (1 pessoa):** ~54 dias Ãºteis @6h  
11. **Em paralelo (2 pessoas):** ~**9 semanas** (+1 buffer)  
12. **Caminho crÃ­tico:** Meta WhatsApp Gateway â†’ Messages â†’ Inbox â†’ Handoff â†’ Pipeline â†’ (IA) â†’ Go-live  
13. **Riscos:** Meta templates/janela 24h; adoÃ§Ã£o pipeline; LLM/LGPD; 1 vs N nÃºmeros  
14. **MitigaÃ§Ã£o:** sandbox cedo; feature flags; motivos obrigatÃ³rios UX; fallback sem IA; workshop POPs  
15. **MVP:** operaÃ§Ã£o humana + funil (+ IA triagem)  
16. **V2:** omnichannel, IA copiloto, automaÃ§Ãµes, BI  
17. **ConfianÃ§a estimativa:** **65%**  

**Fatores que alteram forte a estimativa:** atraso WABA/templates; mÃ­dia obrigatÃ³ria no MVP; SSO; integraÃ§Ã£o com sistema clÃ­nico/agenda; volume muito alto exigindo fila real; mudanÃ§a para omnichannel.

---

## Artefatos sugeridos no repo (apÃ³s aprovaÃ§Ã£o; sem cÃ³digo agora)
- `docs/spec.md` â€” este documento
- `docs/openapi.yaml` â€” contrato
- `docs/data-model.md` â€” entidades
- `docs/backlog.csv` â€” T-IDs
- `docs/adr/0001-simple-monolith.md`

