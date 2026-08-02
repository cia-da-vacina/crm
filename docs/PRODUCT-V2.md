# CRM Cia da Vacina — Product Brief v2 (multicanal)

**Revisão:** 2026-08-02
**Substitui/estende:** `docs/spec.md` e `docs/APPROVED-SCOPE.md` (MVP WhatsApp-only, aprovado em 2026-07-14).
**Equipe:** Felipe (Backend/Go) · Cristian (Frontend/Next.js)

Este documento descreve a versão 2 do produto: o mesmo CRM de atendimento e pipeline comercial, agora multicanal (WhatsApp + Instagram + Facebook Messenger) e com o frontend rodando em produção real (sem mocks, com backend Go via BFF).

---

## 1. Visão

Centralizar **todo** o atendimento das 5 unidades Cia da Vacina que chega pela Meta — WhatsApp, Instagram e Facebook Messenger — em uma única fila de trabalho, com uma camada de triagem por IA que qualifica e prepara o contexto antes de qualquer humano responder, e um pipeline comercial que dá visibilidade de conversão por unidade.

A v2 reconhece que o cliente da Cia da Vacina não escolhe um único canal: ele manda mensagem no WhatsApp, comenta em um post, responde a um story ou manda DM no Instagram — e espera ser reconhecido como o mesmo cliente, com a mesma qualidade de atendimento, em qualquer um deles.

---

## 2. Personas

### Administrador
- Gerencia usuários, unidades, POPs, motivos de não conversão e configurações de canal Meta (tokens, IA, campanhas).
- Enxerga o dashboard consolidado das 5 unidades.
- Único papel com acesso a `Configurações → Meta` (tokens mascarados, flags de IA, prompts).

### Atendente (agente)
- Opera na(s) unidade(s) vinculada(s) a ele.
- Trabalha a partir de duas filas: **Inbox** (conversas) e **Interações** (engagements de rede social ainda não viraram conversa).
- Assume conversas da IA (`claim`), responde, move o pipeline, registra motivo de não conversão, trabalha follow-ups.
- Não vê tokens, prompts de IA ou configurações de outras unidades.

*(Supervisor/Gerente existem no RBAC — `frontend/src/domain/enums.ts` — como papéis intermediários de fila/qualidade e visão de unidade, mas não têm telas dedicadas além do que Admin/Atendente já cobrem nesta versão.)*

---

## 3. Jornadas de usuário

### 3.1 Login → Inbox → Handoff de triagem → Resposta humana → Pipeline

1. Usuário acessa `/login`, autentica com email/senha.
2. Sessão é criada via BFF (cookies httpOnly); ao ter mais de uma unidade vinculada, a primeira é selecionada como ativa automaticamente (trocável no seletor de unidade do topo).
3. Usuário cai em `/inbox`, filtrado pela unidade ativa: lista de conversas com etapa de pipeline, canal, modo (`IA` ou `Humano`) e prévia da última mensagem.
4. Ao abrir uma conversa em `mode: "ai_triage"`, vê um banner de contexto da IA (resumo, intenção identificada) e o composer fica **bloqueado** — só pode digitar depois de **"Assumir atendimento"**.
5. Ao assumir (`claim`), a conversa muda para `mode: "human"`, o `owner_id` vira o usuário atual, e o composer libera. A partir daqui o histórico da IA continua visível (nada é apagado no handoff).
6. Atendente responde livremente, podendo inserir POPs sugeridos por intenção diretamente no composer.
7. Ao concluir a negociação, abre o modal **Pipeline**, escolhe a nova etapa; se for `Não fechado`, é obrigado a escolher um motivo do catálogo antes de confirmar.
8. Conversas em `aguardando_fechamento`/`nao_fechado` alimentam a fila de **Follow-ups**, onde o atendente retoma contato depois.

### 3.2 Engagement (story/comentário) → Resposta/Conversão

1. Um cliente comenta em um post, responde a um story ou é mencionado em um story da unidade.
2. O evento chega via webhook Meta e aparece na fila `/engagements` como um item `open`, com tipo (`story_reply`, `story_mention`, `post_comment`, `live_comment`) e canal.
3. Atendente abre o item, lê o conteúdo (texto + mídia, se houver) e pode:
   - **Responder** diretamente ali (`status → replied`) — sem necessariamente virar uma conversa formal;
   - **Descartar** (`status → dismissed`) se não exigir ação (ex.: spam, comentário genérico);
   - **Converter em conversa** (`status → converted_to_conversation`) quando a interação evolui para um atendimento real — nesse momento uma `ConversationDetail` é criada e o atendente é redirecionado para o Inbox para seguir a jornada normal de pipeline.
4. Interações não resolvidas contam no dashboard (`open_engagements`) como um indicador de fila pendente, separado da fila de conversas.

---

## 4. Decisão de ID de cliente entre canais

A Meta não expõe um identificador único entre WhatsApp, Instagram e Facebook — cada canal usa o seu (`wa_id`, IGSID, PSID).

**Decisão de produto:** a identidade canônica do cliente (`Customer.id`) é do CRM, e a **chave de negócio para unificar canais é o número de telefone (E.164)** — não CPF.

### Como cada canal nasce

| Canal | Telefone na entrada | Resultado |
|---|---|---|
| **WhatsApp** | Meta já entrega o número | Cliente criado já **`identified`** — **não se pede telefone** |
| **Instagram** | Não há telefone nativo | Cliente nasce **`anonymous`** — telefone só se a parede exigir |
| **Facebook Messenger** | Não há telefone nativo | Idem Instagram |

Quando o telefone é informado em IG/FB (e a parede exige), o backend:
1. normaliza para E.164 e inicia verificação (`phone_gate → pending_verification`);
2. envia **confirmação OTP via WhatsApp** para provar posse do número;
3. só após OTP válido: cria ou **funde** o `Customer` pela chave `primary_phone`;
4. vincula a `CustomerIdentity` (IGSID/PSID) a esse `Customer`;
5. promove `identification → identified` e `phone_gate → collected`.

**Sem OTP não há merge nem liberação da parede** — evita que alguém no Instagram use o telefone de outra pessoa.

Merge manual por agente continua disponível como fallback.

### Parede de privacidade (anônimo × identificado)

O cliente **pode conversar sem informar telefone**. Só pedimos quando a intenção de verdade exige um cadastro real. Pedir número em IG/FB **sempre** exige confirmação no WhatsApp antes de liberar o lado “identificado”.

| Sem telefone / OTP pendente (`anonymous`) | Telefone confirmado (`identified`) |
|---|---|
| Tirar dúvidas gerais, preços, horários, campanhas | Histórico completo entre canais |
| Conversar com a IA de triagem / atendimento leve | Agendar / retomar agendamento |
| Engagements leves (comentário, story) sem virar prontuário | Reclamações que precisam de prontuário/histórico |
| Sem merge cross-canal | Follow-ups e dados cadastrais do CRM |
| Agente vê perfil limitado (handle do canal) | Ficha unificada + identidades de canal |

**Regra de coleta:** o backend define `phone_gate` por conversa (`not_needed` | `required` | `pending_verification` | `collected`). Exemplos típicos:
- `precos` / `duvidas` leves → `not_needed` (não perguntar)
- `agendar` / `reclamacao` em IG/FB → `required` → OTP WhatsApp → `collected`
- WhatsApp → quase sempre `not_needed` ou já `collected` (número veio da Meta; posse já implícita)

Frontend **não decide** essa regra — só exibe o estado e bloqueia ações gated enquanto não estiver `collected` (quando a intenção for gated).

Ver contrato técnico em [`BACKEND-CONTRACT.md`](./BACKEND-CONTRACT.md#3-customers-crm-id--telefone-como-chave--identidades-por-canal).

---

## 5. Estratégia de canais

| Canal | Uso | Observação |
|---|---|---|
| **WhatsApp** | Canal principal de atendimento 1:1, agendamento e negociação. | Janela de atendimento de 24h da Meta; fora dela, requer template aprovado. |
| **Instagram** | DM + engagements (story reply/mention, comentários). Canal de descoberta/awareness que vira atendimento. | Identidade via IGSID; telefone **só sob demanda** (parede de privacidade). |
| **Facebook Messenger** | DM da Page + comentários em posts/lives. | Identidade via PSID; mesma regra de telefone do Instagram. |

Todos os três convergem para a mesma fila de trabalho (Inbox + Interações) e o mesmo pipeline comercial — não há tratamento de segunda classe para nenhum canal.

---

## 6. Regra: triagem antes do humano

**Toda conversa nova começa em `mode: "ai_triage"`.** Nenhum agente humano deve iniciar o primeiro contato — a IA sempre:

1. Recebe e cumprimenta o cliente.
2. Identifica a intenção (`agendar`, `precos`, `duvidas`, `reclamacao`, `outro`).
3. Decide se a parede exige telefone (`phone_gate`):
   - WhatsApp → não pede (número já veio da Meta).
   - Instagram / Messenger + intenção leve → **não pede**.
   - Instagram / Messenger + intenção gated (`agendar`, etc.) → pede telefone → envia **OTP no WhatsApp** → só confirma posse e identifica/faz merge após código válido.
4. Coleta demais campos estruturados relevantes (unidade, vacina…).
5. Produz um resumo (`TriageSummary`) para acelerar o handoff.

A transição para humano é **sempre explícita** (`claim`), nunca automática por tempo — o agente decide quando assumir, mas só pode responder depois de assumir. Isso garante que:
- nenhuma resposta "crua" do agente aconteça sem o contexto da triagem já coletado;
- o kill-switch de IA (`triage_enabled: false` em Configurações) força todas as conversas novas direto para fila humana sem quebrar o restante do fluxo.

Esta regra vale para **todos os canais** (WhatsApp, Instagram, Messenger) — não é exclusiva de WhatsApp nesta versão.

---

## 7. Instalação como PWA

O frontend é um Progressive Web App instalável (manifest + service worker via `@ducanh2912/next-pwa`):
- Ícones e cores de tema da Cia da Vacina, modo `standalone` (sem chrome de navegador).
- Prompt de instalação nativo do navegador é capturado e exibido como banner customizado ("Instalar" / "Agora não"), com preferência de dispensa por sessão.
- Página de fallback offline (`/~offline`) quando não há rede.
- Alvo de uso: atendentes acessando pelo celular/tablet nas unidades, sem depender de app nativo publicado em loja.

---

## 8. Fora de escopo (ainda)

Mesmo com a expansão multicanal, os seguintes itens **continuam explicitamente fora** desta versão:

- **Apps mobile nativos** (iOS/Android publicados em loja) — a PWA cobre a necessidade de "acesso rápido no celular".
- **IA conversacional contínua** — a IA só faz triagem inicial + handoff; não substitui o atendente na negociação, nem continua respondendo após o `claim`.
- **BI avançado** — sem exportações complexas, sem data warehouse dedicado; o dashboard é o KPI mínimo operacional (abertos, por etapa, por canal, conversão, follow-up pendente, engagements pendentes).
- Discador, e-mail, SMS como canais.
- Integração com agenda/prontuário clínico.
- SSO corporativo.

---

## 9. Checklist de critérios de aceite

**Canais**
- [ ] Mensagens de WhatsApp, Instagram (DM) e Facebook Messenger chegam na mesma fila de Inbox.
- [ ] Engagements de Instagram/Facebook (story reply, story mention, comentário de post, comentário de live) aparecem na fila de Interações.
- [ ] WhatsApp cria cliente já `identified` (telefone da Meta) sem perguntar número.
- [ ] Instagram/Messenger nascem `anonymous`; telefone só é pedido quando `phone_gate === "required"`.
- [ ] Número informado em IG/FB dispara OTP WhatsApp (`pending_verification`); merge/`identified` só após confirmação.
- [ ] Informar o mesmo telefone **confirmado** em canais diferentes funde no mesmo `Customer` (`primary_phone`).

**Parede de privacidade**
- [ ] Conversa anônima consegue tirar dúvidas/preços sem fornecer telefone.
- [ ] Ações gated ficam bloqueadas até `phone_gate === "collected"` / `identification === "identified"`.
- [ ] Enquanto `pending_verification`, UI mostra número mascarado e estado “Confirmando no WhatsApp”.
- [ ] Agente vê indicação clara de “Sem telefone” vs “Identificado” na conversa.

**Triagem e handoff**
- [ ] Toda conversa nova nasce em modo IA, independente do canal.
- [ ] Composer de resposta humana fica bloqueado até o `claim`.
- [ ] Resumo de triagem (`TriageSummary`) é exibido ao agente antes/durante o handoff, incluindo `phone_gate`.
- [ ] Kill-switch (`triage_enabled: false`) desliga a IA sem quebrar o fluxo de conversas existentes.

**Pipeline e follow-up**
- [ ] Mover para `Não fechado` exige motivo do catálogo.
- [ ] Conversas em `aguardando_fechamento`/`nao_fechado` geram itens de follow-up.
- [ ] Dashboard reflete contagens por etapa, por canal e por unidade em tempo quase real.

**Engagements**
- [ ] Interação pode ser respondida, descartada ou convertida em conversa.
- [ ] Conversão em conversa preserva o vínculo com a interação de origem e redireciona o agente ao Inbox.

**Segurança e configuração**
- [ ] Tokens Meta nunca aparecem completos na UI — apenas mascarados.
- [ ] Rotação de token por canal funciona via `Configurações → Meta` sem expor o valor anterior.
- [ ] Sessão de usuário sobrevive a refresh de página via cookie httpOnly (sem re-login desnecessário).

**PWA**
- [ ] App é instalável em Chrome/Edge (Android/desktop) e Safari (iOS, via "Adicionar à tela de início").
- [ ] Página offline é exibida quando não há conectividade.
