"use client";

import { Flex, Spinner, Text } from "@cia-da-vacina/design-system";

/** Instant route-level fallback so nav clicks don't leave a frozen previous page. */
export default function AuthLoading() {
  return (
    <Flex minHeight="40vh" alignItems="center" justifyContent="center" gap={2}>
      <Spinner />
      <Text muted>Carregando…</Text>
    </Flex>
  );
}
