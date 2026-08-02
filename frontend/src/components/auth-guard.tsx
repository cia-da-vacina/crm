"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import { Flex, Spinner, Text } from "@cia-da-vacina/design-system";
import { useAuth } from "@/providers/auth-provider";

export function AuthGuard({ children }: { children: ReactNode }) {
  const { user, loading, logout } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (loading || user) return;

    // Clear cookies before navigating — otherwise middleware still sees a
    // refresh/access cookie and bounces /login → /inbox forever.
    let cancelled = false;
    void (async () => {
      try {
        await logout();
      } finally {
        if (!cancelled) router.replace("/login");
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [loading, user, logout, router]);

  if (loading || !user) {
    return (
      <Flex minHeight="100vh" alignItems="center" justifyContent="center" gap={2}>
        <Spinner />
        <Text muted>Verificando sessão…</Text>
      </Flex>
    );
  }

  return <>{children}</>;
}
