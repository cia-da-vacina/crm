"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { CiaThemeProvider } from "@cia-da-vacina/design-system";
import { useEffect, useState, type ReactNode } from "react";
import { AuthProvider } from "@/contexts/auth-context";
import { InstallPrompt } from "@/components/install-prompt";
import StyledComponentsRegistry from "@/lib/styled-components-registry";

async function enableMocking() {
  if (process.env.NEXT_PUBLIC_USE_MOCKS !== "true") return;

  // Drop a previous next-pwa worker so it does not steal /api fetches from MSW.
  if (typeof navigator !== "undefined" && "serviceWorker" in navigator) {
    const regs = await navigator.serviceWorker.getRegistrations();
    await Promise.all(
      regs
        .filter((r) => !r.active?.scriptURL.includes("mockServiceWorker"))
        .map((r) => r.unregister()),
    );
  }

  const { worker } = await import("@/mocks/browser");
  await worker.start({
    onUnhandledRequest: "bypass",
    quiet: true,
    serviceWorker: {
      url: "/mockServiceWorker.js",
    },
  });
}

export function Providers({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(process.env.NEXT_PUBLIC_USE_MOCKS !== "true");
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: { retry: 1, refetchOnWindowFocus: false },
        },
      }),
  );

  useEffect(() => {
    enableMocking().then(() => setReady(true));
  }, []);

  return (
    <StyledComponentsRegistry>
      <CiaThemeProvider>
        {!ready ? null : (
          <QueryClientProvider client={queryClient}>
            <AuthProvider>
              {children}
              <InstallPrompt />
            </AuthProvider>
          </QueryClientProvider>
        )}
      </CiaThemeProvider>
    </StyledComponentsRegistry>
  );
}
