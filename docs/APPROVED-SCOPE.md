# Escopo — Aprovado

**Data original:** 2026-07-14
**Revisão:** 2026-08-02 — expansão de canais (mudança de escopo em relação ao MVP original)
**Produto:** CRM Cia da Vacina
**Referência:** [spec.md](./spec.md) · [PRODUCT-V2.md](./PRODUCT-V2.md) · [BACKEND-CONTRACT.md](./BACKEND-CONTRACT.md)

## Decisões do stakeholder

| Decisão | Valor |
|---------|--------|
| Canais | **WhatsApp + Instagram + Facebook Messenger** (Meta Cloud API / Messaging APIs) — **AGORA INCLUÍDOS**, ver nota de mudança de escopo abaixo |
| Papel da IA | **Triagem inicial** (saudação, intenção, roteamento) + **handoff humano obrigatório** — mantido, válido para todos os canais |
| Unidades | 5 unidades Cia da Vacina |
| Equipe | Felipe (Go/DB/APIs) · Cristian (Next.js/UI) |

## Nota de mudança de escopo (2026-08-02)

O MVP aprovado em 2026-07-14 previa **somente WhatsApp**. Nesta revisão, o escopo foi ampliado para incluir **Instagram e Facebook (Messenger)** como canais de primeira classe, incluindo:

- Mensagens diretas (DM) em Instagram e Messenger, na mesma fila de Inbox do WhatsApp.
- **Engagements Meta-nativos** (fora do fluxo 1:1): resposta a story, menção em story, comentário em publicação, comentário em live — tratados em fila própria (`/engagements`), com opção de responder, descartar ou converter em conversa.
- Identidade de cliente unificada no CRM (`Customer.id`), já que a Meta não fornece um ID cross-platform — ver decisão detalhada em [`PRODUCT-V2.md`](./PRODUCT-V2.md#4-decisão-de-id-de-cliente-entre-canais) e contrato técnico em [`BACKEND-CONTRACT.md`](./BACKEND-CONTRACT.md#3-customers-crm-id--identidades-por-canal).

A regra de **triagem por IA antes de qualquer resposta humana** e o **handoff obrigatório** (`claim`) foram mantidos sem alteração e agora se aplicam igualmente a WhatsApp, Instagram e Messenger.

## Incluído no escopo atual

- Autenticação, RBAC e multiunidade
- Inbox multicanal — WhatsApp, Instagram, Facebook Messenger (receber/enviar texto)
- Fila de Engagements — story reply, story mention, comentário de post, comentário de live (**NOW INCLUDED**)
- Identidade de cliente unificada (`Customer` + `CustomerIdentity` por canal), com **telefone E.164 como chave de merge** — WhatsApp já identificado; Instagram/Messenger anônimos até a parede exigir telefone (**NOW INCLUDED**)
- IA de triagem com kill-switch e fallback, válida em todos os canais
- Handoff humano (claim/assign)
- Pipeline: Em atendimento → Em negociação → Aguardando fechamento → Fechado / Não fechado
- Motivos de não conversão
- Templates WhatsApp (janela 24h)
- Follow-up básico
- POPs / scripts
- Dashboard mínimo (unidade + consolidado), agora segmentado por canal
- Auditoria básica + settings Meta multicanal (tokens mascarados por canal)
- Frontend de produção como PWA instalável, sem dependência de mocks (MSW removido)

## Explicitamente fora do escopo

- App mobile nativo (iOS/Android publicado em loja) — coberto pela instalação como PWA
- IA conversacional contínua (a IA segue restrita à triagem inicial + handoff)
- Discador, e-mail, SMS
- BI avançado / exportações complexas
- Microserviços, Kafka, Redis obrigatório
- SSO
- Integração com agenda/prontuário clínico

## Critério de aceite do escopo

Este documento + `docs/spec.md` + `docs/PRODUCT-V2.md` são a fonte da verdade. Mudanças de escopo exigem atualização explícita aqui e no backlog.
