"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import { Flex, Spinner, Text } from "@cia-da-vacina/design-system";
import { useAuth } from "@/contexts/auth-context";

export function AuthGuard({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) {
      router.replace("/login");
    }
  }, [loading, user, router]);

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
