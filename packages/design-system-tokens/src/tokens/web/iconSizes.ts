/** Named icon sizes (px). Prefer these over raw numbers. */
export const iconSizes = {
  xs: 12,
  sm: 14,
  md: 16,
  lg: 20,
  xl: 24,
} as const;

export type IconSizeToken = keyof typeof iconSizes;

export default iconSizes;
