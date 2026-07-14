import styled from "styled-components";
import type { InputHTMLAttributes, ReactNode } from "react";
import Stack from "../Layout/Stack";
import Text from "../Typography/Text";

export type TextFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  hint?: string;
  error?: string;
  leftAddon?: ReactNode;
};

const FieldShell = styled.div<{ $invalid?: boolean }>`
  display: flex;
  align-items: center;
  gap: ${({ theme }) => theme.space[2]};
  min-height: 36px;
  padding: 0 ${({ theme }) => theme.space[3]};
  background: ${({ theme }) => theme.colors["input.bg"]};
  border: 1px solid
    ${({ theme, $invalid }) =>
      $invalid ? theme.colors["border.danger"] : theme.colors["input.border"]};
  border-radius: ${({ theme }) => theme.radii.sm};
  transition: border-color 120ms ease, box-shadow 120ms ease;

  &:hover {
    border-color: ${({ theme, $invalid }) =>
      $invalid ? theme.colors["border.danger"] : theme.colors["input.border.hover"]};
  }

  &:focus-within {
    border-color: ${({ theme }) => theme.colors["input.border.focus"]};
    box-shadow: ${({ theme }) => theme.shadows.focus};
  }
`;

const Input = styled.input`
  flex: 1;
  min-width: 0;
  border: 0;
  outline: none;
  background: transparent;
  color: ${({ theme }) => theme.colors["input.text"]};
  font-size: ${({ theme }) => theme.fontSizes.md};

  &::placeholder {
    color: ${({ theme }) => theme.colors["input.placeholder"]};
  }

  &:disabled {
    cursor: not-allowed;
  }
`;

export default function TextField({
  label,
  hint,
  error,
  leftAddon,
  id,
  ...rest
}: TextFieldProps) {
  const inputId = id ?? rest.name;
  return (
    <Stack gap={1}>
      {label && (
        <Text as="label" htmlFor={inputId} fontSize="xs" fontWeight="medium" muted>
          {label}
        </Text>
      )}
      <FieldShell $invalid={Boolean(error)}>
        {leftAddon}
        <Input id={inputId} {...rest} />
      </FieldShell>
      {(error || hint) && (
        <Text fontSize="xs" color={error ? "text.danger" : "text.muted"}>
          {error ?? hint}
        </Text>
      )}
    </Stack>
  );
}
