"use client";

import styled from "styled-components";
import { Moon01, Sun } from "@cia-da-vacina/icon-system";
import { useThemeMode } from "@/providers/theme-provider";

const Toggle = styled.button`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  margin: 0;
  padding: 0;
  border: 1px solid ${({ theme }) => theme.colors["border.subtle"]};
  border-radius: ${({ theme }) => theme.radii.md};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  color: ${({ theme }) => theme.colors["text.brand"]};
  cursor: pointer;
  transition: background 140ms ease, border-color 140ms ease, transform 120ms ease;

  &:hover {
    background: ${({ theme }) => theme.colors["bg.surface.muted"]};
    border-color: ${({ theme }) => theme.colors["border.default"]};
  }

  &:active {
    transform: scale(0.96);
  }

  &:focus-visible {
    outline: none;
    box-shadow: ${({ theme }) => theme.shadows.focus};
  }
`;

export function ThemeToggle() {
  const { mode, toggle } = useThemeMode();
  const isDark = mode === "dark";

  return (
    <Toggle
      type="button"
      onClick={toggle}
      aria-label={isDark ? "Ativar tema claro" : "Ativar tema escuro"}
      title={isDark ? "Tema claro" : "Tema escuro"}
    >
      {isDark ? <Sun size="sm" /> : <Moon01 size="sm" />}
    </Toggle>
  );
}
