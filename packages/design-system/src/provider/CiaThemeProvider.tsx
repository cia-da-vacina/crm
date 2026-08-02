import type { ReactNode } from "react";
import { ThemeProvider, type DefaultTheme } from "styled-components";
import {
  GlobalStyle,
  webDark,
  webLight,
} from "@cia-da-vacina/design-system-tokens";

export type CiaThemeMode = "light" | "dark";

export type CiaThemeProviderProps = {
  children: ReactNode;
  /** Defaults to light (Cia evergreen CRM look on top of Untitled UI tokens). */
  mode?: CiaThemeMode;
};

export default function CiaThemeProvider({
  children,
  mode = "light",
}: CiaThemeProviderProps) {
  const theme = (mode === "dark" ? webDark : webLight) as DefaultTheme;
  return (
    <ThemeProvider theme={theme}>
      <GlobalStyle />
      {children}
    </ThemeProvider>
  );
}
