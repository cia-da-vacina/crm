import styled from "styled-components";
import {
  color,
  space,
  typography,
  shouldForwardProp,
  type ColorProps,
  type SpaceProps,
  type TypographyProps,
} from "@cia-da-vacina/styled-system";

export type HeadingProps = SpaceProps &
  ColorProps &
  TypographyProps & {
    as?: "h1" | "h2" | "h3" | "h4";
    display?: boolean;
  };

const Heading = styled.h2.withConfig({ shouldForwardProp })<HeadingProps>`
  margin: 0;
  font-family: ${({ theme, display }) =>
    display
      ? (theme?.fonts?.display ?? "Georgia, serif")
      : (theme?.fonts?.body ?? "inherit")};
  font-weight: ${({ theme }) => theme?.fontWeights?.semibold ?? 600};
  letter-spacing: ${({ theme }) => theme?.letterSpacings?.tight ?? "-0.02em"};
  color: ${({ theme }) => theme?.colors?.["text.primary"] ?? "#1B2420"};
  font-size: ${({ theme, as }) => {
    if (as === "h1") return theme?.fontSizes?.["3xl"] ?? "28px";
    if (as === "h3") return theme?.fontSizes?.xl ?? "18px";
    if (as === "h4") return theme?.fontSizes?.lg ?? "16px";
    return theme?.fontSizes?.["2xl"] ?? "22px";
  }};
  line-height: ${({ theme }) => theme?.lineHeights?.tight ?? 1.2};
  ${space}
  ${color}
  ${typography}
`;

export default Heading;
