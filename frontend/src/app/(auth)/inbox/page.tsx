"use client";

import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import {
  ConversationList,
  ConversationRow,
  EmptyState,
  Flex,
  PageHeader,
  Spinner,
  Stack,
  Text,
} from "@cia-da-vacina/design-system";
import { InboxIcon } from "@cia-da-vacina/icon-system";
import { fadeUp, staggerContainer, staggerItem } from "@/lib/motion";
import { useAuth } from "@/providers/auth-provider";
import { useNavigationFeedback } from "@/providers/navigation-provider";
import { inboxService } from "@/services";
import { CHANNEL_LABELS } from "@/domain";

export default function InboxPage() {
  const { navigate } = useNavigationFeedback();
  const { activeUnitId } = useAuth();
  const { data, isLoading, error } = useQuery({
    queryKey: ["inbox", activeUnitId],
    queryFn: () => inboxService.listInbox({ unit_id: activeUnitId ?? undefined }),
    enabled: Boolean(activeUnitId),
    refetchInterval: 5_000,
    refetchOnWindowFocus: true,
  });

  return (
    <motion.div {...fadeUp} style={{ width: "100%" }}>
      <Stack gap={3} width="100%">
        <PageHeader
          title="Inbox"
          description="WhatsApp, Instagram e Facebook, com triagem IA, handoff e pipeline"
        />

        {isLoading && (
          <Flex alignItems="center" gap={2} py={4}>
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
            description="Quando clientes falarem por WhatsApp, Instagram ou Facebook desta unidade, elas aparecem aqui."
          />
        )}

        {(data?.items?.length ?? 0) > 0 && (
          <motion.div variants={staggerContainer} initial="initial" animate="animate">
            <ConversationList>
              {data!.items.map((item) => (
                <motion.div key={item.id} variants={staggerItem}>
                  <ConversationRow
                    contactName={item.customer_name}
                    preview={`${CHANNEL_LABELS[item.channel]}: ${item.last_message_preview}`}
                    stage={item.pipeline_stage}
                    mode={item.mode}
                    aiSummary={item.ai_summary}
                    timestamp={new Date(item.last_message_at).toLocaleString("pt-BR")}
                    onClick={() => navigate(`/inbox/${item.id}`)}
                  />
                </motion.div>
              ))}
            </ConversationList>
          </motion.div>
        )}
      </Stack>
    </motion.div>
  );
}
