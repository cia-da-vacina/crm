import { get, system } from "./core";

const defaults = [0, 4, 8, 12, 16, 24, 32, 48, 64, 96, 128];

const isNumber = (n: unknown): n is number =>
  typeof n === "number" && !Number.isNaN(n);

const getMargin = (n: unknown, scale: unknown) => {
  if (!isNumber(n)) return get(scale, n as string, n);
  const negative = n < 0;
  const absolute = Math.abs(n);
  const value = get(scale, absolute, absolute);
  if (!isNumber(value)) return negative ? `-${value}` : value;
  return value * (negative ? -1 : 1);
};

export const space = system({
  m: { property: "margin", scale: "space", transform: getMargin, defaultScale: defaults },
  mt: { property: "marginTop", scale: "space", transform: getMargin, defaultScale: defaults },
  mr: { property: "marginRight", scale: "space", transform: getMargin, defaultScale: defaults },
  mb: { property: "marginBottom", scale: "space", transform: getMargin, defaultScale: defaults },
  ml: { property: "marginLeft", scale: "space", transform: getMargin, defaultScale: defaults },
  mx: {
    property: ["marginLeft", "marginRight"],
    scale: "space",
    transform: getMargin,
    defaultScale: defaults,
  },
  my: {
    property: ["marginTop", "marginBottom"],
    scale: "space",
    transform: getMargin,
    defaultScale: defaults,
  },
  p: { property: "padding", scale: "space", defaultScale: defaults },
  pt: { property: "paddingTop", scale: "space", defaultScale: defaults },
  pr: { property: "paddingRight", scale: "space", defaultScale: defaults },
  pb: { property: "paddingBottom", scale: "space", defaultScale: defaults },
  pl: { property: "paddingLeft", scale: "space", defaultScale: defaults },
  px: {
    property: ["paddingLeft", "paddingRight"],
    scale: "space",
    defaultScale: defaults,
  },
  py: {
    property: ["paddingTop", "paddingBottom"],
    scale: "space",
    defaultScale: defaults,
  },
  margin: { property: "margin", scale: "space", transform: getMargin, defaultScale: defaults },
  marginTop: { property: "marginTop", scale: "space", transform: getMargin, defaultScale: defaults },
  marginRight: { property: "marginRight", scale: "space", transform: getMargin, defaultScale: defaults },
  marginBottom: { property: "marginBottom", scale: "space", transform: getMargin, defaultScale: defaults },
  marginLeft: { property: "marginLeft", scale: "space", transform: getMargin, defaultScale: defaults },
  padding: { property: "padding", scale: "space", defaultScale: defaults },
  paddingTop: { property: "paddingTop", scale: "space", defaultScale: defaults },
  paddingRight: { property: "paddingRight", scale: "space", defaultScale: defaults },
  paddingBottom: { property: "paddingBottom", scale: "space", defaultScale: defaults },
  paddingLeft: { property: "paddingLeft", scale: "space", defaultScale: defaults },
  gap: { property: "gap", scale: "space", defaultScale: defaults },
  rowGap: { property: "rowGap", scale: "space", defaultScale: defaults },
  columnGap: { property: "columnGap", scale: "space", defaultScale: defaults },
});

export type SpaceProps = {
  m?: number | string;
  mt?: number | string;
  mr?: number | string;
  mb?: number | string;
  ml?: number | string;
  mx?: number | string;
  my?: number | string;
  p?: number | string;
  pt?: number | string;
  pr?: number | string;
  pb?: number | string;
  pl?: number | string;
  px?: number | string;
  py?: number | string;
  margin?: number | string;
  marginTop?: number | string;
  marginRight?: number | string;
  marginBottom?: number | string;
  marginLeft?: number | string;
  padding?: number | string;
  paddingTop?: number | string;
  paddingRight?: number | string;
  paddingBottom?: number | string;
  paddingLeft?: number | string;
  gap?: number | string;
  rowGap?: number | string;
  columnGap?: number | string;
};
