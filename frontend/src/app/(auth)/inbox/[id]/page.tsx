"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useMemo, useState } from "react";
import styled from "styled-components";
import {
  Badge,
  Box,
  Button,
  Flex,
  Heading,
  PageHeader,
  Spinner,
  Stack,
  StageBadge,
  Text,
  TextField,
  useToast,
} from "@cia-da-vacina/design-system";
import { BotIcon, HandshakeIcon, SendIcon } from "@cia-da-vacina/icon-system";
import { ModalShell } from "@/components/modal-shell";
import { PipelineModal } from "@/components/pipeline-modal";
import { fadeUp, staggerContainer, staggerItem } from "@/lib/motion";
import { ApiError } from "@/lib/errors";
import { conversationsService, popsService } from "@/services";
import { CHANNEL_LABELS, CUSTOMER_IDENTIFICATION_LABELS } from "@/domain";

const Workspace = styled.div`
  display: flex;
  flex-direction: column;
  width: 100%;
  height: calc(100dvh - 88px);
  min-height: 480px;
  margin: -8px 0 -40px;
  gap: 12px;
`;

const Split = styled.div`
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(240px, 300px);
  gap: 12px;
  flex: 1;
  min-height: 0;

  @media (max-width: 900px) {
    grid-template-columns: 1fr;
    grid-template-rows: minmax(360px, 1fr) auto;
    overflow: auto;
  }
`;

const ThreadPanel = styled.div`
  display: flex;
  flex-direction: column;
  min-height: 0;
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  border-radius: ${({ theme }) => theme.radii.md};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  overflow: hidden;
`;

const ThreadScroll = styled.div`
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
`;

const SidePanel = styled.aside`
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  overflow-y: auto;
`;

const SideCard = styled.div`
  padding: 12px;
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  border-radius: ${({ theme }) => theme.radii.md};
  background: ${({ theme }) => theme.colors["bg.surface"]};
`;

const AiTriageCard = styled.div`
  padding: 12px;
  border-radius: ${({ theme }) => theme.radii.md};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  box-shadow: inset 3px 0 0 ${({ theme }) => theme.colors["mode.ai.text"]};
`;

const Bubble = styled.div<{ $variant: "ai" | "mine" | "theirs" }>`
  max-width: 85%;
  padding: 8px;
  border-radius: ${({ theme }) => theme.radii.sm};
  background: ${({ theme, $variant }) => {
    if ($variant === "ai") return theme.colors["mode.ai.bg"];
    if ($variant === "mine") return theme.colors["bg.brand.subtle"];
    return theme.colors["bg.surface.muted"];
  }};
  border: 1px solid
    ${({ theme, $variant }) =>
      $variant === "ai" ? theme.colors["border.subtle"] : "transparent"};
`;

const BubbleMeta = styled(Text)<{ $ai?: boolean }>`
  font-size: ${({ theme }) => theme.fontSizes.xs};
  color: ${({ theme, $ai }) =>
    $ai ? theme.colors["mode.ai.text"] : theme.colors["text.muted"]};
  opacity: ${({ $ai }) => ($ai ? 0.85 : 1)};
`;

export default function ConversationPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const qc = useQueryClient();
  const toast = useToast();
  const [draft, setDraft] = useState("");
  const [pipelineOpen, setPipelineOpen] = useState(false);
  const [claimOpen, setClaimOpen] = useState(false);
  const [claimSending, setClaimSending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const conversation = useQuery({
    queryKey: ["conversation", id],
    queryFn: () => conversationsService.get(id),
    refetchInterval: 5_000,
    refetchOnWindowFocus: true,
  });

  const messages = useQuery({
    queryKey: ["messages", id],
    queryFn: () => conversationsService.listMessages(id),
    refetchInterval: 3_000,
    refetchOnWindowFocus: true,
  });

  // Best-effort: the triage endpoint may 404 for conversations already
  // claimed by a human, or fail transiently — never block the page on it.
  const triage = useQuery({
    queryKey: ["triage", id],
    queryFn: () => conversationsService.getTriage(id),
    enabled: Boolean(conversation.data) && conversation.data?.mode === "ai_triage",
    retry: false,
  });

  const pops = useQuery({
    queryKey: ["pops", conversation.data?.intent],
    queryFn: () => popsService.list({ intent: conversation.data?.intent ?? undefined }),
    enabled: Boolean(conversation.data),
  });

  const claim = useMutation({
    mutationFn: () => conversationsService.claim(id),
    onSuccess: async () => {
      toast.push("Atendimento assumido", "success");
      await qc.invalidateQueries({ queryKey: ["conversation", id] });
      await qc.invalidateQueries({ queryKey: ["messages", id] });
      await qc.invalidateQueries({ queryKey: ["inbox"] });
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : "Falha ao assumir"),
  });

  const send = useMutation({
    mutationFn: (body: string) => conversationsService.sendMessage(id, { body }),
    onSuccess: async () => {
      setDraft("");
      toast.push("Mensagem enviada", "success");
      await qc.invalidateQueries({ queryKey: ["messages", id] });
      await qc.invalidateQueries({ queryKey: ["inbox"] });
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : "Falha ao enviar"),
  });

  const busy = send.isPending || claim.isPending || claimSending;

  const requestSend = () => {
    const body = draft.trim();
    if (!body || busy) return;
    if (conversation.data?.mode === "ai_triage") {
      setClaimOpen(true);
      return;
    }
    send.mutate(body);
  };

  const confirmClaimAndSend = async () => {
    const body = draft.trim();
    if (!body || claimSending) return;
    setError(null);
    setClaimSending(true);
    try {
      await claim.mutateAsync();
      await send.mutateAsync(body);
      setClaimOpen(false);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Falha ao assumir e enviar");
    } finally {
      setClaimSending(false);
    }
  };

  const windowHint = useMemo(() => {
    const exp = conversation.data?.window_expires_at;
    if (!exp) return "Janela de atendimento ativa";
    return new Date(exp) > new Date()
      ? "Dentro da janela de 24h"
      : "Fora da janela. Use um template.";
  }, [conversation.data?.window_expires_at]);

  if (conversation.isLoading) {
    return (
      <Flex gap={2} alignItems="center" py={6}>
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
  const truncatedCustomerId = `${c.customer_id.slice(0, 8)}…`;

  return (
    <motion.div {...fadeUp} style={{ width: "100%" }}>
      <Workspace>
        <PageHeader
          eyebrow={<Link href="/inbox">← Inbox</Link>}
          title={c.customer_name}
          description={
            <Flex gap={2} alignItems="center" flexWrap="wrap">
              <Text muted fontSize="xs">
                {CHANNEL_LABELS[c.channel]}
                {c.customer_phone ? `, ${c.customer_phone}` : ""}
              </Text>
              <Text muted fontSize="xs" title={c.customer_id}>
                ID {truncatedCustomerId}
              </Text>
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
              <Badge tone={c.identification === "identified" ? "success" : "neutral"}>
                {CUSTOMER_IDENTIFICATION_LABELS[c.identification]}
              </Badge>
              {c.intent && <Badge tone="brand">{c.intent}</Badge>}
            </Flex>
          }
          actions={
            <Flex gap={2} flexWrap="wrap">
              {isAi && (
                <Button
                  size="sm"
                  leftIcon={<HandshakeIcon size="sm" />}
                  onClick={() => claim.mutate()}
                  disabled={claim.isPending}
                >
                  Assumir atendimento
                </Button>
              )}
              <Button size="sm" variant="secondary" onClick={() => setPipelineOpen(true)}>
                Pipeline
              </Button>
            </Flex>
          }
        />

        <Split>
          <ThreadPanel>
            <Box
              px={3}
              py={2}
              borderBottomWidth="hairline"
              borderBottomStyle="solid"
              borderBottomColor="border.subtle"
            >
              <Text fontSize="xs" muted>
                {windowHint}
              </Text>
            </Box>

            <ThreadScroll>
              {(messages.data?.items ?? []).map((m) => {
                const mine = m.direction === "out";
                const isAiMsg = m.sender_type === "ai";
                return (
                  <Flex key={m.id} justifyContent={mine ? "flex-end" : "flex-start"}>
                    <Bubble $variant={isAiMsg ? "ai" : mine ? "mine" : "theirs"}>
                      <BubbleMeta $ai={isAiMsg}>
                        {isAiMsg
                          ? "IA"
                          : m.sender_type === "agent"
                            ? "Atendente"
                            : m.sender_type === "system"
                              ? "Sistema"
                              : "Cliente"}{" "}
                        {new Date(m.created_at).toLocaleTimeString("pt-BR")}
                      </BubbleMeta>
                      <Text fontSize="sm">{m.body}</Text>
                    </Bubble>
                  </Flex>
                );
              })}
            </ThreadScroll>

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
                      placeholder="Mensagem…"
                      value={draft}
                      disabled={busy}
                      onChange={(e) => setDraft(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" && !e.shiftKey && draft.trim()) {
                          e.preventDefault();
                          requestSend();
                        }
                      }}
                    />
                  </Box>
                  <Button
                    size="sm"
                    leftIcon={<SendIcon size="sm" />}
                    disabled={!draft.trim() || busy}
                    onClick={requestSend}
                  >
                    Enviar
                  </Button>
                </Flex>
                {isAi && (
                  <Text fontSize="xs" muted>
                    A IA ainda conduz esta conversa. Ao enviar, você assume o atendimento.
                  </Text>
                )}
              </Stack>
            </Box>
          </ThreadPanel>

          <SidePanel>
            {isAi && (
              <AiTriageCard>
                <Flex gap={2} alignItems="flex-start">
                  <BotIcon size="md" fill="mode.ai.text" />
                  <Stack gap={1}>
                    <Text fontWeight="semibold" fontSize="sm" color="mode.ai.text">
                      Triagem por IA
                    </Text>
                    <Text fontSize="sm">
                      {triage.data?.summary ?? c.ai_summary ?? "Coletando necessidade…"}
                    </Text>
                    {(triage.data?.intent ?? c.intent) && (
                      <Text fontSize="xs" muted>
                        Intenção: {triage.data?.intent ?? c.intent}
                      </Text>
                    )}
                    {(triage.data?.phone_gate ?? c.phone_gate) === "required" && (
                      <Text fontSize="xs" muted>
                        Aguardando telefone do cliente antes de liberar dados cadastrais.
                      </Text>
                    )}
                    {(triage.data?.phone_gate ?? c.phone_gate) ===
                      "pending_verification" && (
                      <Text fontSize="xs" muted>
                        Código enviado para{" "}
                        {triage.data?.pending_phone_masked ??
                          c.pending_phone_masked ??
                          "o WhatsApp informado"}
                        . Aguardando confirmação.
                      </Text>
                    )}
                    <Text fontSize="xs" fontWeight="medium" color="mode.ai.text">
                      Você pode escrever a resposta abaixo. No envio, confirmamos se deseja
                      assumir o atendimento.
                    </Text>
                  </Stack>
                </Flex>
              </AiTriageCard>
            )}

            <SideCard>
              <Stack gap={2}>
                <Text fontWeight="semibold" fontSize="sm">
                  POPs sugeridos
                </Text>
                {(pops.data ?? []).length === 0 && (
                  <Text fontSize="xs" muted>
                    Nenhum script para esta intenção.
                  </Text>
                )}
                <motion.div variants={staggerContainer} initial="initial" animate="animate">
                  <Stack gap={2}>
                    {(pops.data ?? []).map((pop) => (
                      <motion.div key={pop.id} variants={staggerItem}>
                        <Box
                          p={2}
                          borderWidth="hairline"
                          borderStyle="solid"
                          borderColor="border.subtle"
                          borderRadius="sm"
                          bg="bg.surface.muted"
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
                            disabled={busy}
                          >
                            Inserir
                          </Button>
                        </Box>
                      </motion.div>
                    ))}
                  </Stack>
                </motion.div>
              </Stack>
            </SideCard>

            <SideCard>
              <Stack gap={2}>
                <Text fontWeight="semibold" fontSize="sm">
                  Pipeline
                </Text>
                <Text fontSize="xs" muted>
                  Etapa atual: use o botão Pipeline no cabeçalho para avançar, perder ou
                  agendar follow-up.
                </Text>
                <StageBadge stage={c.pipeline_stage} />
                <Button size="sm" variant="secondary" onClick={() => setPipelineOpen(true)}>
                  Atualizar etapa
                </Button>
              </Stack>
            </SideCard>
          </SidePanel>
        </Split>

        <PipelineModal
          open={pipelineOpen}
          current={c.pipeline_stage}
          onClose={() => setPipelineOpen(false)}
          onConfirm={async (payload) => {
            await conversationsService.updatePipeline(id, payload);
            toast.push("Pipeline atualizado", "success");
            await qc.invalidateQueries({ queryKey: ["conversation", id] });
            await qc.invalidateQueries({ queryKey: ["inbox"] });
            await qc.invalidateQueries({ queryKey: ["followups"] });
            await qc.invalidateQueries({ queryKey: ["dashboard"] });
          }}
        />

        <ModalShell
          open={claimOpen}
          onClose={() => {
            if (!claimSending) setClaimOpen(false);
          }}
          closeOnBackdrop={!claimSending}
        >
          <Stack gap={3}>
            <Flex gap={2} alignItems="center">
              <HandshakeIcon size="md" />
              <Heading as="h3">Deseja assumir o atendimento?</Heading>
            </Flex>
            <Text fontSize="sm" muted>
              Esta conversa ainda está com a IA. Ao confirmar, você assume o atendimento,
              a triagem automática para e a mensagem é enviada ao cliente.
            </Text>
            {error && <Text color="text.danger">{error}</Text>}
            <Flex gap={2} justifyContent="flex-end">
              <Button
                variant="secondary"
                disabled={claimSending}
                onClick={() => setClaimOpen(false)}
              >
                Cancelar
              </Button>
              <Button
                leftIcon={<SendIcon size="sm" />}
                disabled={claimSending}
                onClick={() => void confirmClaimAndSend()}
              >
                {claimSending ? "Assumindo…" : "Assumir e enviar"}
              </Button>
            </Flex>
          </Stack>
        </ModalShell>
      </Workspace>
    </motion.div>
  );
}
