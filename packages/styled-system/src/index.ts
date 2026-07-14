export { get, system, compose, variant } from "./core";
export type { Theme, StyleFn } from "./core";

export { space } from "./space";
export type { SpaceProps } from "./space";

export { color } from "./color";
export type { ColorProps } from "./color";

export { layout } from "./layout";
export type { LayoutProps } from "./layout";

export { flexbox } from "./flexbox";
export type { FlexboxProps } from "./flexbox";

export { border } from "./border";
export type { BorderProps } from "./border";

export { position } from "./position";
export type { PositionProps } from "./position";

export { typography } from "./typography";
export type { TypographyProps } from "./typography";

export { shadow } from "./shadow";
export type { ShadowProps } from "./shadow";

export { shouldForwardProp, shouldForwardPropDiv } from "./shouldForwardProp";

import { compose } from "./core";
import { space } from "./space";
import { color } from "./color";
import { layout } from "./layout";
import { flexbox } from "./flexbox";
import { border } from "./border";
import { position } from "./position";
import { typography } from "./typography";
import { shadow } from "./shadow";
import { system } from "./core";

export const cursor = system({ cursor: true });
export type CursorProps = { cursor?: string };

export const all = compose(
  space,
  color,
  layout,
  flexbox,
  border,
  position,
  typography,
  shadow,
  cursor,
);
