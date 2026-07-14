import styled, { css } from "styled-components";
import type { SelectHTMLAttributes } from "react";
import Stack from "../Layout/Stack";
import Text from "../Typography/Text";

export type SelectOption = { value: string; label: string };
export type SelectFieldSize = "sm" | "md";

export type SelectFieldProps = Omit<
  SelectHTMLAttributes<HTMLSelectElement>,
  "size"
> & {
  label?: string;
  options: SelectOption[];
  fieldSize?: SelectFieldSize;
  /** Default true for forms; false keeps width to content (toolbar). */
  fullWidth?: boolean;
  /** Softer chrome for sticky headers / toolbars. */
  appearance?: "default" | "quiet";
};

const sizeStyles = {
  sm: css`
    min-height: 28px;
    padding: 0 ${({ theme }) => theme.space[2]};
    padding-right: ${({ theme }) => theme.space[5]};
    font-size: ${({ theme }) => theme.fontSizes.xs};
  `,
  md: css`
    min-height: 32px;
    padding: 0 ${({ theme }) => theme.space[3]};
    padding-right: ${({ theme }) => theme.space[6]};
    font-size: ${({ theme }) => theme.fontSizes.sm};
  `,
};

const appearances = {
  default: css`
    border: 1px solid ${({ theme }) => theme.colors["input.border"]};
    background-color: ${({ theme }) => theme.colors["input.bg"]};

    &:hover:not(:disabled) {
      border-color: ${({ theme }) => theme.colors["input.border.hover"]};
    }
  `,
  quiet: css`
    border: 1px solid transparent;
    background-color: transparent;

    &:hover:not(:disabled) {
      background-color: ${({ theme }) => theme.colors["bg.surface.muted"]};
      border-color: ${({ theme }) => theme.colors["border.subtle"]};
    }
  `,
};

const Select = styled.select<{
  $size: SelectFieldSize;
  $fullWidth: boolean;
  $appearance: "default" | "quiet";
}>`
  width: ${({ $fullWidth }) => ($fullWidth ? "100%" : "auto")};
  max-width: ${({ $fullWidth }) => ($fullWidth ? "none" : "13rem")};
  border-radius: ${({ theme }) => theme.radii.sm};
  color: ${({ theme }) => theme.colors["input.text"]};
  font-family: ${({ theme }) => theme.fonts.body};
  font-weight: ${({ theme }) => theme.fontWeights.medium};
  outline: none;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%236B7C76' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 8px center;
  ${({ $size }) => sizeStyles[$size]}
  ${({ $appearance }) => appearances[$appearance]}

  &:focus {
    border-color: ${({ theme }) => theme.colors["input.border.focus"]};
    box-shadow: ${({ theme }) => theme.shadows.focus};
  }
`;

export default function SelectField({
  label,
  options,
  id,
  fieldSize = "md",
  fullWidth = true,
  appearance = "default",
  ...rest
}: SelectFieldProps) {
  const selectId = id ?? rest.name;
  return (
    <Stack gap={1} style={fullWidth ? undefined : { width: "auto" }}>
      {label && (
        <Text as="label" htmlFor={selectId} fontSize="xs" muted>
          {label}
        </Text>
      )}
      <Select
        id={selectId}
        $size={fieldSize}
        $fullWidth={fullWidth}
        $appearance={appearance}
        {...rest}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </Select>
    </Stack>
  );
}
