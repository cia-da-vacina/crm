import { system } from "./core";

export const typography = system({
  fontFamily: { property: "fontFamily", scale: "fonts" },
  fontSize: { property: "fontSize", scale: "fontSizes" },
  fontWeight: { property: "fontWeight", scale: "fontWeights" },
  lineHeight: { property: "lineHeight", scale: "lineHeights" },
  letterSpacing: { property: "letterSpacing", scale: "letterSpacings" },
  textAlign: true,
  fontStyle: true,
  textTransform: true,
  textDecoration: true,
});

export type TypographyProps = {
  fontFamily?: string;
  fontSize?: number | string;
  fontWeight?: number | string;
  lineHeight?: number | string;
  letterSpacing?: number | string;
  textAlign?: string;
  fontStyle?: string;
  textTransform?: string;
  textDecoration?: string;
};
