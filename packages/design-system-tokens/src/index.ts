export { default as webLight } from "./themes/web/light";
export type { WebLightTheme } from "./themes/web/light";
export { default as webDark } from "./themes/web/dark";
export type { WebDarkTheme } from "./themes/web/dark";

export { default as colors } from "./tokens/global/colors";
export { default as borderWidths } from "./tokens/global/borderWidths";
export { default as fontWeights } from "./tokens/global/fontWeights";
export { default as lineHeights } from "./tokens/global/lineHeights";

export { default as space } from "./tokens/web/space";
export { default as radii } from "./tokens/web/radii";
export { default as fonts } from "./tokens/web/fonts";
export { default as fontSizes } from "./tokens/web/fontSizes";
export { default as letterSpacings } from "./tokens/web/letterSpacings";
export { default as breakpoints } from "./tokens/web/breakpoints";
export { default as mediaQueries } from "./tokens/web/mediaQueries";
export { default as zIndices } from "./tokens/web/zIndices";
export { default as shadows } from "./tokens/web/shadows";
export { default as iconSizes } from "./tokens/web/iconSizes";
export type { IconSizeToken } from "./tokens/web/iconSizes";
export { default as GlobalStyle } from "./tokens/web/GlobalStyle";

export const pipelineStages = [
  "em_atendimento",
  "em_negociacao",
  "aguardando_fechamento",
  "fechado",
  "nao_fechado",
] as const;

export type PipelineStage = (typeof pipelineStages)[number];

export const stageLabels: Record<PipelineStage, string> = {
  em_atendimento: "Em atendimento",
  em_negociacao: "Em negociação",
  aguardando_fechamento: "Aguardando fechamento",
  fechado: "Fechado",
  nao_fechado: "Não fechado",
};
