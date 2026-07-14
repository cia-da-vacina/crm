# @cia-da-vacina/design-system

Biblioteca de componentes React do CRM **Cia da Vacina**.

## Pacotes irmãos

| Package | Repo |
|---------|------|
| `@cia-da-vacina/design-system-tokens` | [design-system-tokens](https://github.com/cia-da-vacina/design-system-tokens) |
| `@cia-da-vacina/styled-system` | [styled-system](https://github.com/cia-da-vacina/styled-system) |
| `@cia-da-vacina/icon-system` | [icon-system](https://github.com/cia-da-vacina/icon-system) |
| `@cia-da-vacina/design-system` | this |

## Uso rápido

```tsx
import {
  CiaThemeProvider,
  AppShell,
  Button,
  ConversationList,
  ConversationRow,
} from "@cia-da-vacina/design-system";

export function App() {
  return (
    <CiaThemeProvider>
      <AppShell links={[{ href: "/inbox", label: "Inbox", active: true }]}>
        <ConversationList>
          <ConversationRow
            contactName="Maria Silva"
            preview="Quero agendar a vacina"
            stage="em_atendimento"
            mode="ai_triage"
            timestamp="agora"
          />
        </ConversationList>
        <Button>Assumir conversa</Button>
      </AppShell>
    </CiaThemeProvider>
  );
}
```

## Fontes

Carregue no app (Next.js):

```tsx
import { DM_Sans, Fraunces } from "next/font/google";
```
