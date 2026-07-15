import raw from "../../../tokens/global/colors";

/**
 * Cia CRM overlay — applied ON TOP of Untitled UI semantic colors.
 * Keeps product identity (evergreen) + pipeline/mode tokens the CRM already uses.
 */
const crmOverlay = {
  // backgrounds (Cia naming used by components)
  "bg.canvas": raw.mist[25],
  "bg.surface": raw.base.white,
  "bg.surface.muted": raw.mist[50],
  "bg.surface.raised": raw.base.white,
  "bg.brand.subtle": raw.evergreen[50],
  "bg.brand.solid": raw.evergreen[600],
  "bg.brand.hover": raw.evergreen[700],
  "bg.danger.subtle": raw.coral[50],
  "bg.danger.solid": raw.coral[500],
  "bg.warning.subtle": raw.amber[50],
  "bg.warning.solid": raw.amber[400],
  "bg.info.subtle": raw.sky[50],
  "bg.info.solid": raw.sky[500],

  // text
  "text.primary": raw.mist[900],
  "text.secondary": raw.mist[600],
  "text.muted": raw.mist[500],
  "text.inverse": raw.base.white,
  "text.brand": raw.evergreen[700],
  "text.danger": raw.coral[600],
  "text.warning": raw.amber[700],
  "text.success": raw.evergreen[600],
  "text.link": raw.evergreen[600],

  // borders
  "border.subtle": raw.mist[100],
  "border.default": raw.mist[200],
  "border.strong": raw.mist[300],
  "border.brand": raw.evergreen[400],
  "border.danger": raw.coral[300],
  "border.focus": raw.evergreen[500],

  // button (Cia naming)
  "button.primary.bg": raw.evergreen[600],
  "button.primary.bg.hover": raw.evergreen[700],
  "button.primary.bg.active": raw.evergreen[800],
  "button.primary.text": raw.base.white,
  "button.secondary.bg": raw.base.white,
  "button.secondary.bg.hover": raw.mist[50],
  "button.secondary.border": raw.mist[200],
  "button.secondary.text": raw.mist[800],
  "button.ghost.bg.hover": raw.evergreen[50],
  "button.ghost.text": raw.evergreen[700],
  "button.danger.bg": raw.coral[500],
  "button.danger.bg.hover": raw.coral[600],
  "button.danger.text": raw.base.white,

  // input
  "input.bg": raw.base.white,
  "input.border": raw.mist[200],
  "input.border.hover": raw.mist[300],
  "input.border.focus": raw.evergreen[500],
  "input.placeholder": raw.mist[400],
  "input.text": raw.mist[900],
  "input.disabled.bg": raw.mist[50],

  // pipeline / CRM stages
  "stage.em_atendimento.bg": raw.sky[100],
  "stage.em_atendimento.text": raw.sky[600],
  "stage.em_negociacao.bg": raw.evergreen[100],
  "stage.em_negociacao.text": raw.evergreen[700],
  "stage.aguardando_fechamento.bg": raw.amber[100],
  "stage.aguardando_fechamento.text": raw.amber[700],
  "stage.fechado.bg": raw.evergreen[50],
  "stage.fechado.text": raw.evergreen[800],
  "stage.nao_fechado.bg": raw.coral[100],
  "stage.nao_fechado.text": raw.coral[700],

  // mode badges
  "mode.ai.bg": raw.sand[100],
  "mode.ai.text": raw.sand[500],
  "mode.human.bg": raw.evergreen[50],
  "mode.human.text": raw.evergreen[700],

  // nav
  "nav.bg": "rgba(255,255,255,0.88)",
  "nav.item.active.bg": raw.evergreen[50],
  "nav.item.active.text": raw.evergreen[700],
  "nav.item.text": raw.mist[600],
} as const;

export default crmOverlay;
