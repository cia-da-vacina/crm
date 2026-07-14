# Design System — Cia da Vacina

Espelho estrutural do setup Sencon (4 repositórios), sob medida para o CRM.

| Pacote | Repo | Local |
|--------|------|-------|
| `@cia-da-vacina/styled-system` | https://github.com/cia-da-vacina/styled-system | `E:\Projetos-teste\CiaVacinaStyledSystem` |
| `@cia-da-vacina/design-system-tokens` | https://github.com/cia-da-vacina/design-system-tokens | `E:\Projetos-teste\CiaVacinaDesignSystemTokens` |
| `@cia-da-vacina/icon-system` | https://github.com/cia-da-vacina/icon-system | `E:\Projetos-teste\CiaVacinaIconSystem` |
| `@cia-da-vacina/design-system` | https://github.com/cia-da-vacina/design-system | `E:\Projetos-teste\CiaVacinaDesignSystem` |

## Fluxo

```
tokens (webLight) → ThemeProvider
styled-system → style props (m, bg, …)
icons → ícones tipados
design-system → Box, Button, ConversationRow, AppShell…
     ↓
CRM frontend (Next.js)
```

## Direção visual

- Evergreen clínico (marca)
- Pipeline stages com cores semânticas próprias
- Fraunces (display) + DM Sans (UI)

## Uso local no CRM

Deps `file:../../CiaVacina*` no `frontend/package.json`. Após alterar um pacote:

```bash
cd ../CiaVacinaDesignSystem
npm run build
cd ../CIA-VACINA/frontend
npm install
```
