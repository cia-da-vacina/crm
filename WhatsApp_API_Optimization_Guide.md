# WhatsApp Business API — Guia de Otimização de Custos
## CRM Omnichannel — Cia da Vacina

**Preparado para:** Equipe de Desenvolvimento (Backend & DevOps) e Operadores do CRM (Cia da Vacina)
**Contexto:** Este guia descreve o modelo de cobrança que passa a valer em 1º de outubro de 2026 e como o CRM e a operação devem funcionar dentro dele. Os valores usados são estimativas com base no rate card público mais recente da Meta — os valores definitivos costumam ser publicados até 1º de setembro de 2026.

---

## Índice

1. [Como a Meta Cobra as Mensagens](#cobranca)
2. [Tabela de Referência de Custos](#tabela-custos)
3. [Arquitetura de Atendimento em 3 Camadas](#arquitetura)
4. [Estratégia de Templates](#templates)
5. [Campanhas Click-to-WhatsApp (Entrada Gratuita de 72h)](#ctwa)
6. [Meta Business Agent](#business-agent)
7. [Arquitetura Técnica do CRM](#arquitetura-tecnica)
8. [Treinamento de Atendentes — Como Gastar Menos](#treinamento)
9. [Procedimentos Operacionais Diários](#procedimentos)
10. [Monitoramento de Custos](#monitoramento)
11. [Conformidade e Boas Práticas](#conformidade)
12. [Suporte e Troubleshooting](#suporte)
13. [Cronograma de Implementação](#cronograma)

---

## 1. Como a Meta Cobra as Mensagens {#cobranca}

Tudo gira em torno de um conceito: a **janela de atendimento de 24 horas**. Ela abre toda vez que um cliente manda uma mensagem, e determina o que pode ser enviado e a que custo.

### Regra 1 — Cliente escreveu há menos de 24 horas

A janela está aberta. Você pode responder de três formas:

| Forma de resposta | Custo |
|---|---|
| Texto livre (atendente humano ou IA própria do CRM) | R$ 0,035 por mensagem |
| Template de utilidade (ex: confirmação, lembrete) | R$ 0,0068 por mensagem |
| Template de autenticação (ex: código OTP) | R$ 0,0068 por mensagem |
| Template de marketing (ex: promoção) | R$ 0,31 por mensagem |

Não existe mais "resposta grátis dentro da janela" — toda mensagem enviada pela empresa é cobrada, texto livre incluso.

### Regra 2 — Cliente não escreve há mais de 24 horas

A janela fechou. Só é permitido enviar **templates aprovados** (marketing, utilidade ou autenticação, nos mesmos valores da tabela acima). Texto livre é rejeitado pela Meta.

### Regra 3 — A janela reseta a cada mensagem do cliente

Se o cliente responde, mesmo que 23 horas depois, a contagem de 24h começa de novo. Um cliente engajado, que troca mensagens com frequência, mantém a janela sempre aberta — mas isso não muda o fato de que cada mensagem sua continua sendo cobrada.

### Regra 4 — Cada categoria de template abre sua própria janela

Se o cliente responde a um template de marketing, isso abre uma janela de 24h "de marketing". Se no mesmo dia ele responde também a um lembrete de utilidade, isso abre uma segunda janela, independente da primeira. Na prática, você pode ter mais de uma janela de 24h aberta com o mesmo cliente ao mesmo tempo — o sistema deve rastrear isso por categoria, não de forma genérica.

### Regra 5 — Autenticação é sempre cobrada

Templates de autenticação (OTP) são cobrados mesmo em situações em que tudo mais estaria grátis (por exemplo, dentro da janela de 72h de anúncios — ver seção 5). Existe também a variante `AUTHENTICATION_INTERNATIONAL`, usada quando o destinatário está fora do país da conta, que pode custar mais — vale checar qual subcategoria está configurada na conta da Cia da Vacina.

### Regra 6 — Meta Business Agent é cobrado por token, sempre

Diferente de humano ou IA própria (que são cobrados por mensagem), o Meta Business Agent — a IA nativa da Meta — é cobrado por **token consumido**, a aproximadamente $2 por 1 milhão de tokens. Uma mensagem típica consome de 10 mil a 35 mil tokens, o que dá entre R$ 0,02 e R$ 0,07 por resposta, dependendo da complexidade.

Essa cobrança **não tem exceção**: acontece dentro da janela de 24h, fora dela, e mesmo dentro da janela gratuita de 72h de Click-to-WhatsApp (ver seção 5). Só a "entrega" da mensagem fica grátis nesses casos — o processamento da IA da Meta é cobrado sempre.

### Regra 7 — Descontos por volume existem, mas só para alguns tipos

Templates de utilidade e autenticação ficam mais baratos conforme o volume mensal cresce (faixas por país). **Templates de marketing e mensagens de texto livre (service messages) não têm desconto por volume** — custam o mesmo valor não importa quanto você envie. Isso significa que reduzir a quantidade de mensagens de texto livre é a única forma de baixar esse custo específico.

---

## 2. Tabela de Referência de Custos {#tabela-custos}

| Situação | O que enviar | Custo estimado |
|---|---|---|
| Cliente escreveu há < 24h, resposta simples | Texto livre (humano ou IA própria) | R$ 0,035 |
| Cliente escreveu há < 24h, situação padronizável | Template de utilidade | R$ 0,0068 |
| Cliente não escreve há > 24h | Só template (qualquer categoria) | conforme categoria |
| Login / verificação (OTP) | Template de autenticação | R$ 0,0068 (sempre) |
| Campanha / promoção | Template de marketing | R$ 0,31 (sempre) |
| Cliente veio de anúncio Click-to-WhatsApp, respondido em até 24h após o clique | Qualquer tipo, durante 72h | R$ 0,00 (exceto Business Agent) |
| Meta Business Agent responde, em qualquer contexto acima | — | R$ 0,02 a R$ 0,07 por token consumido |

---

## 3. Arquitetura de Atendimento em 3 Camadas {#arquitetura}

A lógica para manter o custo baixo é simples: **usar a camada mais barata que resolve o atendimento**, e só escalar para a próxima quando necessário.

```
ENTRADA DO CLIENTE
    │
    ├─ [CAMADA 1 — Grátis] Cliente veio de um anúncio Click-to-WhatsApp?
    │  └─ SIM → janela de 72h grátis (templates e texto livre)
    │  └─ Meta Business Agent AINDA cobra por token aqui
    │
    ├─ [CAMADA 2 — Barato] IA resolve o atendimento?
    │  └─ Template de utilidade (FAQ) → R$ 0,0068
    │  └─ Meta Business Agent → R$ 0,02–0,07 por token
    │  └─ Se resolver, encerra aqui
    │
    └─ [CAMADA 3 — Mais caro] Precisa de atendente humano
       └─ Texto livre pelo CRM → R$ 0,035 por mensagem
       └─ Ou template de utilidade, se aplicável → R$ 0,0068
       └─ Usar apenas quando as camadas 1 e 2 não resolvem
```

Todo o atendimento acontece dentro do CRM — não existe canal alternativo. A forma de controlar custo na Camada 3 não é evitar o CRM, é **reduzir quantas mensagens o atendente manda por atendimento** (ver seção 8).

---

## 4. Estratégia de Templates {#templates}

Templates pré-aprovados são, no geral, mais baratos e previsíveis que texto livre — a orientação é estruturar o atendimento para usá-los sempre que a situação for padronizável.

### Marketing (R$ 0,31 — o mais caro)

Use para: promoções, campanhas sazonais, reengajamento de clientes inativos.

```
Exemplo: "🎉 Agosto é mês de imunização! Agende sua vacina da gripe
com 20% de desconto até dia 31. Responda AGENDAR para garantir."
```

Por ser o template mais caro, reserve para campanhas com ROI comprovado (CTWA costuma ser mais eficiente para esse tipo de mensagem — ver seção 5).

### Utilidade (R$ 0,0068 — o mais barato entre os pagos)

Use para: confirmações de agendamento, lembretes, respostas a dúvidas frequentes (FAQ).

```
Exemplo: "✅ Agendamento confirmado! {{1}} às {{2}}, unidade {{3}}.
Leve RG ou CNH. Qualquer dúvida, é só chamar."
```

Esse é o template que deve concentrar o maior volume de mensagens — é a categoria com melhor custo-benefício.

### Autenticação (R$ 0,0068 — sempre cobrado)

Use exclusivamente para códigos de verificação (OTP) e login.

```
Exemplo: "Seu código de verificação Cia da Vacina é {{1}}.
Válido por 5 minutos. Não compartilhe."
```

### Checklist de Aprovação de Templates

```
☐ Texto claro, sem ambiguidade
☐ Variáveis ({{1}}, {{2}}) sempre preenchidas, nunca vazias
☐ Sem URLs encurtadas (Meta rejeita)
☐ Sem caracteres especiais fora do padrão
☐ Categoria correta (marketing ≠ utilidade ≠ autenticação)
☐ Linguagem formal, mas natural — sem parecer spam
```

### Erros Comuns Que Causam Rejeição

- Classificar como "utilidade" um conteúdo que é promocional (a Meta reclassifica e cobra como marketing, ou rejeita)
- Deixar variáveis sem contexto (`{{1}}` sozinho, sem explicar o que preencher)
- Usar linguagem muito próxima de outro template já rejeitado (Meta identifica padrão)

---

## 5. Campanhas Click-to-WhatsApp — Entrada Gratuita de 72h {#ctwa}

### O Que É

Um anúncio no Facebook ou Instagram com botão "Enviar mensagem", que abre uma conversa no WhatsApp. Se a empresa responder em até 24 horas após o clique, se ativa uma janela de **72 horas totalmente gratuita** — templates e texto livre, sem custo (a única exceção é o Meta Business Agent, que cobra por token mesmo aqui).

### Como Funciona

```
Cliente clica no anúncio → manda mensagem no WhatsApp
    │
    ├─ Empresa responde em até 24h → ativa 72h grátis
    │  (a partir do momento da resposta, não do clique)
    │
    └─ Empresa não responde em 24h → a janela grátis nunca ativa,
       volta ao modelo padrão (Regras 1 e 2 da seção 1)
```

Durante as 72h: templates de marketing, utilidade, autenticação e texto livre saem de graça. Só o Meta Business Agent continua sendo cobrado por token.

### Estratégias de Campanha Para a Cia da Vacina

**Campanha contínua — "Agendar Rápido":** anúncio sempre ativo, direcionado para quem busca vacinação, com resposta automática imediata pra garantir a ativação da janela de 72h.

**Campanha sazonal:** picos de demanda (ex: campanha de gripe em maio-junho, reforços no fim do ano), com orçamento e copy específicos pro período.

**Campanha de reengajamento:** direcionada a clientes que já agendaram antes mas não voltaram, trazendo-os de volta pela janela grátis de 72h em vez de mandar um template de marketing pago.

### Por Que Vale a Pena

Considerando que templates de marketing custam R$ 0,31 cada, e uma campanha CTWA bem feita pode gerar o mesmo volume de conversas por R$ 0,00 (fora o investimento em mídia), o retorno costuma ser significativamente melhor que broadcasts tradicionais — principalmente quando a taxa de conversão em agendamento é alta.

---

## 6. Meta Business Agent {#business-agent}

### O Que É

A IA nativa da Meta para WhatsApp, capaz de responder dúvidas, qualificar leads e escalar para humano quando necessário. Cobrada por token consumido (Regra 6, seção 1).

### Quando Usar

Bom para: triagem inicial, perguntas frequentes, qualificação de lead antes de passar pro humano.

Menos indicado para: casos que envolvem julgamento clínico ou emergência (mesmo que a IA consiga responder tecnicamente, a política da Cia da Vacina deve exigir humano nesses casos).

### Arquitetura de Resposta Recomendada

```
Cliente inicia conversa
    │
    ├─ Pergunta simples (horário, preço, documentos)?
    │  └─ Meta Business Agent responde (R$ 0,02–0,04)
    │
    ├─ Precisa qualificar (qual vacina, qual unidade, disponibilidade)?
    │  └─ Meta Business Agent conduz (R$ 0,03–0,05)
    │  └─ Ao final, dispara template de confirmação (R$ 0,0068)
    │
    └─ Reação, emergência, ou pedido de falar com humano?
       └─ Escala direto para atendente (ver seção 8)
```

### Configuração — Passos Principais

1. **Setup inicial:** ativar o Meta Business Agent na conta WhatsApp Business Platform da Cia da Vacina
2. **Base de conhecimento:** alimentar com FAQ, tabela de preços, horários de funcionamento, documentos aceitos, política de cancelamento
3. **Prompts e instruções:** definir o tom de voz (acolhedor, mas objetivo) e os limites (nunca dar orientação médica, sempre escalar reações/emergências)
4. **Rotas de transferência:** configurar quando e como transferir para um atendente humano, preservando o contexto da conversa

### Custo vs. Benefício

Para 100 conversas por dia, comparando as abordagens:

```
Só humano (texto livre): ~3 mensagens/conversa × R$ 0,035
  = R$ 10,50/dia → R$ 315/mês

Meta Business Agent + escalação humana quando necessário:
  80% resolvido por IA/templates (custo baixo, misto)
  20% escala pra humano (1 mensagem de fechamento)
  ≈ R$ 3,70/dia → R$ 112/mês

Economia estimada: ~R$ 200/mês usando IA + templates como
primeira linha, reservando o humano pro que realmente precisa.
```

---

## 7. Arquitetura Técnica do CRM {#arquitetura-tecnica}

### Conceitos Que o Backend Precisa Implementar

**1. Fluxo de integração:** webhook recebe a mensagem do cliente → CRM classifica o tipo de atendimento → decide a camada (IA/template/humano) → registra o canal e o custo no histórico.

**2. Sistema de templates:** cadastro de templates aprovados por categoria (marketing/utilidade/autenticação), com validação de variáveis antes do envio, evitando rejeição.

**3. Rastreamento de janelas de contexto:** por Regra 4 (seção 1), o sistema precisa rastrear uma janela de 24h **por categoria de template**, não uma janela única por cliente. Isso afeta diretamente o cálculo de custo de cada envio.

**4. Cálculo de custo:** núcleo do sistema que, a cada mensagem enviada, determina a categoria (texto livre / utilidade / autenticação / marketing / Business Agent), verifica se está dentro de janela grátis (CTWA 72h) e registra o valor cobrado.

**5. Integração com Meta Business Agent:** chamadas à API do agente, captura de tokens consumidos por resposta, e rota de escalação para fila humana.

**6. Analytics de campanhas CTWA:** rastreamento de cliques, ativação da janela de 72h, conversão em agendamento e ROI por campanha.

**7. Histórico de mensagens:** schema de banco que registra, para cada mensagem, o tipo, a categoria, o custo, o atendente (se humano) ou o modelo de IA (se automatizado), e o momento de envio — necessário tanto para o dashboard de custos quanto para auditoria de conformidade.

**8. Alertas e monitoramento:** ver seção 10.

Tudo isso precisa rodar em tempo real (idealmente abaixo de 100ms de latência no processamento do webhook) e com zero downtime, já que mensagens perdidas viram atendimento perdido.

---

## 8. Treinamento de Atendentes — Como Gastar Menos {#treinamento}

Este é o material de treinamento para os atendentes das 5 unidades da Cia da Vacina. Como todo o atendimento passa pelo CRM, a única alavanca de economia nas mãos do atendente é **quantas mensagens ele manda por atendimento** — por isso todas as estratégias abaixo giram em torno disso.

**Regra simples para fixar:** *"Uma mensagem completa vale mais que três picadas."*

### Ordem de Prioridade

```
1º) A IA ou um template resolve sozinho?
    └─ SIM → deixa resolver. Não interfira.

2º) Existe um template pronto pra essa situação?
    └─ SIM → use o template (mais barato e previsível).

3º) Precisa mesmo de texto livre?
    └─ Escreva tudo numa mensagem só, completa.
```

### Estratégia 1 — Nunca Fragmentar (a "Regra dos 3 em 1")

Cada mensagem separada em texto livre é uma cobrança separada. A mesma informação, mandada de uma vez, custa uma fração do preço.

```
❌ Fragmentado (4 mensagens = 4x custo)
"Oi!"
"Deixa eu ver a agenda"
"Temos horário quarta às 14h"
"Serve pra você?"

✅ Consolidado (1 mensagem = 1x custo)
"Oi! Temos horário quarta-feira às 14h, serve pra você?"
```

```
❌ Fragmentado (5 mensagens)
"Poxa, que pena!"
"Vou cancelar aqui"
"Pronto, cancelado"
"Quer remarcar pra outro dia?"
"Ou prefere só cancelar mesmo?"

✅ Consolidado (1 mensagem)
"Poxa, que pena! Já cancelei seu agendamento. Quer que eu já
deixe marcado outro dia, ou prefere só cancelar por enquanto?"
```

**Na prática:** se você está digitando uma segunda mensagem seguida sem o cliente ter respondido nada no meio, pare e junte as duas antes de enviar.

### Estratégia 2 — WhatsApp Flows Para Coletar Vários Dados de Uma Vez

Quando o atendimento precisa recolher várias informações (nome, idade, vacina desejada, unidade, turno), perguntar uma por uma gera de 5 a 8 mensagens cobradas. O WhatsApp tem um recurso chamado **Flows**: um formulário interativo que abre dentro do chat, o cliente preenche tudo de uma vez e envia — conta como uma única transação.

```
❌ Pergunta por pergunta (7 mensagens)
"Qual seu nome completo?" / "Qual sua idade?" / "Qual vacina
você precisa?" / "Qual unidade prefere?" / "Manhã ou tarde?" /
"Convênio ou particular?" / "Perfeito, confirmando..."

✅ Flow (1 transação)
Cliente toca em "Preencher dados" → formulário abre no chat →
preenche tudo → envia → CRM recebe estruturado, pronto pra agendar
```

Vale priorizar isso no roadmap do CRM para os fluxos de primeira triagem e novo cadastro.

### Estratégia 3 — Botões e Listas Interativas

Para decisões simples (escolher entre 2-4 opções), use botões de resposta rápida ou listas — mensagens interativas nativas do WhatsApp que contam como uma única mensagem.

```
❌ Texto livre em várias mensagens (3 mensagens)
"Você prefere qual unidade?" / "1 - Centro, 2 - Zona Norte,
3 - Zona Sul" / "Manda o número da opção"

✅ Botões interativos (1 mensagem)
"Qual unidade você prefere?"
( Centro )  ( Zona Norte )  ( Zona Sul )
```

### Estratégia 4 — Biblioteca de Respostas Rápidas no CRM

Em vez do atendente digitar a resposta na hora (o que naturalmente sai fragmentado), o CRM deve oferecer respostas prontas e completas por assunto, que o atendente seleciona e ajusta.

```
[DOCUMENTOS]
"Aceitamos RG, CNH ou passaporte (para estrangeiros). Você já
tem algum desses em mãos?"

[MEDO DE AGULHA - CRIANÇA]
"Entendo! 💙 Temos técnicas especiais para crianças com medo.
Qual a idade dele(a)? Posso te passar a melhor abordagem."

[REAÇÃO LEVE - NÃO EMERGÊNCIA]
"Isso costuma ser normal e passa em 1-2 dias. Se piorar ou vier
febre alta, procure atendimento médico. Posso ajudar em algo?"

[REAÇÃO GRAVE - EMERGÊNCIA]
"⚠️ Procure o hospital mais próximo AGORA. Me conta o que
aconteceu que eu já registro aqui pro acompanhamento."
```

### Estratégia 5 — Cortesias Vão Junto, Não Separadas

```
❌ Separado (2 mensagens)
"Confirmado para quarta às 14h!"
"Qualquer coisa é só chamar 😊"

✅ Junto (1 mensagem)
"Confirmado para quarta às 14h! Qualquer coisa é só chamar 😊"
```

### Meta de Mensagens por Atendimento

| Tipo de atendimento | Meta de mensagens (humano) |
|---|---|
| Agendamento simples | 1-2 mensagens |
| Dúvida (documentos, horário, preço) | 1 mensagem |
| Cancelamento/remarcação | 1-2 mensagens |
| Reação/emergência | 2-3 mensagens (qualidade acima de economia) |
| Reclamação | 2-3 mensagens (empatia justifica um pouco mais) |

**Importante:** essa meta é referência de treinamento, não regra rígida. Em casos de saúde ou reclamação, o atendente nunca deve cortar comunicação por causa de custo — o bem-estar do cliente vem primeiro.

### Checklist de Onboarding — Novo Atendente

```
☐ Entende a diferença entre template e resposta livre
☐ Sabe que toda mensagem de texto livre no CRM tem custo, e por
  isso vale juntar tudo numa mensagem
☐ Conhece a biblioteca de respostas rápidas do CRM
☐ Sabe usar botões/listas interativas para perguntas de múltipla escolha
☐ Treinou a "regra dos 3 em 1"
☐ Sabe reconhecer emergência médica (nunca economizar mensagem aqui)
☐ Sabe transferir para farmacêutico/médico responsável
☐ Fez 5 atendimentos supervisionados antes de atender sozinho
```

---

## 9. Procedimentos Operacionais Diários {#procedimentos}

### Rotina — Início do Dia

```
☐ Abrir dashboard do CRM Querência
☐ Verificar fila de mensagens recebidas
☐ Notar quais estão marcadas como "PRECISA HUMANO"
☐ Iniciar atendimento com prioridade:
   1. Reações/problemas de saúde
   2. Confirmações de agendamento
   3. Dúvidas gerais
```

**Exemplo de fila no dashboard:**

```
╔════════════════════════════════════════════════════╗
║           CRM — FILA DE ATENDIMENTO                ║
╠════════════════════════════════════════════════════╣
║  🔴 3 MENSAGENS AGUARDANDO (Prioridade Alta)        ║
║                                                      ║
║  1. João Silva | 09:15                              ║
║     "Tive coceira depois da vacina, normal?"         ║
║     [REAÇÃO MÉDICA] → Transferir para Dr. Pedro      ║
║                                                      ║
║  2. Maria Santos | 09:32                             ║
║     "Qual horário de amanhã disponível?"             ║
║     [Meta Business Agent já qualificou]              ║
║     Status: Aguardando confirmação                   ║
║                                                      ║
║  3. Carlos Oliveira | 10:05                          ║
║     "Preciso de comprovante de vacinação"             ║
║     [DOCUMENTAÇÃO] → Buscar no sistema                ║
║                                                      ║
║  🟢 15 conversas resolvidas hoje (IA)                ║
║  🟡 2 transferências pendentes                        ║
╚════════════════════════════════════════════════════╝
```

### Regras de Ouro Para Responder Mensagens

**Regra 1: Nunca responda algo que a IA já respondeu corretamente.** Se o histórico mostra que o Meta Business Agent já esclareceu a dúvida, não repita a informação manualmente — isso é uma cobrança desnecessária.

**Regra 2: Se precisar responder em texto livre, junte tudo numa única mensagem completa.**

```
CENÁRIO — Cliente: "Qual documento preciso levar?"

❌ Errado: responder manualmente do zero quando a IA já respondeu
✅ Correto: verificar se o Meta Business Agent já respondeu com o
   FAQ. Se sim, só confirme ou detalhe exceções. Se não, responda
   uma vez, de forma completa.

CENÁRIO — Cliente: "Tive uma reação estranha"

✅ Correto: avaliar se é emergência.
   Se sim: "PROCURE HOSPITAL AGORA" + registrar o relato.
   Se não: acionar o profissional de saúde responsável, sendo
   empático mas objetivo.
```

### Quando Usar Template vs. Resposta Livre

| Situação | Tipo | Custo |
|---|---|---|
| Agendamento confirmado | Template (confirmacao_agendamento) | GRÁTIS dentro da janela de 24h |
| Lembrete dia anterior | Template (lembrete_consulta) | R$ 0,0068 |
| OTP de login | Template (otp_verificacao) | R$ 0,0068 |
| Reação médica | Resposta livre (chamar profissional) | R$ 0,035 |
| Dúvida sobre documentos | Template FAQ; se falhar, resposta livre | GRÁTIS / R$ 0,035 |
| Reclamação | Resposta livre (empatia) | R$ 0,035 |
| Escolha entre poucas opções | Botão/lista interativa | R$ 0,035 (1 mensagem resolve vários passos) |

**Regra prática:** sempre que existir um template pronto pra situação, use-o. Quando precisar de resposta livre, capriche pra sair tudo numa mensagem só.

---

## 10. Monitoramento de Custos {#monitoramento}

### Dashboard Recomendado

```
╔═══════════════════════════════════════════════════════════╗
║         QUERÊNCIA CRM — MONITOR DE CUSTOS META             ║
╠═══════════════════════════════════════════════════════════╣
║  📊 RESUMO DO MÊS                                          ║
║  Mensagens enviadas:     4.250                             ║
║  Custo total Meta:       R$ 892,45                         ║
║  Custo/mensagem (média): R$ 0,21                           ║
║  Orçamento mensal:       R$ 1.000                          ║
║  Margem restante:        R$ 107,55                         ║
║                                                              ║
║  💰 BREAKDOWN POR TIPO                                     ║
║  Templates marketing:    850 msgs × R$ 0,31  = R$ 263,50   ║
║  Templates utilidade:  2.100 msgs × R$ 0,0068 = R$ 14,28   ║
║  Texto livre (humano):   800 msgs × R$ 0,035 = R$ 28,00    ║
║  Meta Business Agent:    500 msgs × R$ 0,045 = R$ 22,50    ║
║  Click-to-WhatsApp:      600 msgs × R$ 0,00  = R$ 0,00     ║
║                          (dentro da janela de 72h)          ║
║                                                              ║
║  📈 MÉDIA DE MENSAGENS POR ATENDIMENTO                     ║
║  Média atual: 2,3 mensagens/atendimento (meta: ≤ 2)         ║
║  Atendimentos acima da meta: 18% do total                  ║
║  Economia potencial se todos ficarem na meta: ~R$ 45/mês    ║
║                                                              ║
║  📣 CAMPANHAS CTWA                                          ║
║  Vakão de Agosto: 580 cliques | ROI 340% | 145 agendamentos║
║  Reengajamento:   280 cliques | ROI 210% |  58 agendamentos║
║                                                              ║
║  ⚠️  ALERTAS                                                ║
║  🟡 18% dos atendimentos acima da meta de mensagens         ║
║     Sugestão: reforçar treinamento de consolidação          ║
║  🟢 CTWA com performance excelente — considerar aumentar     ║
║     investimento                                             ║
║  🟢 Meta Business Agent com taxa de escalação de 15%         ║
║     (dentro do esperado)                                     ║
╚═══════════════════════════════════════════════════════════╝
```

### Checklist Semanal

```
SEGUNDA-FEIRA (10:00) — Revisar semana anterior

☐ Total de conversas: _____ (meta: 500+)
☐ Custo total da semana: R$ _____ (orçamento: R$ 250)
☐ Taxa de resolução pela IA: ____% (meta: 75%+)
☐ Taxa de escalação para humano: ____% (meta: 25% máx)
☐ Média de mensagens por atendimento humano: ____ (meta: ≤ 2)
☐ Templates criados/aprovados: _____
☐ Mensagens rejeitadas: _____ (zero é ideal)
☐ CTWA ROI: ____% (meta: 200%+)
☐ Tempo médio de resposta: __ min (meta: < 5 min)

Se algum número está abaixo da meta:
→ Documentar razão → Propor solução → Ajustar próxima semana
```

### Alertas Automáticos a Configurar

```
ALERTA 1 — Gastos acima do previsto
Trigger: custo diário > R$ 40 → Notificar gerente, revisar campanhas

ALERTA 2 — Taxa de erro em templates
Trigger: rejeições > 5/dia → Notificar dev, revisar qualidade

ALERTA 3 — Meta Business Agent não escalando
Trigger: tempo de espera > 5 min → Verificar se o agente está
respondendo

ALERTA 4 — Baixa performance de CTWA
Trigger: conversão < 10% → Revisar copy do anúncio, testar variações

ALERTA 5 — Estouro de orçamento do mês
Trigger: 80% do orçamento gasto antes do dia 20 → Pausar campanhas
pagas, usar apenas CTWA

ALERTA 6 — Mensagens fragmentadas acima do normal
Trigger: média de mensagens/atendimento > 3 em um dia → Reforçar
treinamento de consolidação com o atendente/unidade específica
```

---

## 11. Conformidade e Boas Práticas {#conformidade}

### LGPD — Lei Geral de Proteção de Dados

```
ANTES DE ENVIAR QUALQUER MENSAGEM:
☐ Cliente aceitou receber comunicações? 
☐ Cliente optou por não receber? Respeitar
☐ Dados sensíveis? Nunca enviar por template — dados médicos
  são tratados apenas por humano
☐ Pode se descadastrar a qualquer momento — incluir
  "Responda SAIR para descadastrar"
```

### Políticas da Meta

```
❌ Proibido:
- Spam (mensagens em volume sem consentimento)
- IA genérica de propósito geral no lugar do Meta Business Agent
- Phishing ou coleta indevida de dados
- Mensagens fora da janela sem template
- Enviar antes de obter consentimento

✅ Permitido:
- Triagem com o Meta Business Agent
- Templates pré-aprovados
- Respostas dentro da janela de 24h
- Campanhas CTWA (o próprio clique já gera consentimento)
- Qualificação de leads
```

### Checklist de Auditoria (Quinzenal)

```
☐ Nenhuma mensagem enviada fora de template quando exigido
☐ Nenhum cliente reclamando de spam
☐ Taxa de rejeição de mensagens < 5%
☐ Nenhuma tentativa de phishing detectada
☐ Documentação de consentimento organizada
☐ Respostas respeitam horários (não enviar de madrugada)
☐ Meta Business Agent não está "escondendo" o acesso a humano
☐ Histórico de mensagens legível e completo

Se algo falhar: parar imediatamente → revisar procedimento →
reportar ao time de desenvolvimento
```

---

## 12. Suporte e Troubleshooting {#suporte}

**"Minha mensagem foi rejeitada pela Meta"**
```
1. Verificar categoria: marketing precisa de template de
   marketing, não de utilidade
2. Revisar variáveis: {{1}}, {{2}} preenchidas e no formato certo?
3. Revisar texto: sem URLs encurtadas, sem caracteres estranhos
4. Se persistir: clonar o template, renomear (ex: _v2) e resubmeter
```

**"O atendente respondeu e não deveria ter custado o que custou"**
```
1. Cliente respondeu há menos de 24h? Confirma se está dentro
   da janela certa
2. Foi texto livre ou template? Confirma a classificação
3. Se houver erro de cobrança: registrar reclamação com a Meta
   (reembolso costuma levar ~48h) e reforçar treinamento
```

**"Meta Business Agent não está respondendo"**
```
☐ Agente está ativado nas configurações do WhatsApp?
☐ Base de conhecimento está configurada?
☐ Cliente está em Android/iOS? (desktop não é suportado)
☐ Há erros de API nos logs?
☐ Token de acesso expirou?

Solução: renovar token → verificar webhooks → testar com número
de teste → reiniciar o serviço se necessário
```

**"Campanhas Click-to-WhatsApp não estão gerando leads"**
```
1. O anúncio está sendo exibido? (checar no Ads Manager)
2. Há cliques acontecendo? (se sim, o problema é no WhatsApp)
3. A mensagem inicial é clara?

Melhorar copy: "Clique para agendar vacinação em 2 min" converte
mais que "Quer falar conosco?"

Melhorar abertura: "Que tipo de vacina você precisa?" converte
mais que "Olá, como posso ajudar?"

Se o problema for volume: considerar aumentar o orçamento
diário e medir o resultado por 7 dias antes de mudar de novo.
```

---

## 13. Cronograma de Implementação {#cronograma}

**Semana 1 — Preparação**
- [ ] Criar templates (marketing, utilidade, autenticação)
- [ ] Submeter para aprovação da Meta
- [ ] Configurar o Meta Business Agent em ambiente de teste
- [ ] Testar com número interno

**Semana 2 — Deploy**
- [ ] Ativar templates aprovados em produção
- [ ] Ligar o Meta Business Agent
- [ ] Configurar webhooks e rotas de escalação
- [ ] Rastrear janelas de contexto por categoria (Regra 4, seção 1)

**Semana 3 — Campanhas CTWA**
- [ ] Criar a primeira campanha ("Agendar Rápido")
- [ ] Definir orçamento inicial de teste
- [ ] Monitorar conversões
- [ ] Testar variações de copy

**Semana 4 — Treinamento de Atendentes**
- [ ] Rodar o módulo "Treinamento de Atendentes — Como Gastar Menos"
      com todas as 5 unidades
- [ ] Implementar a biblioteca de respostas rápidas no CRM
- [ ] Ativar a métrica "média de mensagens por atendimento" no dashboard
- [ ] Auditoria de 10 conversas por atendente na primeira semana de uso

**Semana 5 — Otimização Contínua**
- [ ] Analisar dados reais de custo
- [ ] Ajustar orçamento de CTWA conforme o ROI
- [ ] Revisar templates com base nas rejeições e no volume real
- [ ] Documentar aprendizados e ajustar metas do dashboard

---

## Suporte Contínuo

**Contato Querência Software:**
📧 contato@felipemeneguzzi.dev · 📱 54 99701-4602 · 🏢 Marau/RS

**Contato Meta (escalação urgente):**
Dashboard: facebook.com/developers (WhatsApp Settings)
Fórum: developers.facebook.com/community/whatsapp

---

*Documento preparado para Cia da Vacina LTDA como parte do projeto CRM Omnichannel.*
*Confidencial — uso restrito à Cia da Vacina e Querência Software.*
