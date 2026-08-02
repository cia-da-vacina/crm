"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import Link from "next/link";
import {
  Badge,
  Button,
  DataList,
  DataListRow,
  EmptyState,
  Flex,
  PageHeader,
  Spinner,
  Stack,
  StageBadge,
  Text,
  useToast,
} from "@cia-da-vacina/design-system";
import { ClockIcon } from "@cia-da-vacina/icon-system";
import { fadeUp, staggerContainer, staggerItem } from "@/lib/motion";
import { useAuth } from "@/providers/auth-provider";
import { followupsService } from "@/services";

export default function FollowUpsPage() {
  const { activeUnitId } = useAuth();
  const qc = useQueryClient();
  const toast = useToast();
  const { data, isLoading } = useQuery({
    queryKey: ["followups", activeUnitId],
    queryFn: () =>
      followupsService.list({ unit_id: activeUnitId ?? undefined, status: "open" }),
    enabled: Boolean(activeUnitId),
  });

  const complete = useMutation({
    mutationFn: (id: string) => followupsService.complete(id),
    onSuccess: async () => {
      toast.push("Follow-up concluído", "success");
      await qc.invalidateQueries({ queryKey: ["followups"] });
      await qc.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });

  return (
    <motion.div {...fadeUp} style={{ width: "100%" }}>
      <Stack gap={3} width="100%">
        <PageHeader
          title="Follow-ups"
          description="Retome clientes em aguardando fechamento / não fechado"
        />

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

        {(data?.items.length ?? 0) > 0 && (
          <motion.div variants={staggerContainer} initial="initial" animate="animate">
            <DataList>
              {(data?.items ?? []).map((f) => {
                const overdue = new Date(f.due_at) < new Date();
                return (
                  <motion.div key={f.id} variants={staggerItem}>
                    <DataListRow
                      interactive={false}
                      leading={
                        <Stack gap={1}>
                          <Flex gap={2} alignItems="center" flexWrap="wrap">
                            <Text fontWeight="semibold">{f.customer_name}</Text>
                            <StageBadge stage={f.pipeline_stage} />
                            {overdue ? (
                              <Badge tone="danger">Atrasado</Badge>
                            ) : (
                              <Badge tone="warning">Agendado</Badge>
                            )}
                          </Flex>
                          {f.customer_phone && (
                            <Text muted fontSize="sm">
                              {f.customer_phone}
                            </Text>
                          )}
                          <Text fontSize="sm">{f.note}</Text>
                          <Text fontSize="xs" muted>
                            Vencimento: {new Date(f.due_at).toLocaleString("pt-BR")}
                          </Text>
                        </Stack>
                      }
                      trailing={
                        <Flex gap={2} alignItems="center">
                          <Link href={`/inbox/${f.conversation_id}`}>
                            <Button variant="secondary" size="sm">
                              Abrir
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
                      }
                    />
                  </motion.div>
                );
              })}
            </DataList>
          </motion.div>
        )}
      </Stack>
    </motion.div>
  );
}
