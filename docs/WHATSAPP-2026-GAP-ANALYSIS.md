# Relatório — Adequação do CRM ao Novo Modelo de Cobrança WhatsApp (out/2026)

**Objeto:** análise de `WhatsApp_API_Optimization_Guide.md` (raiz do repo) contra o estado real do backend/frontend, pra responder três perguntas do Felipe: (1) falta adaptar mais alguma coisa? (2) os KPIs do guia são implementáveis como estão? (3) o que falta é código ou é só treinar o operador?
**Não escrevi código** — isto é só levantamento, conforme pedido.

---

## Resposta curta

O guia descreve corretamente **como a Meta vai cobrar**, mas descreve um **CRM que ainda não existe**. Ele foi escrito na camada de processo/operação (seções 8–13, treinamento e checklists) como se a camada técnica (seção 7, "Conceitos que o backend precisa implementar") já estivesse pronta. Ela não está — nenhum dos 8 itens da seção 7 tem hoje um módulo, tabela ou endpoint correspondente no backend atual. Hoje isso **não é** "questão de operador": é questão de engenharia primeiro, treinamento depois. Treinar os atendentes na "regra dos 3 em 1" ou nas respostas rápidas já vale a pena começar agora (não depende de nada técnico), mas o dashboard de custos, o rastreamento de janela por categoria e o CTWA analytics do guia **não têm onde rodar** ainda.

Também há um ponto anterior a tudo isso, já mapeado em `backend/ARCHITECTURE.md`: os clients HTTP reais da Meta (WhatsApp Cloud API) **ainda não foram escritos** — só existe `pkg/meta.MockClient`. Sem isso, nada do guia é executável em produção, independente do resto.

---

## 1. O que o guia exige da engenharia (seção 7 do guia) vs. o que existe hoje

| # | Exigência do guia | Existe hoje? | Onde olhei |
|---|---|---|---|
| 1 | Webhook classifica tipo de atendimento → decide camada (IA/template/humano) → registra canal e custo | Webhook ingere e roteia por canal/identidade (`internal/module/webhook`), mas **não decide camada por custo** nem grava custo — esse conceito não existe no schema | `backend/ARCHITECTURE.md` §1/§3, `messages` (migration 000010) |
| 2 | Cadastro de templates aprovados por categoria + validação de variáveis | **Não existe.** `messages.template_name` é um `TEXT` livre gravado no momento do envio — não há tabela de templates, categoria (marketing/utilidade/autenticação), nem status de aprovação | `conf/migrations/000010_create_messages_table.up.sql` |
| 3 | Janela de 24h **por categoria de template** (Regra 4 do guia) | **Não existe.** `conversations.window_expires_at` é uma **única** coluna por conversa — não há como representar "janela de marketing aberta E janela de utilidade aberta ao mesmo tempo com o mesmo cliente" | `conf/migrations/000009_create_conversations_table.up.sql:18` |
| 4 | Núcleo de cálculo de custo por mensagem enviada | **Não existe.** Nenhuma coluna de custo em `messages`, nenhuma tabela de preço/rate card, nenhuma agregação. O próprio contrato do dashboard diz explicitamente "sem receita/ticket/período" (decisão de escopo do MVP, não esquecimento) | `docs/BACKEND-CONTRACT.md:365` |
| 5 | Integração com Meta Business Agent (tokens consumidos, escalação) | **Não existe e é um produto diferente do que já foi construído.** A IA atual do CRM (`internal/module/triage`) é IA **própria** (OpenAI `gpt-4o-mini`), cobrada por *mensagem* pela lógica do próprio guia — não é o Meta Business Agent (IA nativa da Meta, hospedada por ela, cobrada por token diretamente pela Meta). São dois produtos distintos com dois modelos de cobrança distintos (ver §3 abaixo) | `backend/ARCHITECTURE.md` §6, `internal/module/triage` |
| 6 | Analytics de campanhas CTWA (cliques, ativação de janela 72h, conversão, ROI) | **Não existe — e há colisão de nome.** A tabela/entidade `ai_campaigns` (`GET/PUT /settings/meta`, UI em `/campaigns`) **não é** campanha de anúncio CTWA — é só um período com título/descrição que alimenta o *prompt* da IA de triagem (ex.: "campanha de gripe em maio"). Não tem `ad_id`, clique, janela de 72h nem conversão. O webhook também não faz parse do objeto `referral` que a Meta manda quando a mensagem vem de um clique em anúncio | `frontend/src/app/(auth)/campaigns/page.tsx:274-277`, `conf/migrations/000015_create_ai_campaigns_table.up.sql`, `docs/BACKEND-CONTRACT.md:385` |
| 7 | Histórico de mensagens com tipo, categoria, custo, atendente/modelo | **Parcial.** `messages` já grava `sender_type` (`contact/agent/ai/system`), `kind`, `channel`, `template_name`, `sender_user_id` — a base existe. Faltam só as colunas de **categoria de cobrança** e **custo** | `conf/migrations/000010_create_messages_table.up.sql` |
| 8 | Alertas e monitoramento (seção 10 do guia) | **Não existe.** Não há mecanismo de alerta automático no backend hoje (nem para isso nem para outra coisa) | busca em `internal/`, `pkg/` |

**Fundação que já ajuda:** `pkg/cursor` (paginação já pronta pra qualquer novo endpoint de alto volume, tipo histórico de custo), `pkg/audit` (padrão de log append-only replicável pra granularidade de custo por mensagem, embora provavelmente o caminho certo aqui seja colunas na própria `messages`, não uma tabela de auditoria separada), e o fato de `messages.kind` já ter `'template'` como valor válido — a modelagem de mensagem não precisa ser refeita, só estendida.

---

## 2. Os KPIs do guia — o que dá pra calcular hoje e o que exige dado novo

O guia pede, nas seções 10 e 9, um conjunto de métricas. Cruzei cada uma com `internal/module/dashboard/repository/repository.go` (as únicas queries de métrica que existem hoje):

| KPI pedido pelo guia | Dá pra calcular com o schema atual? |
|---|---|
| Total de conversas / taxa de resolução por IA / taxa de escalação | **Parcialmente.** `conversations.mode` (`ai_triage`/`human`) já existe e já é contado no dashboard (`GetByStage`, `GetCounts`) — dá pra aproximar "resolvido por IA" vs "escalado", mas a métrica não existe pronta, precisa de uma query nova (não de schema novo) |
| Custo total / custo por mensagem / breakdown por tipo de template | **Não.** Precisa de coluna de custo + tabela de categoria/preço — dado que não existe hoje em lugar nenhum |
| Média de mensagens por atendimento (meta ≤ 2) | **Sim, com query nova.** `messages` já tem `conversation_id` + `sender_type` — um `COUNT(*) GROUP BY conversation_id` filtrando `sender_type='agent'` já responde isso hoje, sem mudar schema |
| Mensagens rejeitadas pela Meta | **Não.** Não há captura de erro/rejeição de envio de template em lugar nenhum — nem hoje existe client HTTP real que receberia esse erro da Meta |
| ROI/conversão de campanhas CTWA | **Não.** Depende do item 6 da tabela acima (não existe rastreamento de clique/anúncio) |
| Tempo médio de resposta | **Parcialmente.** `messages.created_at` existe; dá pra calcular delta entre última mensagem `in` e primeira `out` por conversa com uma query nova — viável sem schema novo |
| Taxa de rejeição de template < 5% (auditoria quinzenal) | **Não**, mesma razão do item de rejeição acima |

Ou seja: **dois dos sete KPIs são só query nova** (esforço baixo, sem migration); **os outros cinco exigem schema novo e, em alguns casos, a integração Meta real** que ainda não existe.

---

## 3. Ponto que precisa de decisão do time antes de qualquer código

Isso não é um gap técnico, é uma decisão de produto que muda o escopo do trabalho:

**O guia assume adoção do Meta Business Agent.** Esse é um produto da Meta que roda a IA *hospedada por ela*, cobrado por token direto pela Meta — diferente da IA de triagem que já foi construída (OpenAI própria, cobrada por chamada, com regras de negócio determinísticas em cima da resposta — `phone_gate`, handoff por intent — ver `backend/ARCHITECTURE.md` §6). Adotar o Business Agent da Meta, além do que já existe, significa:
- Um novo fluxo de configuração (base de conhecimento, prompts, rotas de transferência — seção 6 do guia) que hoje não tem endpoint nem UI;
- Rastrear tokens consumidos por resposta via API da Meta (dado que a IA atual nunca precisou expor, porque quem cobra por token é a OpenAI, não a Meta);
- Decidir como as duas IAs convivem: a de triagem decide `phone_gate`/handoff hoje; se o Business Agent também responde no WhatsApp, quem manda em qual conversa?

Antes de desenhar schema/endpoint pra "Business Agent", vale confirmar com o Felipe (e possivelmente com quem aprovou o guia) se a Cia da Vacina vai mesmo ativar esse produto da Meta, ou se o guia está descrevendo uma opção genérica e a intenção real é continuar só com a IA de triagem já construída — cobrada como "texto livre"/"IA própria" pela tabela de custo do guia, não como Business Agent. São dois roadmaps de engenharia bem diferentes.

---

## 4. E as funcionalidades de UX do guia (Flows, botões/listas)?

A seção 8 do guia (treinamento) recomenda WhatsApp Flows e botões/listas interativas pra reduzir mensagens fragmentadas. Isso também não é "treinamento puro" — `messages.kind` (migration 000010) só aceita `'text', 'image', 'document', 'audio', 'video', 'template', 'system'`. Não há `'interactive'`/`'flow'` como tipo de mensagem, nem parser de payload de Flow no webhook, nem componente de UI pra atendente montar/enviar Flow. É trabalho de engenharia real (schema + parser + client Meta real, que por sua vez depende do client HTTP que ainda não existe), não algo que o operador ativa numa configuração.

---

## 5. O que já é, hoje, só questão de operador (pode começar já)

Nem tudo depende de código. Estas partes do guia são **puramente operacionais** e não esperam nada do backend:

- Treinamento da "regra dos 3 em 1" e das metas de mensagens por atendimento (seção 8) — depende de disciplina do atendente, não de feature.
- Biblioteca de respostas rápidas — se for só um documento/checklist pros atendentes hoje (não uma feature nova no CRM), pode rodar já.
- Onboarding/checklist de novo atendente (seção 8) — processo de RH/gestão, independe do sistema.
- Checklist de aprovação de templates (seção 4) — processo de submissão à Meta, roda fora do CRM.

Vale começar por aí enquanto o backlog técnico é priorizado — não há motivo pra esperar o código pronto pra treinar a equipe na parte comportamental.

---

## 6. Sequenciamento sugerido (não é plano de implementação, é ordem de dependência)

1. **Pré-requisito de tudo:** clients HTTP reais da Meta (`pkg/meta`, já mapeado como gap conhecido em `backend/ARCHITECTURE.md` §5/§6) — sem isso não existe envio real de template, nem erro de rejeição real pra capturar, nem nada do guia roda em produção.
2. **Decisão de produto:** confirmar se o Meta Business Agent entra no escopo ou não (seção 3 deste relatório) — muda o tamanho do trabalho de IA.
3. **Schema mínimo pra custo:** categoria de cobrança + custo por `message` (extensão simples de `messages`, sem quebrar nada existente).
4. **Catálogo de templates:** tabela nova (categoria, variáveis, status de aprovação) — hoje `template_name` é texto livre sem validação.
5. **Janela por categoria:** substituir/complementar `conversations.window_expires_at` (coluna única) por um tracking por categoria — é o item de maior risco de retrabalho se for adiado, porque `window_expires_at` já é usado pelo dashboard (`window_expiring`) e pela UI (`ConversationSummary.window_expires_at` no contrato) — mudar a forma de representar isso depois de mais telas dependerem dele custa mais.
6. **CTWA:** parse do `referral` no webhook + tabela de clique/anúncio — só depois de 1–2 resolvidos, porque também depende de client Meta real pra saber se a resposta em 24h ativou a janela de 72h.
7. **Dashboard de custo + alertas:** por último, porque é leitura de tudo que os itens acima passam a gravar.

---

## Resumo pro Felipe

- **Falta adaptar mais coisa?** Sim, bastante — o guia descreve 8 capacidades de backend que hoje não existem (tabela §1), não pequenos ajustes.
- **Os KPIs são implementáveis como estão?** Dois de sete métricas dão pra fazer só com query nova; os outros cinco exigem schema novo e, em parte, a integração Meta real que ainda está pendente.
- **Já é questão de operador?** Só a parte comportamental (seção 8 do guia) pode rodar já. O dashboard de custo, janela por categoria, catálogo de templates e CTWA analytics são trabalho de engenharia que ainda não começou.
