# Plano de Adaptação — CRM ao Novo Modelo de Cobrança WhatsApp (out/2026)

**Baseado em:** `docs/WHATSAPP-2026-GAP-ANALYSIS.md` (relatório anterior) + decisões tomadas com o Felipe em 2026-08-04 + pesquisa de padrões reais da API da Meta (feita agora, conforme pedido, mesmo sem conta de API ativa ainda).
**Não é implementação** — é o plano que orienta a implementação quando ela começar. Schema é descrito em prosa, não em SQL/Go, pelo mesmo motivo do relatório anterior: alinhamento antes de código.

---

## 1. Decisões registradas

| # | Decisão | Escolha |
|---|---|---|
| D1 | IA de triagem vs. Meta Business Agent | **Substituir** a IA de triagem própria (`internal/module/triage`, OpenAI) pelo Meta Business Agent |
| D2 | Client HTTP real da Meta | **Manter o mock** por enquanto (sem conta de API ainda) — mas pesquisar os padrões reais agora pra já deixar o sistema alinhado (feito na seção 2) |
| D3 | Regras determinísticas (phone_gate/handoff) sobre a resposta da IA | **Não** — o Business Agent decide sozinho quando escalar pra humano (configurado do lado da Meta, não mais no backend) |
| D4 | Escopo do Business Agent por canal (WhatsApp vs. IG/FB) | **Em aberto** — resolvido parcialmente pela pesquisa (seção 2.1): Business Agent é onboardado por `phone_number_id`, ou seja, é um produto **específico de WhatsApp**. Recomendação: IG/FB mantêm a IA própria já construída, por não haver equivalente nativo da Meta pra esses canais — **precisa confirmação final do Felipe antes de remover código de IG/FB** |
| D5 | Catálogo de templates | **Próprio no CRM** — tabela + CRUD + tela, backend é fonte de verdade (categoria, variáveis, status de aprovação) |
| D6 | Janela de 24h por categoria | **Entra agora**, junto com o resto do schema de custo (não fica pra depois) |
| D7 | CTWA (Click-to-WhatsApp) | **Entra agora** — parse do `referral`/objeto de anúncio no webhook + entidade nova de campanha de anúncio, incluindo resolver a colisão de nome com `ai_campaigns` |
| D8 | Granularidade do custo | **Instrumentação + dashboard + alertas automáticos** (não só instrumentação) |
| D9 | Canal de alerta | **Dentro do CRM, mas persistente** — não só broadcast SSE pra quem está com a tela aberta; precisa sobreviver e aparecer quando o usuário loga depois |
| D10 | WhatsApp Flows / botões interativos | **Entram no plano**, pesquisados na seção 2.3 antes de desenhar o schema |
| D11 | Prazo | **Sem data rígida** — backlog priorizado por dependência técnica, não por calendário de 1º/out |

---

## 2. Padrões reais da API pesquisados

Pesquisa feita agora (WebSearch/WebFetch, ago/2026) porque o Felipe ainda não tem conta de API ativa pra validar na prática — mesma ressalva epistêmica que já existe em `backend/ARCHITECTURE.md` §5 pros clients Meta atuais: **isto é o que a documentação pública e agregadores confiáveis descrevem hoje, não algo validado contra tráfego real**. Quando a conta existir, revalidar campo a campo antes de codar em cima disso, exatamente como já é prática neste projeto.

### 2.1 Meta Business Agent

- Onboarding é uma chamada `POST` num endpoint de agente **com `phone_number_id` como parâmetro de rota** — confirma que é um produto por número de WhatsApp, não multicanal. Isso sustenta a recomendação da D4.
- Corpo da criação inclui: nome do agente, instruções iniciais, e autorização de fontes de dados (texto/FAQ, documentos/PDF, URLs de site, catálogo de produto).
- Escalação pra humano é configurável **programaticamente**, com "tópicos gatilho" por palavra-chave, intenção ou sentimento — ou seja, existe uma API de configuração, não é só um painel visual (bom, dá pra automatizar/versionar a config via `PUT /settings/meta` do próprio CRM, mantendo o padrão já usado pra Instagram/Facebook).
- Webhooks notificam: novas mensagens, **gatilhos de transferência** (handoff), erros do agente, atualizações de desempenho — o gatilho de transferência é o evento que o CRM vai precisar escutar pra saber quando uma conversa "virou humana" (equivalente ao que `mode: human` já representa hoje).
- **Não confirmado na pesquisa:** o formato exato de billing/consumo de tokens por resposta (nem endpoint, nem campo de webhook). Isso é uma lacuna real de documentação pública, não só falta de tempo de pesquisa — precisa ser resolvido com a conta de API ativa ou com suporte Meta antes de desenhar a coluna de custo do Business Agent.

Fontes: [Meta Business Agent API: The Technical Onboarding Guide for Developers](https://www.memacon.com/meta-business-agent-api-the-technical-onboarding-guide-for-developers/), [Webhooks - Meta for Developers](https://developers.facebook.com/documentation/business-messaging/whatsapp/webhooks/overview)

### 2.2 Categoria de cobrança no webhook de status

- Desde 1º de julho de 2025 a Meta já cobra por mensagem (não mais por conversa) — mudança anterior à de outubro/2026 descrita no guia, mas que já estabelece o mecanismo de reporte que a mudança de out/2026 usa.
- O webhook de **status** de mensagem enviada (`statuses`) inclui um objeto `pricing` com três campos: `billable` (bool), `pricing_model` (ex.: `"CBP"`), `category` (o que efetivamente foi cobrado — ex.: `"marketing"`, `"utility"`, `"authentication"`, `"service"`).
- Existe também `pricing_type`, com valores como `regular` (cobrado conforme `category`), `free_customer_service` (grátis por ser utilidade dentro da janela, ou texto livre dentro da janela de atendimento) e `free_entry_point` (grátis por CTWA, provavelmente).
- **Implicação direta pro plano:** o backend não precisa *calcular* o custo sozinho a partir de regras — a Meta já devolve a categoria de cobrança real no webhook de status de cada mensagem enviada. O "núcleo de cálculo de custo" do guia (seção 7 item 4) é, na prática, **capturar e persistir esse campo do webhook**, e só then mapear categoria → valor com uma tabela de preço local (pro caso de a Meta atrasar/não confirmar o rate card definitivo, que segundo o guia só sai por volta de 1º/set/2026). Isso simplifica bastante o núcleo de custo: não é uma engine de regras complexa, é ingestão de um campo + lookup de preço.

Fontes: [Pricing on the WhatsApp Business Platform](https://developers.facebook.com/documentation/business-messaging/whatsapp/pricing), [Guide to WhatsApp Webhooks](https://hookdeck.com/webhooks/platforms/guide-to-whatsapp-webhooks-features-and-best-practices)

### 2.3 WhatsApp Flows

- Dois tipos: **estáticos** (coleta de dados estruturados, sem necessidade de backend) e **dinâmicos** (interação em tempo real, precisam de troca de dados com backend — mais parecido com "mini site dentro do WhatsApp").
- Só Flows **publicados** podem ser enviados; um Flow é anexado a uma mensagem de template (`interactive.type: "flow"`), com um `flow_token` (identifica a sessão), `flow_id`, CTA e ação de navegação/dados iniciais.
- Resposta do usuário ao completar o Flow chega como mensagem inbound do tipo `interactive`, subtipo `nfm_reply`, contendo um `response_json` (string JSON serializada com os campos preenchidos, chaveados por tela/campo) — é esse payload que o webhook precisa parsear e estruturar antes de entregar pro usecase de conversa/customer.
- **Implicação pro plano:** pra Flows *estáticos* (o caso de uso do guia — nome/idade/vacina/unidade/turno), não é necessário nenhum backend de Flow em si (a Meta hospeda o formulário); o trabalho real do CRM é (a) montar e publicar o Flow no Meta Flow Builder — trabalho de configuração, não de código —, (b) o CRM saber **enviar** a mensagem `interactive.flow` referenciando esse Flow publicado, e (c) o CRM saber **parsear** o `nfm_reply` de volta pra dados estruturados de conversa/triagem. Ou seja, é código de envio + parser de recebimento, não uma reimplementação do formulário.

Fontes: [WhatsApp Flows API - Postman](https://www.postman.com/meta/whatsapp-business-platform/documentation/y5swede/whatsapp-flows-api), [WhatsApp Business Messaging: Create and manage WhatsApp Flows - Infobip](https://www.infobip.com/docs/whatsapp/whatsapp-flows)

### 2.4 Referral / CTWA (`ctwa_clid`)

- Quando a mensagem inbound vem de um clique em anúncio Click-to-WhatsApp, a Meta anexa um objeto de referral **à primeira mensagem recebida** daquela conversa — carrega origem do anúncio (tipo, id, url), texto do anúncio (headline/body), tipo de mídia, e o `ctwa_clid` (o identificador de clique usado pra atribuição/ROI).
- Três condições precisam ser verdadeiras pra esse dado aparecer: campanha CTWA ativa, a mensagem realmente se originou de um toque no anúncio, e atribuição habilitada nas configurações do WhatsApp Business da conta.
- **Ressalva de fonte:** os exemplos encontrados vêm de agregadores terceiros (Whapi.Cloud, Twilio), que às vezes normalizam o payload de forma diferente do JSON cru da Meta (ex.: aninhando sob `context.ad` em vez de um campo `referral` de primeiro nível). O nome exato do campo/aninhamento **precisa ser confirmado contra a documentação oficial `developers.facebook.com/docs/whatsapp` com uma conta de API real** antes de codar o parser — isso é exatamente o tipo de "shape inferido, nunca visto de um app real" que `backend/ARCHITECTURE.md` §8 já registra como ressalva conhecida pros webhooks de engagement, e vai se repetir aqui.

Fontes: [Track Click-to-WhatsApp Ad ROI with ctwa_clid](https://whapi.cloud/blog/track-click-to-whatsapp-ctwa-clid), [New 'Click ID' Callback Parameter - Twilio](https://www.twilio.com/en-us/changelog/new--click-id--callback-parameter-for-inbound-whatsapp-messages-)

---

## 3. O que muda na arquitetura atual

Em relação a `backend/ARCHITECTURE.md`:

- **`internal/module/triage`** deixa de ser "IA de triagem própria via OpenAI síncrona no webhook" pro WhatsApp — vira, na prática, um listener de eventos do Business Agent (recebe o gatilho de transferência, seta `mode: human`, dispara follow-up se aplicável). O código de prompt/`RunTriage`/`pkg/openai` **provavelmente continua existindo só pra Instagram/Facebook** (pendente D4). Isso é uma redução de escopo do módulo pro WhatsApp, não uma reescrita completa.
- **`pkg/meta`** ganha um terceiro tipo de capacidade além de `Sender`/`CommentResponder`: algo como `AgentConfigurer` (configura o Business Agent via API) — mas só quando D2 avançar (hoje fica só documentado/mapeado, sem client real, por decisão explícita do Felipe).
- **`messages`** ganha colunas de cobrança: categoria de pricing (a que veio do webhook de status, não calculada localmente) e custo. Também ganha um novo `kind` pra Flow (`interactive`/`flow`) além dos já existentes.
- **`conversations.window_expires_at`** (coluna única) é substituída por um rastreamento por categoria — precisa de uma tabela nova (ex.: uma linha por conversa+categoria com `expires_at`), já que o guia deixa claro que mais de uma janela pode estar aberta ao mesmo tempo com o mesmo cliente.
- **`ai_campaigns`** — resolver a ambiguidade de nome antes de criar a entidade nova de CTWA ao lado dela. Sugestão (a confirmar com o Felipe, não é decisão automática): renomear `ai_campaigns` → algo como `triage_context_windows` (o que ela realmente é: contexto textual por período pro prompt da IA), liberando o nome "campanha" pra uso real de anúncio/CTWA (`ad_campaigns` ou `ctwa_campaigns`). Isso é uma migration de rename, não de reescrita de dado.
- **Novo:** tabela de templates (catálogo próprio, D5), tabela/mecanismo de alerta persistente (D9 — mais parecido com o padrão de `follow_ups`/`audit_logs` já existentes do que com o hub de SSE, já que precisa sobreviver a login/logout).

---

## 4. Frentes de trabalho

Cada frente lista: o que precisa existir, do que depende, e o que ainda está pendente de confirmação.

### Frente A — Núcleo de custo (base de tudo)
- **O quê:** capturar `pricing.category`/`pricing.billable`/`pricing_model` do webhook de status (mock, por enquanto — D2) e persistir por mensagem; tabela local de rate card (categoria → valor, editável, porque o valor definitivo da Meta só é confirmado ~1º/set/2026 segundo o guia); dashboard de custo (extensão de `GET /dashboard/summary` ou endpoint novo).
- **Depende de:** nada além do schema de `messages` — pode começar já, mesmo em mock, porque o mock pode simular o webhook de status com `pricing` preenchido.
- **Pendente:** nenhuma, tecnicamente — é a frente mais desbloqueada.

### Frente B — Catálogo de templates
- **O quê:** tabela de templates (nome, categoria, variáveis esperadas, status de aprovação), tela de CRUD, validação de variáveis antes do envio (evita rejeição, alimenta o KPI "mensagens rejeitadas" do guia — ainda que a *captura* de rejeição real dependa do client real, D2).
- **Depende de:** nada além do schema — desbloqueada.

### Frente C — Janela por categoria
- **O quê:** substituir `conversations.window_expires_at` por rastreamento por categoria; ajustar dashboard (`window_expiring`) e contrato (`ConversationSummary.window_expires_at`) pra refletir a granularidade nova — provavelmente `window_expires_at` continua existindo como "a mais próxima de expirar", mas passa a ser derivada, não a fonte de verdade.
- **Depende de:** Frente B (categoria de template já precisa existir como conceito modelado antes de rastrear janela por categoria).
- **Pendente:** decidir com o Felipe se `window_expires_at` no contrato do frontend muda de shape ou só de como é calculada por trás (impacto de frontend a avaliar, não coberto neste plano).

### Frente D — Business Agent (WhatsApp)
- **O quê:** endpoint(s) de configuração do agente (nome, instruções, fontes de conhecimento, tópicos de transferência) — provavelmente estendendo `PUT /settings/meta` no mesmo padrão já usado pra IG/FB; listener de webhook pro "gatilho de transferência" que seta `mode: human`.
- **Depende de:** D2 (client real) pra funcionar de verdade — mas o *schema* e o *contrato* de configuração podem ser desenhados e mockados antes disso, como já é o padrão deste projeto pros outros clients Meta.
- **Pendente:** D4 (escopo por canal — recomendação é WhatsApp-only, falta confirmação), e a lacuna de billing de tokens (seção 2.1) — essa parte da Frente A específica do Business Agent fica bloqueada até a Meta documentar isso melhor ou até existir conta real pra inspecionar.

### Frente E — CTWA
- **O quê:** parser de `referral` no webhook (webhook/usecase, mesmo módulo que já faz `ingestEngagements`); entidade nova de campanha de anúncio (cliques, ativação de janela 72h, conversão); resolução do rename de `ai_campaigns` (seção 3).
- **Depende de:** confirmação do shape exato do `referral` contra doc oficial antes de codar o parser final (ressalva da seção 2.4) — mas a modelagem de tabela pode ser feita com o shape inferido, seguindo o mesmo padrão de risco aceito que `social_engagements` já usa hoje pros webhooks de comentário/story.

### Frente F — Alertas persistentes
- **O quê:** mecanismo de notificação dentro do CRM que sobrevive a sessão (não é só SSE broadcast) — provavelmente uma tabela `alerts`/`notifications` com leitura via endpoint próprio, populada por regras batidas contra a Frente A (estouro de orçamento, taxa de rejeição, etc. — os 6 alertas da seção 10 do guia).
- **Depende de:** Frente A (não dá pra alertar sobre custo sem o custo estar sendo capturado).

### Frente G — WhatsApp Flows / interativos
- **O quê:** novo `kind` de mensagem (`interactive`), client de envio de Flow (referenciando Flow publicado no Meta Flow Builder — a montagem do formulário em si é configuração fora do CRM, não código), parser de `nfm_reply` no webhook, UI pro atendente disparar um Flow existente.
- **Depende de:** D2 pra funcionar de verdade (Flow real só existe com conta Meta ativa) — mas schema/parser podem ser feitos contra payload mockado seguindo a seção 2.3.
- **Pendente:** confirmar o shape exato de `nfm_reply`/`response_json` com doc oficial quando a conta existir (mesma ressalva de sempre).

---

## 5. Backlog priorizado por dependência (sem datas — D11)

1. **Frente A** (núcleo de custo) — base de tudo, zero dependência.
2. **Frente B** (catálogo de templates) — zero dependência, pode andar em paralelo com A.
3. **Frente C** (janela por categoria) — depende de B.
4. **Frente F** (alertas persistentes) — depende de A.
5. **Frente E** (CTWA) — independente das anteriores, mas menor prioridade por ter mais incerteza de shape (2.4); pode andar em paralelo com C/F se houver capacidade.
6. **Frente D** (Business Agent) — schema/contrato podem ser desenhados cedo, mas a parte funcional real fica presa em D2 (sem conta) e na lacuna de billing (2.1) — tratar como "modelar agora, ativar quando a conta existir".
7. **Frente G** (Flows/interativos) — mesma lógica da D: schema/parser cedo, ativação real presa em D2.

**Nota:** como D2 mantém tudo em mock por ora, as frentes D e G podem ser desenvolvidas e testadas inteiramente contra payloads sintéticos — exatamente como o resto do `pkg/meta` já funciona hoje (`backend/ARCHITECTURE.md` §5/§6) — sem bloquear o resto do backlog.

---

## 6. Pendências explícitas (não decididas ainda, precisam voltar pro Felipe)

- **D4** — confirmar se IG/FB mantêm a IA de triagem própria (recomendação deste plano, baseada em o Business Agent ser onboardado por `phone_number_id`) ou se há outra intenção.
- **Billing de tokens do Business Agent** (2.1) — não documentado publicamente de forma clara; decidir se o CRM tenta modelar isso agora com um placeholder, ou espera confirmação da Meta/conta real antes de desenhar essa coluna específica.
- **Nome final da entidade renomeada de `ai_campaigns`** (seção 3) — sugestão dada, não é decisão.
- **Impacto no contrato do frontend** de `window_expires_at` deixar de ser uma coluna única (Frente C) — não avaliado neste plano, precisa de uma passada equivalente pelo lado frontend antes de codar.
