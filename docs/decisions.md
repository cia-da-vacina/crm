# Decisões fechadas (dúvidas do plano)

**Data:** 2026-07-14  
**Regra:** onde o stakeholder ainda não respondeu, a decisão abaixo é a **default operacional** para desbloquear Sprint 1. Mudanças futuras atualizam este arquivo e o backlog.

| # | Dúvida | Decisão | Tipo |
|---|--------|---------|------|
| D-01 | 1 número por unidade ou número único? | **1 número WhatsApp Business por unidade** (N `phone_number_id` sob 1 WABA). Roteamento automático pela linha de entrada. | DECISÃO DEFAULT |
| D-02 | Provedor LLM + DPA/LGPD? | **OpenAI API** (`gpt-4o-mini` ou equivalente barato) atrás de feature flag `ai.enabled`. Sem treinar modelos com dados do cliente. Contrato DPA a ser assinado pelo negócio antes do go-live com PII real. | DECISÃO DEFAULT |
| D-03 | Conteúdo oficial dos POPs? | Seed com **POPs placeholder** editáveis no admin. Conteúdo real entra via workshop com operação (não bloqueia código). | DECISÃO DEFAULT |
| D-04 | SLA primeira resposta? | Meta operacional **5 minutos** em horário comercial (KPI soft no dashboard; sem cobrança automática). | DECISÃO DEFAULT |
| D-05 | Integração agenda/prontuário? | **Não no MVP.** Nenhum conector clínico. | DECISÃO DEFAULT |
| D-06 | Catálogo motivos não conversão? | Seed inicial (abaixo). Admin pode ativar/desativar. | DECISÃO DEFAULT |
| D-07 | Volume msgs/dia? | Assumir **até ~2.000 msgs/dia** total nas 5 unidades. Arquitetura monólito + Postgres suficiente. | SUPOSIÇÃO |
| D-08 | Hospedagem? | **AWS** default: ECS/Fargate ou EC2 simples + RDS Postgres + (opcional) S3 mídia V1.1. Alternativa VPS aceitável se custo for prioridade. | DECISÃO DEFAULT |

## Catálogo de intenções (IA triagem)

| code | label |
|------|-------|
| agendar | Agendamento / vacinas |
| precos | Preços / pacotes |
| duvidas | Dúvidas gerais |
| reclamacao | Reclamação / suporte |
| outro | Não classificado |

## Catálogo de motivos de não conversão (seed)

| code | label |
|------|-------|
| preco | Preço elevado |
| concorrente | Foi para concorrente |
| sem_retorno | Cliente sem retorno |
| prazo | Sem disponibilidade de agenda |
| nao_interesse | Perdeu o interesse |
| outro | Outro (detalhar em texto) |

## Auth

- Email/senha (sem SSO) — mantido.
- JWT access (15 min) + refresh (7 dias).

## Tempo real / mídia

- SSE para inbox.
- MVP **texto-first**; mídia WhatsApp em V1.1.

## Retenção LGPD

- **24 meses** de mensagens/conversas (config `lgpd.retention_months`).
