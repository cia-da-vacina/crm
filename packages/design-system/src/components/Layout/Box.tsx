import styled from "styled-components";
import type { HTMLAttributes } from "react";
import {
  border,
  color,
  flexbox,
  layout,
  position,
  shadow,
  space,
  typography,
  shouldForwardProp,
  type BorderProps,
  type ColorProps,
  type FlexboxProps,
  type LayoutProps,
  type PositionProps,
  type ShadowProps,
  type SpaceProps,
  type TypographyProps,
} from "@cia-da-vacina/styled-system";

export type BoxProps = SpaceProps &
  ColorProps &
  LayoutProps &
  FlexboxProps &
  BorderProps &
  PositionProps &
  ShadowProps &
  TypographyProps &
  HTMLAttributes<HTMLDivElement> & {
    as?: React.ElementType;
    cursor?: string;
  };

const Box = styled.div.withConfig({ shouldForwardProp })<BoxProps>`
  box-sizing: border-box;
  min-width: 0;
  ${space}
  ${color}
  ${layout}
  ${flexbox}
  ${border}
  ${position}
  ${shadow}
  ${typography}
  ${({ cursor }) => (cursor ? `cursor: ${cursor};` : "")}
`;

export default Box;
