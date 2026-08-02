# Documentação — CRM Cia da Vacina

## Fonte da verdade (produto)

| Documento | Uso |
|-----------|-----|
| [`APPROVED-SCOPE.md`](./APPROVED-SCOPE.md) | Escopo aprovado e mudanças de escopo |
| [`PRODUCT-V2.md`](./PRODUCT-V2.md) | Visão de produto, personas e jornadas |
| [`BACKEND-CONTRACT.md`](./BACKEND-CONTRACT.md) | Contrato REST completo (domínio CRM) |
| [`FRONTEND-ARCHITECTURE.md`](./FRONTEND-ARCHITECTURE.md) | Arquitetura do frontend (BFF, proxy, PWA) |

## Complementares

| Documento | Uso |
|-----------|-----|
| [`openapi.yaml`](./openapi.yaml) | OpenAPI formal — Auth, Users, Units, Me (parcial; o restante está no BACKEND-CONTRACT) |
| [`decisions.md`](./decisions.md) | Defaults operacionais (LLM, seeds, JWT, LGPD) |
| [`adr/0001-simple-monolith.md`](./adr/0001-simple-monolith.md) | ADR: monólito Go + Next + Postgres |

A implementação Go vive em [`../backend/`](../backend/) (hoje só README + contrato acima).

## Marketing / comercial

Artefatos de deck, proposta e esboço contratual ficam em [`marketing/`](./marketing/) — não fazem parte da fonte da verdade de engenharia.
