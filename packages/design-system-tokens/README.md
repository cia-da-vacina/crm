# @cia-da-vacina/design-system-tokens

Tokens e tema semântico do Design System da **Cia da Vacina** — pensado para o CRM de atendimento WhatsApp (inbox, pipeline, multiunidade).

## Uso

```tsx
import { ThemeProvider } from "styled-components";
import { webLight, GlobalStyle } from "@cia-da-vacina/design-system-tokens";

<ThemeProvider theme={webLight}>
  <GlobalStyle />
  {children}
</ThemeProvider>
```

## Direção visual

- Evergreen (marca clínica / confiança)
- Mist neutros para UI densa
- Amber / coral para urgência e perda no funil
- Fraunces (display) + DM Sans (UI)

## Keys semânticas úteis

`bg.brand.solid`, `text.secondary`, `border.default`,  
`stage.em_negociacao.bg`, `mode.ai.text`, `button.primary.bg`
