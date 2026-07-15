import styled from "styled-components";
import {
  color,
  space,
  shouldForwardProp,
  type ColorProps,
  type SpaceProps,
} from "@cia-da-vacina/styled-system";
import type { ReactNode } from "react";

/** Predefined sizes — do not pass raw pixels in app code */
export const iconSizes = {
  xs: 12,
  sm: 14,
  md: 16,
  lg: 20,
  xl: 24,
} as const;

export type IconSizeToken = keyof typeof iconSizes;

export type IconProps = SpaceProps &
  ColorProps & {
    size?: IconSizeToken;
    /** Theme color token or CSS color; applied as stroke (Untitled UI style). */
    fill?: string;
    className?: string;
    title?: string;
  };

const Svg = styled.svg.withConfig({
  shouldForwardProp: (prop: string) =>
    prop !== "size" && shouldForwardProp(prop),
})<{ $px: number } & IconProps>`
  display: block;
  flex-shrink: 0;
  width: ${({ $px }) => $px}px !important;
  height: ${({ $px }) => $px}px !important;
  max-width: ${({ $px }) => $px}px;
  max-height: ${({ $px }) => $px}px;
  ${space}
  ${color}

  /* Untitled UI stroke style: force stroke icons, never filled blobs */
  & * {
    fill: none !important;
    stroke: inherit;
  }
`;

export function createIcon(displayName: string, path: ReactNode) {
  function Icon({ size = "md", fill = "currentColor", ...rest }: IconProps) {
    const px = iconSizes[size] ?? iconSizes.md;
    return (
      <Svg
        $px={px}
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        width={px}
        height={px}
        aria-hidden
        // Resolve theme tokens via styled-system (`stroke` scale → colors)
        stroke={fill}
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
        {...rest}
      >
        {path}
      </Svg>
    );
  }
  Icon.displayName = displayName;
  return Icon;
}
