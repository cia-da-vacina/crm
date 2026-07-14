import styled from "styled-components";
import type React from "react";
import {
  color,
  space,
  typography,
  shouldForwardProp,
  type ColorProps,
  type SpaceProps,
  type TypographyProps,
} from "@cia-da-vacina/styled-system";

export type TextProps = SpaceProps &
  ColorProps &
  TypographyProps &
  React.HTMLAttributes<HTMLElement> & {
    as?: React.ElementType;
    muted?: boolean;
    htmlFor?: string;
  };

const Text = styled.p.withConfig({ shouldForwardProp })<TextProps>`
  margin: 0;
  font-family: ${({ theme }) => theme?.fonts?.body ?? "inherit"};
  font-size: ${({ theme }) => theme?.fontSizes?.md ?? "14px"};
  line-height: ${({ theme }) => theme?.lineHeights?.normal ?? 1.5};
  color: ${({ theme, muted, color: colorProp }) =>
    colorProp
      ? undefined
      : muted
        ? (theme?.colors?.["text.secondary"] ?? "#4F6158")
        : (theme?.colors?.["text.primary"] ?? "#1B2420")};
  ${space}
  ${color}
  ${typography}
`;

export default Text;
