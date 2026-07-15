import type { ReactNode } from "react";
import { ThemeProvider } from "styled-components";
import {
  GlobalStyle,
  webDark,
  webLight,
} from "@cia-da-vacina/design-system-tokens";

export type CiaThemeMode = "light" | "dark";

export type CiaThemeProviderProps = {
  children: ReactNode;
  /** Defaults to light (Cia evergreen CRM look on top of Sencon tokens). */
  mode?: CiaThemeMode;
};

export default function CiaThemeProvider({
  children,
  mode = "light",
}: CiaThemeProviderProps) {
  const theme = mode === "dark" ? webDark : webLight;
  return (
    <ThemeProvider theme={theme}>
      <GlobalStyle />
      {children}
    </ThemeProvider>
  );
}
