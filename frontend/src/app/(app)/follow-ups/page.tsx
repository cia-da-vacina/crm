"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import {
  Badge,
  Box,
  Button,
  EmptyState,
  Flex,
  Heading,
  Spinner,
  Stack,
  StageBadge,
  Text,
} from "@cia-da-vacina/design-system";
import { ClockIcon } from "@cia-da-vacina/icon-system";
import { useAuth } from "@/contexts/auth-context";
import { api } from "@/lib/api";

export default function FollowUpsPage() {
  const { activeUnitId } = useAuth();
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ["followups", activeUnitId],
    queryFn: () =>
      api.listFollowUps({ unit_id: activeUnitId ?? undefined, status: "open" }),
    enabled: Boolean(activeUnitId),
  });

  const complete = useMutation({
    mutationFn: (id: string) => api.completeFollowUp(id),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["followups"] });
      await qc.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });

  return (
    <Stack gap={3} width="100%">
      <Stack gap={1} alignItems="center" style={{ textAlign: "center" }}>
        <Flex gap={2} alignItems="center" justifyContent="center">
          <ClockIcon size="lg" fill="text.brand" />
          <Heading as="h1" display>
            Follow-ups
          </Heading>
        </Flex>
        <Text muted fontSize="sm">
          Retome clientes em aguardando fechamento / não fechado
        </Text>
      </Stack>

      {isLoading && (
        <Flex gap={2} alignItems="center">
          <Spinner />
          <Text muted>Carregando fila…</Text>
        </Flex>
      )}

      {!isLoading && (data?.items.length ?? 0) === 0 && (
        <EmptyState
          icon={<ClockIcon size="xl" fill="text.muted" />}
          title="Fila vazia"
          description="Nenhum follow-up aberto para a unidade selecionada."
        />
      )}

      <Stack gap={0}>
        {(data?.items ?? []).map((f) => {
          const overdue = new Date(f.due_at) < new Date();
          return (
            <Box
              key={f.id}
              p={3}
              bg="bg.surface"
              borderWidth="hairline"
              borderStyle="solid"
              borderColor="border.default"
              borderRadius="md"
              mb={2}
            >
              <Flex justifyContent="space-between" gap={3} flexWrap="wrap">
                <Stack gap={2}>
                  <Flex gap={2} alignItems="center" flexWrap="wrap">
                    <Text fontWeight="semibold">{f.contact_name}</Text>
                    <StageBadge stage={f.pipeline_stage} />
                    {overdue ? (
                      <Badge tone="danger">Atrasado</Badge>
                    ) : (
                      <Badge tone="warning">Agendado</Badge>
                    )}
                  </Flex>
                  <Text muted fontSize="sm">
                    {f.contact_phone}
                  </Text>
                  <Text fontSize="sm">{f.note}</Text>
                  <Text fontSize="xs" muted>
                    Due: {new Date(f.due_at).toLocaleString("pt-BR")}
                  </Text>
                </Stack>
                <Flex gap={2} alignItems="flex-start">
                  <Link href={`/inbox/${f.conversation_id}`}>
                    <Button variant="secondary" size="sm">
                      Abrir conversa
                    </Button>
                  </Link>
                  <Button
                    size="sm"
                    onClick={() => complete.mutate(f.id)}
                    disabled={complete.isPending}
                  >
                    Concluir
                  </Button>
                </Flex>
              </Flex>
            </Box>
          );
        })}
      </Stack>
    </Stack>
  );
}
