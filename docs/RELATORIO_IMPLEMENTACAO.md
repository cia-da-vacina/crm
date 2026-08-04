# Relatório de Implementação — Adaptação WhatsApp 2026

> **LEIA ISTO PRIMEIRO ao retomar o trabalho.** Esta é a fonte de verdade de *onde a implementação parou*. Antes de continuar:
> 1. Leia `docs/WHATSAPP-2026-GAP-ANALYSIS.md` — o que falta adaptar e por quê (relatório original).
> 2. Leia `docs/WHATSAPP-2026-ADAPTATION-PLAN.md` — as decisões tomadas com o Felipe (D1–D11) e as 7 "Frentes" (A–G) de trabalho, na ordem de dependência decidida.
> 3. Leia este arquivo — o que dessas 7 frentes já foi **codado**, o que falta, e as ressalvas técnicas descobertas durante a implementação que não estavam nos dois documentos acima.
> 4. `backend/ARCHITECTURE.md` continua sendo a referência de arquitetura geral do backend (padrão handler→usecase→repository, wiring, etc.) — as convenções descritas lá foram seguidas à risca em tudo que foi implementado aqui.
>
> Este arquivo deve ser atualizado a cada frente concluída/avançada — não deixe ele ficar desatualizado.

---

## Estado geral

**Frentes A e B do plano de adaptação estão implementadas em código** (schema, módulos, wiring, testes). **Frentes C, D, E, F, G ainda não foram iniciadas.** Nenhum client HTTP real da Meta foi escrito (decisão D2 do plano: ficam em mock por enquanto, sem conta de API ativa).

**Testes:** o ambiente onde isso foi implementado não tem Docker/Postgres disponível (`docker` não está instalado neste WSL). Todo o código foi validado com `go build ./...`, `go vet ./...` e `gofmt -l .` (todos limpos) e os testes que não dependem de banco **passam** (`go test ./...`). Os testes que dependem de Postgres real (a maioria dos testes de usecase deste projeto, por convenção — ver `internal/testutil/testutil.go`) **dão skip automaticamente** sem `DATABASE_URL` — foram escritos e compilam, mas **nunca rodaram de verdade contra um banco**. Antes de confiar neles, rode `make db-migrate && make test` dentro do container (`docker compose up -d`, depois `make test`) — é o primeiro passo ao retomar.

---

## Frente A — Núcleo de custo (IMPLEMENTADA)

### O que foi feito

1. **Schema** (`backend/conf/migrations/000020` e `000021`):
   - `messages` ganhou `pricing_category`, `pricing_billable`, `pricing_model`, `pricing_confirmed` (bool, default false), `cost_brl` (NUMERIC).
   - Nova tabela `message_pricing_rates` (categoria → valor em BRL + `billable`), com seed dos valores estimados do `WhatsApp_API_Optimization_Guide.md` (marketing 0,31 / utility 0,0068 / authentication 0,0068 / service 0,0350 / free_entry_point e free_customer_service = 0). Editável via API, não hardcoded — o próprio guia avisa que os valores definitivos só saem perto de 1º/set/2026.

2. **Módulo novo `internal/module/pricing`** — `GET /settings/pricing-rates` e `PATCH /settings/pricing-rates/{category}`, admin-only. Expõe `pricing.NewUseCase(a)` (não só `pricing.New(a)`) pra outros módulos consumirem só a leitura de taxa (`GetRate`) sem depender do módulo HTTP inteiro — mesmo padrão de `CustomerReader`/`Triage`/`Engagement` já usado no projeto.

3. **`conversation/usecase.SendMessage`** — toda mensagem de texto livre enviada por um agente agora é gravada com `pricing_category = "service"` e `cost_brl` estimado a partir do rate card (função `applyEstimatedPricing`). `pricing_confirmed` fica `false` — é uma estimativa local no momento do envio, não o que a Meta de fato cobrou.

4. **Reconciliação via webhook de status** (`internal/module/webhook`) — **isto é a peça mais especulativa desta frente**, ver ressalva abaixo:
   - `parseWhatsAppStatuses` (novo, em `webhook/usecase/parser.go`) extrai o array `statuses` do payload do webhook da WhatsApp Cloud API, incluindo o objeto `pricing` (`category`/`billable`/`pricing_model`).
   - `webhook/repository.UpdateMessagePricing` localiza a mensagem por `meta_message_id` e atualiza status + preço, marcando `pricing_confirmed = true` só quando `category` veio preenchido.
   - `webhook/usecase.ingestStatuses` chama isso pra cada evento, com lookup de custo via `PricingReader` (mesmo rate card da Frente A).

5. **Dashboard de custo** — endpoint novo `GET /dashboard/costs` (admin/manager only — dado financeiro, diferente de `GET /dashboard/summary` que é aberto a qualquer papel autenticado). **Não estende** `GET /dashboard/summary` de propósito: `docs/BACKEND-CONTRACT.md §6` documenta esse endpoint explicitamente como "sem receita/ticket/período", então criar um endpoint novo evita quebrar essa garantia documentada pro frontend existente. Devolve total gasto, contagem de mensagens precificadas/confirmadas, e breakdown por categoria.

6. **Modelo de API** (`conversation/model.Message`) ganhou `pricing_category`, `pricing_billable`, `pricing_confirmed`, `cost_brl` — mudança aditiva (novos campos opcionais), não quebra consumidores existentes do contrato.

### Ressalva técnica importante — shape do webhook de status

O shape de `statuses[].pricing` em `parser.go` foi montado a partir da pesquisa de documentação pública feita durante o planejamento (`docs/WHATSAPP-2026-ADAPTATION-PLAN.md §2.2`), **nunca confirmado contra um payload real da Meta** — mesma ressalva que já existe em `backend/ARCHITECTURE.md §8` pros parsers de engagement (comment/story). Quando existir conta de API real (D2 do plano — ainda não existe), o primeiro passo antes de confiar nisso em produção é validar campo a campo contra um payload de verdade.

### Limitação conhecida, não resolvida nesta rodada

**O envio de template (OTP, `conversation/usecase/phone.go:sendOTP`) não grava uma linha em `messages`** — isso já era assim antes desta implementação (não é uma regressão), mas significa que o custo de OTP (`authentication`, R$ 0,0068 por envio) **não é rastreado** pelo núcleo de custo. Criar uma `Message` pro envio de OTP exigiria decidir em qual `conversation_id` gravar (o fluxo de OTP às vezes roda antes de existir uma conversa clara pro número informado) — não foi resolvido pra não expandir escopo sem uma decisão de produto. **Próximo passo, se for priorizado:** decidir com o Felipe se OTP passa a gerar uma `Message` (kind=template, sender_type=system) só pra fins de custo/auditoria.

**`SendMessage` não aceita `kind: "template"`** — hoje só envia texto livre (era assim antes desta implementação também). O catálogo de templates da Frente B existe e está pronto pra ser consultado, mas nada no fluxo de conversa chama `sender.SendTemplate` pra uma mensagem de negócio ainda (só o OTP, que é um fluxo interno separado). Ligar os dois — permitir um agente escolher um template aprovado do catálogo e enviar via `POST /conversations/:id/messages` — é trabalho novo, não coberto aqui, e envolve mudar o contrato de `SendMessageRequest` (hoje só `{body, kind}`; precisaria de `template_name` + `params`). Fica pra quando isso for priorizado explicitamente.

---

## Frente B — Catálogo de templates (IMPLEMENTADA)

### O que foi feito

1. **Schema** (`backend/conf/migrations/000022`): tabela `message_templates` — `name`, `category` (marketing/utility/authentication), `language_code`, `body`, `variable_count`, `approval_status` (pending/approved/rejected), `active`. Único por `(name, language_code)`.

2. **Módulo novo `internal/module/template`** — CRUD completo (`GET/POST /templates`, `GET/PATCH/DELETE /templates/{id}`), leitura pra qualquer autenticado, escrita restrita a admin/manager (mesmo padrão RBAC do módulo `pop`).

3. **Validação de variáveis** (`template/usecase.validateVariableCount`) — conta os placeholders `{{n}}` no `body` e exige que bata exatamente com `variable_count` declarado, tanto na criação quanto quando o `body` é editado depois. Isso implementa diretamente o item do checklist do guia (`WhatsApp_API_Optimization_Guide.md §4`): "Deixar variáveis sem contexto" é uma das causas de rejeição pela Meta que o guia lista — agora é pego em 400 antes de qualquer tentativa de envio, não só documentado como boa prática.

4. **`repository.GetApprovedByNameAndLanguage`** já existe, pronta pra ser consumida quando `SendMessage` (Frente A, limitação acima) passar a validar um template antes de enviar — não é chamada por ninguém ainda, é só a peça que falta plugar.

### Não feito nesta frente

- Nenhuma tela de frontend — só a API.
- Nenhuma integração com a API de templates da Meta (consultar status de aprovação real via API) — `approval_status` é setado manualmente hoje, porque não existe client HTTP real (D2).

---

## Frentes C, D, E, F, G — NÃO INICIADAS

Seguindo a ordem de dependência do plano (`docs/WHATSAPP-2026-ADAPTATION-PLAN.md §5`):

| Frente | O que é | Status | Observação de quem retomar |
|---|---|---|---|
| **C** | Janela de 24h por categoria (substituir `conversations.window_expires_at` único) | Não iniciada | Depende da Frente B (feita) — pode começar. Maior risco de retrabalho se continuar adiando (mais telas passam a depender da coluna única) — ver plano §3. |
| **D** | Meta Business Agent (substitui a IA de triagem — decisão D1) | Não iniciada | Schema/contrato podem ser desenhados mesmo sem conta real (mesmo padrão mock de tudo mais), mas a lacuna de billing de tokens (plano §2.1) segue sem resposta pública clara. Pendência D4 do plano (escopo por canal) ainda não confirmada pelo Felipe. |
| **E** | CTWA (parse de `referral`, entidade de campanha de anúncio, resolver colisão de nome com `ai_campaigns`) | Não iniciada | Shape do `referral` tem a mesma ressalva de "nunca visto de payload real" que o `statuses.pricing` desta rodada (plano §2.4) — inclusive **mais incerto**, porque as fontes encontradas na pesquisa divergiam entre si (`referral` de primeiro nível vs. aninhado em `context.ad`, conforme agregador). Confirmar contra doc oficial antes de codar o parser. |
| **F** | Alertas persistentes (não só SSE) | Não iniciada | Depende da Frente A (feita) — os dados de custo já existem pra alimentar os 6 alertas do guia §10. Falta decidir o mecanismo de persistência (tabela `alerts`/`notifications` + endpoint de leitura, análogo a `audit_logs`/`follow_ups`) — decisão D9 já define que precisa sobreviver a login/logout, não é só broadcast SSE. |
| **G** | WhatsApp Flows / botões interativos | Não iniciada | `messages.kind` ainda não tem `interactive`/`flow` no CHECK (migration 000010, não tocada nesta rodada). Pesquisa de padrões da API já foi feita durante o planejamento (plano §2.3) — resta desenhar schema + parser de `nfm_reply` + client de envio, tudo em mock por enquanto (D2). |

---

## Arquivos tocados nesta rodada (para revisão rápida)

Migrations novas:
```
backend/conf/migrations/000020_add_messages_pricing.{up,down}.sql
backend/conf/migrations/000021_create_message_pricing_rates_table.{up,down}.sql
backend/conf/migrations/000022_create_message_templates_table.{up,down}.sql
```

Módulos novos (completos):
```
backend/internal/module/pricing/   (repository, model, usecase, handler, module.go)
backend/internal/module/template/  (repository, model, usecase + usecase_test.go, handler, module.go)
```

Módulos existentes, modificados:
```
backend/internal/domain/entity/entity.go            — PricingCategory, MessagePricingRate, TemplateCategory,
                                                        TemplateApprovalStatus, MessageTemplate; Message ganhou
                                                        campos de pricing
backend/internal/module/conversation/{model,module,repository,usecase}/*  — SendMessage estima custo; wiring do PricingReader
backend/internal/module/webhook/{model,module,repository,usecase}/*      — parseWhatsAppStatuses, UpdateMessagePricing,
                                                                             ingestStatuses; wiring do PricingReader
backend/internal/module/dashboard/{model,module,repository,usecase,handler}/*  — GET /dashboard/costs
backend/cmd/server/main.go                            — registra pricing.New(a) e template.New(a)
```

Testes novos/estendidos:
```
backend/internal/module/template/usecase/usecase_test.go            (novo — puro, sem DB)
backend/internal/module/webhook/usecase/parser_pricing_test.go       (novo — puro, sem DB)
backend/internal/module/webhook/usecase/usecase_test.go              (+1 teste, precisa de DB)
backend/internal/module/conversation/usecase/usecase_test.go         (+1 teste, precisa de DB)
```

## Como validar ao retomar

```bash
cd backend
go build ./...        # deve compilar limpo
go vet ./...           # deve ficar limpo
gofmt -l .              # não deve listar nada

# com Docker disponível:
make up                 # sobe api + postgres
make db-migrate          # aplica as migrations 000020-000022
make test                # roda a suíte inteira, incluindo os testes novos que dependem de DB
```

Os três testes de DB mais importantes pra conferir depois de `make test` (todos novos nesta rodada):
- `TestSendMessage_EstimatesServiceCostFromRateCard` (conversation/usecase) — confirma que enviar texto livre grava `pricing_category`/`cost_brl` corretos.
- `TestIngestPayload_WhatsApp_StatusWithPricing_ReconcilesMessageCost` (webhook/usecase) — confirma que um evento de status com `pricing` reconcilia a mensagem existente.
- `TestCreate_VariableCountMismatch_Rejected` (template/usecase) — já rodou e passou nesta sessão (não precisa de DB), mas vale conferir de novo se algo mudar na validação.
