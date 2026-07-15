# @cia-da-vacina/styled-system

Engine de style props para o Design System da Cia da Vacina (inspirado no styled-system).

Resolve props como `m={4}`, `bg="bg.brand.solid"`, `fontSize="md"` contra o tema injetado pelo ThemeProvider dos tokens.

```tsx
import styled from "styled-components";
import { space, color, layout, shouldForwardProp } from "@cia-da-vacina/styled-system";

const Box = styled.div.withConfig({ shouldForwardProp })`
  ${space}
  ${color}
  ${layout}
`;
```

## Publish

GitHub Packages (`@cia-da-vacina` scope).
