# Backend — CRM Cia da Vacina

Pasta reservada para a API Go. A implementação de demonstração foi removida de propósito para o Felipe montar do jeito dele.

## Contrato (fonte da verdade)

Implementar conforme:

| Doc | Conteúdo |
|-----|----------|
| [`../docs/BACKEND-CONTRACT.md`](../docs/BACKEND-CONTRACT.md) | Contrato REST completo do domínio |
| [`../docs/openapi.yaml`](../docs/openapi.yaml) | OpenAPI (Auth, Users, Units, Me) |
| [`../docs/PRODUCT-V2.md`](../docs/PRODUCT-V2.md) | Produto, jornadas, identidade cross-canal |
| [`../docs/APPROVED-SCOPE.md`](../docs/APPROVED-SCOPE.md) | Escopo aprovado |
| [`../docs/decisions.md`](../docs/decisions.md) | Defaults (JWT, seeds, LLM, LGPD) |
| [`../docs/adr/0001-simple-monolith.md`](../docs/adr/0001-simple-monolith.md) | Monólito Go + Postgres |

Índice geral: [`../docs/README.md`](../docs/README.md).

## Expectativa do frontend

O Next.js (BFF) chama a API em `API_URL` (default local: `http://localhost:8080/api/v1`). Prefixos e cookies: ver [`../docs/FRONTEND-ARCHITECTURE.md`](../docs/FRONTEND-ARCHITECTURE.md).

Enquanto esta pasta não tiver servidor rodando, o frontend sobe, mas login e dados falham — esperado.
