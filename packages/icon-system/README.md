# @cia-da-vacina/icon-system

Ícones SVG do CRM Cia da Vacina.

## Uso

```tsx
import { InboxIcon, BotIcon } from "@cia-da-vacina/icon-system";

<InboxIcon size="lg" fill="text.brand" />
<BotIcon size="md" fill="mode.ai.text" />
```

## Tamanhos (`IconSizeToken`)

| Token | px |
|-------|----|
| `xs`  | 12 |
| `sm`  | 14 |
| `md`  | 16 (default) |
| `lg`  | 20 |
| `xl`  | 24 |

Não passe números crutos — o `styled-system` trata `size` numérico como token de layout.
