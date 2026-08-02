"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import Link from "next/link";
import { useState } from "react";
import {
  Badge,
  Box,
  Button,
  DataList,
  DataListRow,
  EmptyState,
  Flex,
  PageHeader,
  SelectField,
  Spinner,
  Stack,
  Text,
  TextField,
  Toolbar,
  useToast,
} from "@cia-da-vacina/design-system";
import { AtSign } from "@cia-da-vacina/icon-system";
import { fadeUp, staggerContainer, staggerItem } from "@/lib/motion";
import { useAuth } from "@/providers/auth-provider";
import { useNavigationFeedback } from "@/providers/navigation-provider";
import { engagementsService } from "@/services";
import {
  CHANNEL_LABELS,
  ENGAGEMENT_STATUS_LABELS,
  ENGAGEMENT_TYPE_LABELS,
  type EngagementStatus,
  type SocialEngagement,
} from "@/domain";

const STATUS_FILTERS: Array<{ value: EngagementStatus | ""; label: string }> = [
  { value: "open", label: "Em aberto" },
  { value: "replied", label: "Respondidas" },
  { value: "dismissed", label: "Descartadas" },
  { value: "converted_to_conversation", label: "Convertidas" },
  { value: "", label: "Todas" },
];

function statusTone(status: EngagementStatus) {
  switch (status) {
    case "open":
      return "warning" as const;
    case "replied":
      return "success" as const;
    case "converted_to_conversation":
      return "brand" as const;
    default:
      return "neutral" as const;
  }
}

export default function EngagementsPage() {
  const { navigate } = useNavigationFeedback();
  const qc = useQueryClient();
  const toast = useToast();
  const { activeUnitId } = useAuth();
  const [status, setStatus] = useState<EngagementStatus | "">("open");
  const [replyingId, setReplyingId] = useState<string | null>(null);
  const [replyDraft, setReplyDraft] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["engagements", activeUnitId, status],
    queryFn: () =>
      engagementsService.list({
        unit_id: activeUnitId ?? undefined,
        status: status || undefined,
      }),
    enabled: Boolean(activeUnitId),
    refetchInterval: 5_000,
  });

  function invalidateAll() {
    return Promise.all([
      qc.invalidateQueries({ queryKey: ["engagements"] }),
      qc.invalidateQueries({ queryKey: ["dashboard"] }),
    ]);
  }

  const reply = useMutation({
    mutationFn: ({ id, body }: { id: string; body: string }) =>
      engagementsService.reply(id, body),
    onSuccess: async () => {
      setReplyingId(null);
      setReplyDraft("");
      toast.push("Resposta enviada", "success");
      await invalidateAll();
    },
  });

  const dismiss = useMutation({
    mutationFn: (id: string) => engagementsService.dismiss(id),
    onSuccess: async () => {
      toast.push("Interação descartada", "success");
      await invalidateAll();
    },
  });

  const convert = useMutation({
    mutationFn: (id: string) => engagementsService.convertToConversation(id),
    onSuccess: async (conversation) => {
      toast.push("Convertida em conversa", "success");
      await invalidateAll();
      await qc.invalidateQueries({ queryKey: ["inbox"] });
      navigate(`/inbox/${conversation.id}`);
    },
  });

  function startReply(e: SocialEngagement) {
    setReplyingId(e.id);
    setReplyDraft("");
  }

  return (
    <motion.div {...fadeUp} style={{ width: "100%" }}>
      <Stack gap={3} width="100%">
        <PageHeader
          title="Interações"
          description="Respostas a stories e comentários no Instagram e Facebook"
        />

        <Toolbar
          trailing={
            <SelectField
              aria-label="Status"
              fieldSize="sm"
              fullWidth={false}
              value={status}
              onChange={(e) => setStatus(e.target.value as EngagementStatus | "")}
              options={STATUS_FILTERS.map((f) => ({ value: f.value, label: f.label }))}
            />
          }
        >
          <Text fontSize="sm" muted>
            Filtro por status
          </Text>
        </Toolbar>

        {isLoading && (
          <Flex gap={2} alignItems="center">
            <Spinner />
            <Text muted>Carregando interações…</Text>
          </Flex>
        )}

        {!isLoading && (data?.items.length ?? 0) === 0 && (
          <EmptyState
            icon={<AtSign size="xl" fill="text.muted" />}
            title="Nenhuma interação"
            description="Respostas a stories e comentários de Instagram/Facebook aparecem aqui."
          />
        )}

        {(data?.items.length ?? 0) > 0 && (
          <motion.div variants={staggerContainer} initial="initial" animate="animate">
            <DataList>
              {(data?.items ?? []).map((e) => (
                <motion.div key={e.id} variants={staggerItem}>
                  <DataListRow
                    interactive={false}
                    leading={
                      <Stack gap={2} width="100%">
                        <Flex gap={2} alignItems="center" flexWrap="wrap">
                          <Text fontWeight="semibold">
                            {e.customer_name ?? "Contato não identificado"}
                          </Text>
                          <Badge tone="brand">{CHANNEL_LABELS[e.channel]}</Badge>
                          <Badge tone="neutral">{ENGAGEMENT_TYPE_LABELS[e.type]}</Badge>
                          <Badge tone={statusTone(e.status)}>
                            {ENGAGEMENT_STATUS_LABELS[e.status]}
                          </Badge>
                        </Flex>
                        <Text fontSize="sm">{e.body}</Text>
                        {e.media_caption && (
                          <Text fontSize="xs" muted>
                            Mídia: {e.media_caption}
                          </Text>
                        )}
                        <Text fontSize="xs" muted>
                          {new Date(e.created_at).toLocaleString("pt-BR")}
                        </Text>

                        {replyingId === e.id ? (
                          <Flex gap={2} alignItems="flex-end" flexWrap="wrap">
                            <Box flex={1} minWidth="200px">
                              <TextField
                                placeholder="Escreva a resposta…"
                                value={replyDraft}
                                onChange={(ev) => setReplyDraft(ev.target.value)}
                                autoFocus
                              />
                            </Box>
                            <Button
                              size="sm"
                              disabled={!replyDraft.trim() || reply.isPending}
                              onClick={() =>
                                reply.mutate({ id: e.id, body: replyDraft.trim() })
                              }
                            >
                              {reply.isPending ? "Enviando…" : "Enviar resposta"}
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => {
                                setReplyingId(null);
                                setReplyDraft("");
                              }}
                            >
                              Cancelar
                            </Button>
                          </Flex>
                        ) : (
                          <Flex gap={2} flexWrap="wrap">
                            {e.status === "open" && (
                              <>
                                <Button size="sm" onClick={() => startReply(e)}>
                                  Responder
                                </Button>
                                <Button
                                  size="sm"
                                  variant="secondary"
                                  disabled={dismiss.isPending}
                                  onClick={() => dismiss.mutate(e.id)}
                                >
                                  Descartar
                                </Button>
                              </>
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
                                onClick={() => convert.mutate(e.id)}
                              >
                                Converter em conversa
                              </Button>
                            )}
                            <Link href={`/engagements/${e.id}`}>
                              <Button size="sm" variant="ghost">
                                Ver detalhes
                              </Button>
                            </Link>
                          </Flex>
                        )}
                      </Stack>
                    }
                  />
                </motion.div>
              ))}
            </DataList>
          </motion.div>
        )}
      </Stack>
    </motion.div>
  );
}
