import styled, { css } from "styled-components";
import type { ButtonHTMLAttributes, ReactNode } from "react";

export type ButtonVariant = "primary" | "secondary" | "ghost" | "danger";
export type ButtonSize = "sm" | "md" | "lg";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  size?: ButtonSize;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
  fullWidth?: boolean;
};

const iconSlotByButtonSize: Record<ButtonSize, string> = {
  sm: "14px",
  md: "16px",
  lg: "18px",
};

const sizes = {
  sm: css`
    height: 32px;
    padding: 0 ${({ theme }) => theme.space[3]};
    font-size: ${({ theme }) => theme.fontSizes.sm};
    gap: ${({ theme }) => theme.space[2]};
  `,
  md: css`
    height: 36px;
    padding: 0 ${({ theme }) => theme.space[4]};
    font-size: ${({ theme }) => theme.fontSizes.md};
    gap: ${({ theme }) => theme.space[2]};
  `,
  lg: css`
    height: 40px;
    padding: 0 ${({ theme }) => theme.space[5]};
    font-size: ${({ theme }) => theme.fontSizes.lg};
    gap: ${({ theme }) => theme.space[2]};
  `,
};

const variants = {
  primary: css`
    background: ${({ theme }) => theme.colors["button.primary.bg"]};
    color: ${({ theme }) => theme.colors["button.primary.text"]};
    &:hover:not(:disabled) {
      background: ${({ theme }) => theme.colors["button.primary.bg.hover"]};
    }
  `,
  secondary: css`
    background: ${({ theme }) => theme.colors["button.secondary.bg"]};
    color: ${({ theme }) => theme.colors["button.secondary.text"]};
    border-color: ${({ theme }) => theme.colors["button.secondary.border"]};
    &:hover:not(:disabled) {
      background: ${({ theme }) => theme.colors["button.secondary.bg.hover"]};
    }
  `,
  ghost: css`
    background: transparent;
    color: ${({ theme }) => theme.colors["button.ghost.text"]};
    border-color: transparent;
    &:hover:not(:disabled) {
      background: ${({ theme }) => theme.colors["button.ghost.bg.hover"]};
    }
  `,
  danger: css`
    background: ${({ theme }) => theme.colors["button.danger.bg"]};
    color: ${({ theme }) => theme.colors["button.danger.text"]};
    &:hover:not(:disabled) {
      background: ${({ theme }) => theme.colors["button.danger.bg.hover"]};
    }
  `,
};

const IconSlot = styled.span<{ $slot: string }>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: ${({ $slot }) => $slot};
  height: ${({ $slot }) => $slot};
  flex-shrink: 0;
  line-height: 0;

  & > svg {
    width: ${({ $slot }) => $slot} !important;
    height: ${({ $slot }) => $slot} !important;
    max-width: 100% !important;
    max-height: 100% !important;
  }
`;

const StyledButton = styled.button<{
  $variant: ButtonVariant;
  $size: ButtonSize;
  $fullWidth?: boolean;
}>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid transparent;
  border-radius: ${({ theme }) => theme.radii.sm};
  font-family: ${({ theme }) => theme.fonts.body};
  font-weight: ${({ theme }) => theme.fontWeights.medium};
  cursor: pointer;
  transition: background 120ms ease, border-color 120ms ease, transform 80ms ease;
  width: ${({ $fullWidth }) => ($fullWidth ? "100%" : "auto")};
  ${({ $size }) => sizes[$size]}
  ${({ $variant }) => variants[$variant]}

  &:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  &:active:not(:disabled) {
    transform: translateY(0.5px);
  }

  &:focus-visible {
    outline: none;
    box-shadow: ${({ theme }) => theme.shadows.focus};
  }
`;

export default function Button({
  variant = "primary",
  size = "md",
  leftIcon,
  rightIcon,
  fullWidth,
  children,
  type = "button",
  ...rest
}: ButtonProps) {
  const slot = iconSlotByButtonSize[size];
  return (
    <StyledButton
      type={type}
      $variant={variant}
      $size={size}
      $fullWidth={fullWidth}
      {...rest}
    >
      {leftIcon ? <IconSlot $slot={slot}>{leftIcon}</IconSlot> : null}
      {children}
      {rightIcon ? <IconSlot $slot={slot}>{rightIcon}</IconSlot> : null}
    </StyledButton>
  );
}
