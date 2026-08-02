"use client";

import type { ReactNode } from "react";
import { ToastProvider } from "@cia-da-vacina/design-system";
import { InstallPrompt } from "@/components/install-prompt";
import StyledComponentsRegistry from "@/lib/styled-components-registry";
import { AuthProvider } from "./auth-provider";
import { NavigationFeedbackProvider } from "./navigation-provider";
import { QueryProvider } from "./query-provider";
import { ThemeModeProvider } from "./theme-provider";
import { NavigationProgress } from "@/components/navigation-progress";

/** Root provider tree for the whole app — composed once here so `layout.tsx` stays trivial. */
export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <StyledComponentsRegistry>
      <ThemeModeProvider>
        <ToastProvider>
          <QueryProvider>
            <AuthProvider>
              <NavigationFeedbackProvider>
                <NavigationProgress />
                {children}
                <InstallPrompt />
              </NavigationFeedbackProvider>
            </AuthProvider>
          </QueryProvider>
        </ToastProvider>
      </ThemeModeProvider>
    </StyledComponentsRegistry>
  );
}
