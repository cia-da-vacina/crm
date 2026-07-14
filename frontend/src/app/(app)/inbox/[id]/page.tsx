"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useMemo, useState } from "react";
import {
  Badge,
  Box,
  Button,
  Flex,
  Heading,
  Spinner,
  Stack,
  StageBadge,
  Text,
  TextField,
} from "@cia-da-vacina/design-system";
import { BotIcon, HandshakeIcon, SendIcon } from "@cia-da-vacina/icon-system";
import { PipelineModal } from "@/components/pipeline-modal";
import { api, ApiError } from "@/lib/api";

export default function ConversationPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const qc = useQueryClient();
  const [draft, setDraft] = useState("");
  const [pipelineOpen, setPipelineOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const conversation = useQuery({
    queryKey: ["conversation", id],
    queryFn: () => api.getConversation(id),
    refetchInterval: 5_000,
    refetchOnWindowFocus: true,
  });

  const messages = useQuery({
    queryKey: ["messages", id],
    queryFn: () => api.listMessages(id),
    refetchInterval: 3_000,
    refetchOnWindowFocus: true,
  });

  const pops = useQuery({
    queryKey: ["pops", conversation.data?.intent],
    queryFn: () => api.listPops(conversation.data?.intent ?? undefined),
    enabled: Boolean(conversation.data),
  });

  const claim = useMutation({
    mutationFn: () => api.claim(id),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["conversation", id] });
      await qc.invalidateQueries({ queryKey: ["messages", id] });
      await qc.invalidateQueries({ queryKey: ["inbox"] });
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : "Falha no claim"),
  });

  const send = useMutation({
    mutationFn: (body: string) => api.sendMessage(id, body),
    onSuccess: async () => {
      setDraft("");
      await qc.invalidateQueries({ queryKey: ["messages", id] });
      await qc.invalidateQueries({ queryKey: ["inbox"] });
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : "Falha ao enviar"),
  });

  const windowHint = useMemo(() => {
    const exp = conversation.data?.window_expires_at;
    if (!exp) return "Janela 24h ativa (demo)";
    return new Date(exp) > new Date()
      ? "Dentro da janela 24h"
      : "Fora da janela — use template";
  }, [conversation.data?.window_expires_at]);

  if (conversation.isLoading) {
    return (
      <Flex gap={2} alignItems="center" justifyContent="center" py={6}>
        <Spinner />
        <Text muted>Carregando conversa…</Text>
      </Flex>
    );
  }

  if (!conversation.data) {
    return <Text color="text.danger">Conversa não encontrada.</Text>;
  }

  const c = conversation.data;
  const isAi = c.mode === "ai_triage";

  return (
    <Stack gap={3} width="100%">
      <Flex justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
        <Stack gap={1}>
          <Text fontSize="xs">
            <Link href="/inbox">← Inbox</Link>
          </Text>
          <Heading as="h1" display>
            {c.contact_name}
          </Heading>
          <Text muted fontSize="xs">
            {c.contact_phone}
          </Text>
          <Flex gap={1} flexWrap="wrap" alignItems="center">
            <StageBadge stage={c.pipeline_stage} />
            <Badge tone={isAi ? "ai" : "human"}>
              {isAi ? (
                <>
                  <BotIcon size="xs" /> IA
                </>
              ) : (
                "Humano"
              )}
            </Badge>
            {c.intent && <Badge tone="brand">{c.intent}</Badge>}
          </Flex>
        </Stack>
        <Flex gap={2} flexWrap="wrap">
          {isAi && (
            <Button
              size="sm"
              leftIcon={<HandshakeIcon size="sm" />}
              onClick={() => claim.mutate()}
              disabled={claim.isPending}
            >
              Assumir
            </Button>
          )}
          <Button size="sm" variant="secondary" onClick={() => setPipelineOpen(true)}>
            Pipeline
          </Button>
        </Flex>
      </Flex>

      {isAi && (
        <Box
          p={3}
          bg="mode.ai.bg"
          borderRadius="md"
          borderWidth="hairline"
          borderStyle="solid"
          borderColor="border.default"
        >
          <Flex gap={2} alignItems="flex-start">
            <BotIcon size="md" fill="mode.ai.text" />
            <Stack gap={1}>
              <Text fontWeight="semibold" fontSize="sm" color="mode.ai.text">
                Triagem por IA
              </Text>
              <Text fontSize="sm">{c.ai_summary ?? "Coletando necessidade…"}</Text>
            </Stack>
          </Flex>
        </Box>
      )}

      <Flex gap={3} alignItems="stretch" flexWrap="wrap" width="100%">
        <Box
          flex={1}
          minWidth="260px"
          borderWidth="hairline"
          borderStyle="solid"
          borderColor="border.default"
          borderRadius="md"
          bg="bg.surface"
          overflow="hidden"
        >
          <Box px={3} py={2} borderBottomWidth="hairline" borderBottomStyle="solid" borderBottomColor="border.subtle">
            <Text fontSize="xs" muted>
              {windowHint}
            </Text>
          </Box>
          <Stack gap={2} p={3} style={{ maxHeight: 320, overflowY: "auto" }}>
            {(messages.data?.items ?? []).map((m) => {
              const mine = m.direction === "out";
              return (
                <Flex key={m.id} justifyContent={mine ? "flex-end" : "flex-start"}>
                  <Box
                    maxWidth="85%"
                    px={2}
                    py={2}
                    borderRadius="sm"
                    bg={
                      m.sender_type === "ai"
                        ? "mode.ai.bg"
                        : mine
                          ? "bg.brand.subtle"
                          : "bg.surface.muted"
                    }
                  >
                    <Text fontSize="xs" muted>
                      {m.sender_type === "ai"
                        ? "IA"
                        : m.sender_type === "agent"
                          ? "Atendente"
                          : m.sender_type === "system"
                            ? "Sistema"
                            : "Cliente"}{" "}
                      · {new Date(m.created_at).toLocaleTimeString("pt-BR")}
                    </Text>
                    <Text fontSize="sm">{m.body}</Text>
                  </Box>
                </Flex>
              );
            })}
          </Stack>
          <Box p={2} borderTopWidth="hairline" borderTopStyle="solid" borderTopColor="border.subtle">
            <Stack gap={2}>
              {error && (
                <Text color="text.danger" fontSize="xs">
                  {error}
                </Text>
              )}
              <Flex gap={2} alignItems="flex-end">
                <Box flex={1}>
                  <TextField
                    placeholder={
                      isAi ? "Assuma a conversa para responder…" : "Mensagem…"
                    }
                    value={draft}
                    disabled={isAi || send.isPending}
                    onChange={(e) => setDraft(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && !e.shiftKey && draft.trim()) {
                        e.preventDefault();
                        send.mutate(draft);
                      }
                    }}
                  />
                </Box>
                <Button
                  size="sm"
                  leftIcon={<SendIcon size="sm" />}
                  disabled={isAi || !draft.trim() || send.isPending}
                  onClick={() => send.mutate(draft)}
                >
                  Enviar
                </Button>
              </Flex>
            </Stack>
          </Box>
        </Box>

        <Stack gap={2} width="100%" maxWidth="220px">
          <Text fontWeight="semibold" fontSize="sm">
            POPs
          </Text>
          {(pops.data?.items ?? []).map((pop) => (
            <Box
              key={pop.id}
              p={2}
              borderWidth="hairline"
              borderStyle="solid"
              borderColor="border.default"
              borderRadius="sm"
              bg="bg.surface"
            >
              <Text fontWeight="semibold" fontSize="xs">
                {pop.title}
              </Text>
              <Text fontSize="xs" muted style={{ marginTop: 4 }}>
                {pop.body}
              </Text>
              <Button
                size="sm"
                variant="ghost"
                style={{ marginTop: 6 }}
                onClick={() => setDraft(pop.body)}
                disabled={isAi}
              >
                Inserir
              </Button>
            </Box>
          ))}
        </Stack>
      </Flex>

      <PipelineModal
        open={pipelineOpen}
        current={c.pipeline_stage}
        onClose={() => setPipelineOpen(false)}
        onConfirm={async (payload) => {
          await api.updatePipeline(id, payload);
          await qc.invalidateQueries({ queryKey: ["conversation", id] });
          await qc.invalidateQueries({ queryKey: ["inbox"] });
          await qc.invalidateQueries({ queryKey: ["followups"] });
          await qc.invalidateQueries({ queryKey: ["dashboard"] });
        }}
      />
    </Stack>
  );
}
