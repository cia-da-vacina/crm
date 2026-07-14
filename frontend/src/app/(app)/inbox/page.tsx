"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import {
  ConversationList,
  ConversationRow,
  EmptyState,
  Flex,
  Heading,
  Spinner,
  Stack,
  Text,
} from "@cia-da-vacina/design-system";
import { InboxIcon } from "@cia-da-vacina/icon-system";
import { useAuth } from "@/contexts/auth-context";
import { api } from "@/lib/api";

export default function InboxPage() {
  const router = useRouter();
  const { activeUnitId } = useAuth();
  const { data, isLoading, error } = useQuery({
    queryKey: ["inbox", activeUnitId],
    queryFn: () => api.listInbox({ unit_id: activeUnitId ?? undefined }),
    enabled: Boolean(activeUnitId),
    refetchInterval: 5_000,
    refetchOnWindowFocus: true,
  });

  return (
    <Stack gap={3} width="100%">
      <Stack gap={1}>
        <Heading as="h1" display>
          Inbox
        </Heading>
        <Text muted fontSize="sm">
          WhatsApp · triagem IA · handoff · pipeline
        </Text>
      </Stack>

      {isLoading && (
        <Flex alignItems="center" justifyContent="center" gap={2} py={6}>
          <Spinner />
          <Text muted>Carregando…</Text>
        </Flex>
      )}

      {error && (
        <Text color="text.danger">Não foi possível carregar o inbox.</Text>
      )}

      {!isLoading && (data?.items?.length ?? 0) === 0 && (
        <EmptyState
          icon={<InboxIcon size="xl" fill="text.muted" />}
          title="Nenhuma conversa"
          description="Quando clientes falarem no WhatsApp desta unidade, elas aparecem aqui."
        />
      )}

      {(data?.items?.length ?? 0) > 0 && (
        <ConversationList>
          {data!.items.map((item) => (
            <ConversationRow
              key={item.id}
              contactName={item.contact_name}
              preview={item.last_message_preview}
              stage={item.pipeline_stage}
              mode={item.mode}
              aiSummary={item.ai_summary}
              timestamp={new Date(item.last_message_at).toLocaleString("pt-BR")}
              onClick={() => router.push(`/inbox/${item.id}`)}
            />
          ))}
        </ConversationList>
      )}
    </Stack>
  );
}
