"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
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
  Text,
  TextField,
  useToast,
} from "@cia-da-vacina/design-system";
import { fadeUp, staggerContainer, staggerItem } from "@/lib/motion";
import { useAuth } from "@/providers/auth-provider";
import { metaService } from "@/services";
import {
  CHANNELS,
  CHANNEL_LABELS,
  type ChannelType,
  type MetaChannelConfig,
} from "@/domain";

const DEFAULT_AI_SYSTEM_PROMPT = `Você é a assistente virtual da Cia da Vacina. Atenda com empatia e objetividade,
tire dúvidas sobre vacinas e agendamentos, e nunca dê orientações médicas
definitivas. Sempre direcione casos clínicos para um profissional da unidade.
Assim que perceber intenção de agendamento, reclamação ou algo fora do escopo,
repasse a conversa para um atendente humano.`;

const DEFAULT_AI_CONTEXT = `Unidades abertas de segunda a sábado. Priorize campanhas vigentes ao sugerir
vacinas. Nunca prometa disponibilidade de estoque sem confirmar com a unidade.`;

const TextArea = styled.textarea`
  width: 100%;
  min-height: 120px;
  resize: vertical;
  padding: 10px 12px;
  font-family: inherit;
  font-size: ${({ theme }) => theme.fontSizes.sm};
  line-height: 1.5;
  color: ${({ theme }) => theme.colors["input.text"]};
  background: ${({ theme }) => theme.colors["input.bg"]};
  border: 1px solid ${({ theme }) => theme.colors["input.border"]};
  border-radius: ${({ theme }) => theme.radii.sm};
  outline: none;

  &:hover {
    border-color: ${({ theme }) => theme.colors["input.border.hover"]};
  }

  &:focus {
    border-color: ${({ theme }) => theme.colors["input.border.focus"]};
    box-shadow: ${({ theme }) => theme.shadows.focus};
  }

  &::placeholder {
    color: ${({ theme }) => theme.colors["input.placeholder"]};
  }
`;

const Section = styled(Box)`
  padding: ${({ theme }) => theme.space[3]};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  border-radius: ${({ theme }) => theme.radii.md};
  width: 100%;
`;

type ChannelFormState = {
  enabled: boolean;
  account_id: string;
  display_name: string;
  newToken: string;
};

function blankChannelForm(): ChannelFormState {
  return { enabled: false, account_id: "", display_name: "", newToken: "" };
}

function toFormState(config?: MetaChannelConfig): ChannelFormState {
  if (!config) return blankChannelForm();
  return {
    enabled: config.enabled,
    account_id: config.account_id,
    display_name: config.display_name,
    newToken: "",
  };
}

export default function MetaSettingsPage() {
  const router = useRouter();
  const qc = useQueryClient();
  const toast = useToast();
  const { user, loading: authLoading } = useAuth();
  const isAdmin = user?.role === "admin";

  const { data, isLoading } = useQuery({
    queryKey: ["meta-settings"],
    queryFn: () => metaService.getSettings(),
    enabled: isAdmin,
  });

  const [aiEnabled, setAiEnabled] = useState(true);
  const [triageEnabled, setTriageEnabled] = useState(true);
  const [aiPrompt, setAiPrompt] = useState(DEFAULT_AI_SYSTEM_PROMPT);
  const [aiContext, setAiContext] = useState(DEFAULT_AI_CONTEXT);
  const [channelForms, setChannelForms] = useState<Record<ChannelType, ChannelFormState>>({
    whatsapp: blankChannelForm(),
    instagram: blankChannelForm(),
    facebook: blankChannelForm(),
  });

  useEffect(() => {
    if (!authLoading && user && !isAdmin) {
      router.replace("/inbox");
    }
  }, [authLoading, user, isAdmin, router]);

  useEffect(() => {
    if (!data) return;
    setAiEnabled(data.ai_enabled);
    setTriageEnabled(data.triage_enabled);
    setAiPrompt(data.ai_system_prompt?.trim() || DEFAULT_AI_SYSTEM_PROMPT);
    setAiContext(data.ai_context?.trim() || DEFAULT_AI_CONTEXT);
    const byChannel = new Map(data.channels.map((c) => [c.channel, c] as const));
    setChannelForms({
      whatsapp: toFormState(byChannel.get("whatsapp")),
      instagram: toFormState(byChannel.get("instagram")),
      facebook: toFormState(byChannel.get("facebook")),
    });
  }, [data]);

  const tokenMaskByChannel = useMemo(() => {
    const map = new Map<ChannelType, string>();
    for (const c of data?.channels ?? []) map.set(c.channel, c.token_masked);
    return map;
  }, [data?.channels]);

  const webhookVerifiedByChannel = useMemo(() => {
    const map = new Map<ChannelType, boolean>();
    for (const c of data?.channels ?? []) map.set(c.channel, c.webhook_verified);
    return map;
  }, [data?.channels]);

  const campaigns = data?.ai_campaigns ?? [];
  const activeCampaigns = campaigns.filter((c) => c.active).length;
  const today = new Date().toISOString().slice(0, 10);
  const vigenteCount = campaigns.filter(
    (c) =>
      c.active &&
      (!c.starts_on || c.starts_on <= today) &&
      (!c.ends_on || c.ends_on >= today),
  ).length;

  function updateChannel(channel: ChannelType, patch: Partial<ChannelFormState>) {
    setChannelForms((prev) => ({ ...prev, [channel]: { ...prev[channel], ...patch } }));
  }

  const save = useMutation({
    mutationFn: () => {
      const channelTokens: Partial<Record<ChannelType, string>> = {};
      for (const channel of CHANNELS) {
        const token = channelForms[channel].newToken.trim();
        if (token) channelTokens[channel] = token;
      }
      return metaService.updateSettings({
        ai_enabled: aiEnabled,
        triage_enabled: triageEnabled,
        ai_system_prompt: aiPrompt,
        ai_context: aiContext,
        channels: CHANNELS.map((channel) => ({
          channel,
          enabled: channelForms[channel].enabled,
          account_id: channelForms[channel].account_id,
          display_name: channelForms[channel].display_name,
        })),
        channel_tokens: Object.keys(channelTokens).length ? channelTokens : undefined,
      });
    },
    onSuccess: async () => {
      setChannelForms((prev) => {
        const next = { ...prev };
        for (const channel of CHANNELS) next[channel] = { ...next[channel], newToken: "" };
        return next;
      });
      toast.push("Configurações salvas", "success");
      await qc.invalidateQueries({ queryKey: ["meta-settings"] });
    },
  });

  if (authLoading || !user) {
    return (
      <Flex gap={2} alignItems="center">
        <Spinner />
        <Text muted>Carregando…</Text>
      </Flex>
    );
  }

  if (!isAdmin) return null;

  return (
    <motion.div {...fadeUp} style={{ width: "100%" }}>
      <Stack gap={3} width="100%">
        <PageHeader
          title="Canais Meta & IA"
          description="Conexão WhatsApp, Instagram e Facebook, e comportamento da triagem. Campanhas ficam na Agenda."
          actions={
            <Button onClick={() => save.mutate()} disabled={save.isPending || isLoading}>
              {save.isPending ? "Salvando…" : "Salvar configurações"}
            </Button>
          }
        />

        {isLoading || !data ? (
          <Flex gap={2} alignItems="center">
            <Spinner />
            <Text muted>Carregando…</Text>
          </Flex>
        ) : (
          <>
            <Flex gap={2} flexWrap="wrap">
              <Badge tone={aiEnabled ? "ai" : "neutral"}>
                IA {aiEnabled ? "ligada" : "desligada"}
              </Badge>
              <Badge tone={triageEnabled ? "ai" : "neutral"}>
                Triagem {triageEnabled ? "ativa" : "desativada"}
              </Badge>
              <Badge tone="brand">
                {activeCampaigns} ativa(s), {vigenteCount} vigente(s)
              </Badge>
            </Flex>

            <Section>
              <Stack gap={3}>
                <Heading as="h3">Como a IA funciona</Heading>
                <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
                  <input
                    type="checkbox"
                    checked={aiEnabled}
                    onChange={(e) => setAiEnabled(e.target.checked)}
                  />
                  <Text fontSize="sm">IA de atendimento habilitada</Text>
                </label>
                <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
                  <input
                    type="checkbox"
                    checked={triageEnabled}
                    onChange={(e) => setTriageEnabled(e.target.checked)}
                  />
                  <Text fontSize="sm">Triagem automática antes do handoff humano</Text>
                </label>
                <Stack gap={1}>
                  <Text as="label" htmlFor="ai-prompt" fontSize="xs" fontWeight="medium" muted>
                    Regras e comportamento
                  </Text>
                  <TextArea
                    id="ai-prompt"
                    value={aiPrompt}
                    onChange={(e) => setAiPrompt(e.target.value)}
                    rows={12}
                    placeholder="Como a IA deve agir, o que não deve responder, tom de voz…"
                  />
                  <Text fontSize="xs" muted>
                    Inclua tom de voz, o que pode prometer, limites médicos e quando passar
                    para humano.
                  </Text>
                </Stack>
              </Stack>
            </Section>

            <Section>
              <Stack gap={3}>
                <Heading as="h3">Contexto operacional</Heading>
                <Text muted fontSize="sm">
                  O que está acontecendo agora: horários, foco da operação, restrições e infos
                  que a IA deve considerar nas respostas.
                </Text>
                <TextArea
                  id="ai-context"
                  value={aiContext}
                  onChange={(e) => setAiContext(e.target.value)}
                  rows={6}
                  placeholder="Ex.: unidades abertas, horários, prioridade da semana…"
                />
              </Stack>
            </Section>

            <Stack gap={2} width="100%">
              <Heading as="h3">Canais conectados</Heading>
              <motion.div variants={staggerContainer} initial="initial" animate="animate">
                <Stack gap={2}>
                  {CHANNELS.map((channel) => {
                    const form = channelForms[channel];
                    const webhookVerified = webhookVerifiedByChannel.get(channel) ?? false;
                    return (
                      <motion.div key={channel} variants={staggerItem}>
                        <Section>
                          <Stack gap={3}>
                            <Flex
                              justifyContent="space-between"
                              alignItems="center"
                              flexWrap="wrap"
                              gap={2}
                            >
                              <Flex gap={2} alignItems="center">
                                <Text fontWeight="semibold">{CHANNEL_LABELS[channel]}</Text>
                                <Badge tone={form.enabled ? "success" : "neutral"}>
                                  {form.enabled ? "Ativo" : "Inativo"}
                                </Badge>
                                {webhookVerified ? (
                                  <Badge tone="success">Webhook verificado</Badge>
                                ) : (
                                  <Badge tone="warning">Webhook pendente</Badge>
                                )}
                              </Flex>
                              <label
                                style={{ display: "flex", gap: 8, alignItems: "center" }}
                              >
                                <input
                                  type="checkbox"
                                  checked={form.enabled}
                                  onChange={(e) =>
                                    updateChannel(channel, { enabled: e.target.checked })
                                  }
                                />
                                <Text fontSize="sm">Habilitado</Text>
                              </label>
                            </Flex>
                            <TextField
                              label="Nome de exibição"
                              value={form.display_name}
                              onChange={(e) =>
                                updateChannel(channel, { display_name: e.target.value })
                              }
                            />
                            <TextField
                              label={
                                channel === "whatsapp"
                                  ? "WABA / conta"
                                  : channel === "instagram"
                                    ? "Conta comercial do Instagram"
                                    : "Página do Facebook"
                              }
                              value={form.account_id}
                              onChange={(e) =>
                                updateChannel(channel, { account_id: e.target.value })
                              }
                            />
                            <TextField
                              label={`Novo token (opcional). Atual: ${tokenMaskByChannel.get(channel) ?? "não configurado"}`}
                              type="password"
                              placeholder="Cole um novo token para rotacionar"
                              value={form.newToken}
                              onChange={(e) =>
                                updateChannel(channel, { newToken: e.target.value })
                              }
                            />
                          </Stack>
                        </Section>
                      </motion.div>
                    );
                  })}
                </Stack>
              </motion.div>
            </Stack>

            <Section>
              <Stack gap={2}>
                <Heading as="h3">Campanhas e promoções</Heading>
                <Text muted fontSize="sm">
                  Criação, calendário e vigência ficam na Agenda. A IA usa as campanhas
                  vigentes na triagem.
                </Text>
                <Text fontSize="sm">
                  {campaigns.length === 0
                    ? "Nenhuma campanha cadastrada."
                    : `${vigenteCount} vigente(s) agora, ${activeCampaigns} marcada(s) como ativa(s), ${campaigns.length} no total.`}
                </Text>
                <Flex>
                  <Link href="/campaigns" style={{ textDecoration: "none" }}>
                    <Button type="button" variant="secondary">
                      Abrir Agenda
                    </Button>
                  </Link>
                </Flex>
              </Stack>
            </Section>

            <Flex justifyContent="flex-end" gap={2}>
              <Button onClick={() => save.mutate()} disabled={save.isPending}>
                {save.isPending ? "Salvando…" : "Salvar configurações"}
              </Button>
            </Flex>
          </>
        )}
      </Stack>
    </motion.div>
  );
}
