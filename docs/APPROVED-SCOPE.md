# Escopo MVP — Aprovado

**Data:** 2026-07-14  
**Produto:** CRM Cia da Vacina  
**Referência:** [spec.md](./spec.md)

## Decisões do stakeholder

| Decisão | Valor |
|---------|--------|
| Canais no MVP | **Somente WhatsApp** (Meta Cloud API) — opção 1-A |
| Papel da IA | **Triagem inicial** (saudação, intenção, roteamento) + **handoff humano** — opção 2-B |
| Unidades | 5 unidades Cia da Vacina |
| Equipe | Felipe (Go/DB/APIs) · Cristian (Next.js/UI) |

## Incluído no MVP

- Autenticação, RBAC e multiunidade
- Inbox WhatsApp (receber/enviar texto)
- IA de triagem com kill-switch e fallback
- Handoff humano (claim/assign)
- Pipeline: Em atendimento → Em negociação → Aguardando fechamento → Fechado / Não fechado
- Motivos de não conversão
- Templates WhatsApp (janela 24h)
- Follow-up básico
- POPs / scripts
- Dashboard mínimo (unidade + consolidado)
- Auditoria básica + settings Meta

## Explicitamente fora do MVP

- Instagram / Facebook / outros canais
- IA conversacional contínua
- Discador, e-mail, SMS
- BI avançado / exportações complexas
- Microserviços, Kafka, Redis obrigatório
- App mobile nativo / SSO

## Critério de aceite do escopo

Este documento + `docs/spec.md` são a fonte da verdade. Mudanças de escopo exigem atualização explícita aqui e no backlog.
