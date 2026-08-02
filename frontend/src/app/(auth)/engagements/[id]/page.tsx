"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import {
  Badge,
  Box,
  Button,
  Flex,
  PageHeader,
  Spinner,
  Stack,
  Text,
  TextField,
  useToast,
} from "@cia-da-vacina/design-system";
import { fadeUp } from "@/lib/motion";
import { useNavigationFeedback } from "@/providers/navigation-provider";
import { engagementsService } from "@/services";
import {
  CHANNEL_LABELS,
  ENGAGEMENT_STATUS_LABELS,
  ENGAGEMENT_TYPE_LABELS,
} from "@/domain";

export default function EngagementDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const { navigate } = useNavigationFeedback();
  const qc = useQueryClient();
  const toast = useToast();
  const [draft, setDraft] = useState("");

  const engagement = useQuery({
    queryKey: ["engagement", id],
    queryFn: () => engagementsService.get(id),
  });

  function invalidateAll() {
    return Promise.all([
      qc.invalidateQueries({ queryKey: ["engagement", id] }),
      qc.invalidateQueries({ queryKey: ["engagements"] }),
      qc.invalidateQueries({ queryKey: ["dashboard"] }),
    ]);
  }

  const reply = useMutation({
    mutationFn: (body: string) => engagementsService.reply(id, body),
    onSuccess: async () => {
      setDraft("");
      toast.push("Resposta enviada", "success");
      await invalidateAll();
    },
  });

  const dismiss = useMutation({
    mutationFn: () => engagementsService.dismiss(id),
    onSuccess: async () => {
      toast.push("Interação descartada", "success");
      await invalidateAll();
    },
  });

  const convert = useMutation({
    mutationFn: () => engagementsService.convertToConversation(id),
    onSuccess: async (conversation) => {
      toast.push("Convertida em conversa", "success");
      await invalidateAll();
      await qc.invalidateQueries({ queryKey: ["inbox"] });
      navigate(`/inbox/${conversation.id}`);
    },
  });

  if (engagement.isLoading) {
    return (
      <Flex gap={2} alignItems="center" py={6}>
        <Spinner />
        <Text muted>Carregando interação…</Text>
      </Flex>
    );
  }

  if (!engagement.data) {
    return <Text color="text.danger">Interação não encontrada.</Text>;
  }

  const e = engagement.data;

  return (
    <motion.div {...fadeUp} style={{ width: "100%" }}>
      <Stack gap={3} width="100%">
        <PageHeader
          eyebrow={<Link href="/engagements">← Interações</Link>}
          title={e.customer_name ?? "Contato não identificado"}
          description={
            <Flex gap={1} flexWrap="wrap" alignItems="center">
              <Badge tone="brand">{CHANNEL_LABELS[e.channel]}</Badge>
              <Badge tone="neutral">{ENGAGEMENT_TYPE_LABELS[e.type]}</Badge>
              <Badge tone="info">{ENGAGEMENT_STATUS_LABELS[e.status]}</Badge>
            </Flex>
          }
        />

        <Box
          p={3}
          bg="bg.surface"
          borderWidth="hairline"
          borderStyle="solid"
          borderColor="border.default"
          borderRadius="md"
          width="100%"
        >
          <Stack gap={2}>
            <Text fontSize="sm">{e.body}</Text>
            {e.media_url && (
              <Text fontSize="xs" muted>
                Mídia: {e.media_caption ?? e.media_url}
              </Text>
            )}
            <Text fontSize="xs" muted>
              Recebida em {new Date(e.created_at).toLocaleString("pt-BR")}
            </Text>
            {e.replied_at && (
              <Text fontSize="xs" muted>
                Respondida em {new Date(e.replied_at).toLocaleString("pt-BR")}
              </Text>
            )}
          </Stack>
        </Box>

        {e.status === "open" && (
          <Box
            p={3}
            bg="bg.surface"
            borderWidth="hairline"
            borderStyle="solid"
            borderColor="border.default"
            borderRadius="md"
            width="100%"
          >
            <Stack gap={2}>
              <Text fontWeight="semibold" fontSize="sm">
                Responder
              </Text>
              <Flex gap={2} alignItems="flex-end" flexWrap="wrap">
                <Box flex={1} minWidth="200px">
                  <TextField
                    placeholder="Escreva a resposta…"
                    value={draft}
                    onChange={(ev) => setDraft(ev.target.value)}
                  />
                </Box>
                <Button
                  size="sm"
                  disabled={!draft.trim() || reply.isPending}
                  onClick={() => reply.mutate(draft.trim())}
                >
                  {reply.isPending ? "Enviando…" : "Enviar resposta"}
                </Button>
              </Flex>
            </Stack>
          </Box>
        )}

        <Flex gap={2} flexWrap="wrap">
          {e.status === "open" && (
            <Button
              size="sm"
              variant="secondary"
              disabled={dismiss.isPending}
              onClick={() => dismiss.mutate()}
            >
              Descartar
            </Button>
          )}
          {e.conversation_id ? (
            <Link href={`/inbox/${e.conversation_id}`}>
              <Button size="sm" variant="secondary">
                Abrir conversa
              </Button>
            </Link>
          ) : (
            <Button
              size="sm"
              variant="secondary"
              disabled={convert.isPending}
              onClick={() => convert.mutate()}
            >
              Converter em conversa
            </Button>
          )}
        </Flex>
      </Stack>
    </motion.div>
  );
}
