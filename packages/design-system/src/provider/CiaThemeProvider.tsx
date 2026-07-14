import type { ReactNode } from "react";
import { ThemeProvider } from "styled-components";
import { GlobalStyle, webLight } from "@cia-da-vacina/design-system-tokens";

export type CiaThemeProviderProps = {
  children: ReactNode;
};

export default function CiaThemeProvider({ children }: CiaThemeProviderProps) {
  return (
    <ThemeProvider theme={webLight}>
      <GlobalStyle />
      {children}
    </ThemeProvider>
  );
}
