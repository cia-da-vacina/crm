"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import styled from "styled-components";
import {
  Badge,
  Box,
  Button,
  Flex,
  Heading,
  Spinner,
  Stack,
  Text,
  TextField,
} from "@cia-da-vacina/design-system";
import { SettingsIcon } from "@cia-da-vacina/icon-system";
import { api } from "@/lib/api";
import {
  DEFAULT_AI_CONTEXT,
  DEFAULT_AI_SYSTEM_PROMPT,
} from "@/lib/ai-defaults";
import type { AICampaign } from "@/lib/types";

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

function emptyCampaign(): AICampaign {
  const today = new Date();
  const end = new Date(today);
  end.setDate(end.getDate() + 30);
  const fmt = (d: Date) => d.toISOString().slice(0, 10);
  return {
    id: "",
    title: "",
    description: "",
    starts_on: fmt(today),
    ends_on: fmt(end),
    active: true,
  };
}

function campaignStatus(c: AICampaign) {
  if (!c.active) return { label: "Inativa", tone: "neutral" as const };
  const today = new Date().toISOString().slice(0, 10);
  if (c.starts_on && today < c.starts_on)
    return { label: "Agendada", tone: "info" as const };
  if (c.ends_on && today > c.ends_on)
    return { label: "Encerrada", tone: "warning" as const };
  return { label: "Vigente", tone: "success" as const };
}

export default function WhatsAppSettingsPage() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ["wa-settings"],
    queryFn: () => api.getWhatsAppSettings(),
  });

  const [displayPhone, setDisplayPhone] = useState("");
  const [wabaId, setWabaId] = useState("");
  const [phoneNumberId, setPhoneNumberId] = useState("");
  const [token, setToken] = useState("");
  const [aiEnabled, setAiEnabled] = useState(true);
  const [aiPrompt, setAiPrompt] = useState(DEFAULT_AI_SYSTEM_PROMPT);
  const [aiContext, setAiContext] = useState(DEFAULT_AI_CONTEXT);
  const [campaigns, setCampaigns] = useState<AICampaign[]>([]);
  const [draft, setDraft] = useState<AICampaign>(emptyCampaign());
  const [editingCampaignId, setEditingCampaignId] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (!data) return;
    setDisplayPhone(data.display_phone);
    setWabaId(data.waba_id);
    setPhoneNumberId(data.phone_number_id);
    setAiEnabled(data.ai_enabled);
    setAiPrompt(data.ai_system_prompt?.trim() || DEFAULT_AI_SYSTEM_PROMPT);
    setAiContext(data.ai_context?.trim() || DEFAULT_AI_CONTEXT);
    setCampaigns(data.ai_campaigns ?? []);
  }, [data]);

  const save = useMutation({
    mutationFn: () =>
      api.updateWhatsAppSettings({
        display_phone: displayPhone,
        waba_id: wabaId,
        phone_number_id: phoneNumberId,
        ai_enabled: aiEnabled,
        ai_system_prompt: aiPrompt,
        ai_context: aiContext,
        ai_campaigns: campaigns,
        token: token || undefined,
      }),
    onSuccess: async () => {
      setToken("");
      setSaved(true);
      await qc.invalidateQueries({ queryKey: ["wa-settings"] });
      setTimeout(() => setSaved(false), 2000);
    },
  });

  function upsertCampaign() {
    const title = draft.title.trim();
    if (!title) return;
    const next: AICampaign = {
      ...draft,
      title,
      description: draft.description.trim(),
      id: draft.id || `camp-${Date.now()}`,
    };
    setCampaigns((prev) => {
      if (editingCampaignId) {
        return prev.map((c) => (c.id === editingCampaignId ? next : c));
      }
      return [next, ...prev];
    });
    setDraft(emptyCampaign());
    setEditingCampaignId(null);
  }

  function editCampaign(c: AICampaign) {
    setDraft({ ...c });
    setEditingCampaignId(c.id);
  }

  function removeCampaign(id: string) {
    setCampaigns((prev) => prev.filter((c) => c.id !== id));
    if (editingCampaignId === id) {
      setDraft(emptyCampaign());
      setEditingCampaignId(null);
    }
  }

  return (
    <Stack gap={3} width="100%">
      <Stack gap={1}>
        <Flex gap={2} alignItems="center">
          <SettingsIcon size="lg" fill="text.brand" />
          <Heading as="h1" display>
            WhatsApp & IA
          </Heading>
        </Flex>
        <Text muted fontSize="sm">
          Conexão Meta, comportamento da triagem e campanhas que a IA pode repassar
        </Text>
      </Stack>

      {isLoading || !data ? (
        <Flex gap={2} alignItems="center">
          <Spinner />
          <Text muted>Carregando…</Text>
        </Flex>
      ) : (
        <>
          <Flex gap={2} flexWrap="wrap">
            {data.webhook_verified ? (
              <Badge tone="success">Webhook verificado</Badge>
            ) : (
              <Badge tone="warning">Webhook pendente</Badge>
            )}
            <Badge tone={aiEnabled ? "ai" : "neutral"}>
              IA {aiEnabled ? "ligada" : "desligada"}
            </Badge>
            <Badge tone="brand">
              {campaigns.filter((c) => c.active).length} campanha(s) ativa(s)
            </Badge>
          </Flex>

          <Section>
            <Stack gap={3}>
              <Heading as="h3">Conexão Meta</Heading>
              <TextField
                label="Telefone exibido"
                value={displayPhone}
                onChange={(e) => setDisplayPhone(e.target.value)}
              />
              <TextField
                label="WABA ID"
                value={wabaId}
                onChange={(e) => setWabaId(e.target.value)}
              />
              <TextField
                label="Phone Number ID"
                value={phoneNumberId}
                onChange={(e) => setPhoneNumberId(e.target.value)}
              />
              <TextField
                label={`Token Meta (atual: ${data.token_masked})`}
                type="password"
                placeholder="Cole um novo token para rotacionar"
                value={token}
                onChange={(e) => setToken(e.target.value)}
              />
            </Stack>
          </Section>

          <Section>
            <Stack gap={3}>
              <Heading as="h3">Como a IA funciona</Heading>
              <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
                <input
                  type="checkbox"
                  checked={aiEnabled}
                  onChange={(e) => setAiEnabled(e.target.checked)}
                />
                <Text fontSize="sm">IA de triagem habilitada</Text>
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
                  Inclua tom de voz, o que pode prometer, limites médicos e quando passar para
                  humano.
                </Text>
              </Stack>
            </Stack>
          </Section>

          <Section>
            <Stack gap={3}>
              <Heading as="h3">Contexto operacional</Heading>
              <Text muted fontSize="sm">
                O que está acontecendo agora: horários, foco da operação, restrições e infos que a
                IA deve considerar nas respostas.
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

          <Section>
            <Stack gap={3}>
              <Heading as="h3">Campanhas e promoções</Heading>
              <Text muted fontSize="sm">
                Textos com período de vigência para a IA mencionar ao cliente quando fizer
                sentido.
              </Text>

              <Box
                p={3}
                bg="bg.surface.muted"
                borderRadius="md"
                borderWidth="hairline"
                borderStyle="solid"
                borderColor="border.subtle"
              >
                <Stack gap={3}>
                  <Text fontWeight="semibold" fontSize="sm">
                    {editingCampaignId ? "Editar campanha" : "Nova campanha"}
                  </Text>
                  <TextField
                    label="Título"
                    value={draft.title}
                    onChange={(e) => setDraft((d) => ({ ...d, title: e.target.value }))}
                    placeholder="Ex.: Campanha Gripe"
                  />
                  <Stack gap={1}>
                    <Text
                      as="label"
                      htmlFor="camp-desc"
                      fontSize="xs"
                      fontWeight="medium"
                      muted
                    >
                      O que a IA deve repassar
                    </Text>
                    <TextArea
                      id="camp-desc"
                      value={draft.description}
                      onChange={(e) =>
                        setDraft((d) => ({ ...d, description: e.target.value }))
                      }
                      rows={3}
                      placeholder="Desconto, condições, validade, mensagens-chave…"
                    />
                  </Stack>
                  <Flex gap={3} flexWrap="wrap">
                    <Box style={{ flex: "1 1 140px" }}>
                      <TextField
                        label="Início"
                        type="date"
                        value={draft.starts_on}
                        onChange={(e) =>
                          setDraft((d) => ({ ...d, starts_on: e.target.value }))
                        }
                      />
                    </Box>
                    <Box style={{ flex: "1 1 140px" }}>
                      <TextField
                        label="Fim"
                        type="date"
                        value={draft.ends_on}
                        onChange={(e) =>
                          setDraft((d) => ({ ...d, ends_on: e.target.value }))
                        }
                      />
                    </Box>
                  </Flex>
                  <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
                    <input
                      type="checkbox"
                      checked={draft.active}
                      onChange={(e) =>
                        setDraft((d) => ({ ...d, active: e.target.checked }))
                      }
                    />
                    <Text fontSize="sm">Campanha ativa (IA pode citar)</Text>
                  </label>
                  <Flex gap={2} justifyContent="flex-end">
                    {editingCampaignId ? (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          setDraft(emptyCampaign());
                          setEditingCampaignId(null);
                        }}
                      >
                        Cancelar
                      </Button>
                    ) : null}
                    <Button
                      type="button"
                      size="sm"
                      variant="secondary"
                      onClick={upsertCampaign}
                      disabled={!draft.title.trim()}
                    >
                      {editingCampaignId ? "Atualizar campanha" : "Adicionar campanha"}
                    </Button>
                  </Flex>
                </Stack>
              </Box>

              <Stack gap={2}>
                {campaigns.length === 0 ? (
                  <Text muted fontSize="sm">
                    Nenhuma campanha cadastrada.
                  </Text>
                ) : (
                  campaigns.map((c) => {
                    const status = campaignStatus(c);
                    return (
                      <Box
                        key={c.id}
                        p={3}
                        borderWidth="hairline"
                        borderStyle="solid"
                        borderColor="border.default"
                        borderRadius="md"
                      >
                        <Flex
                          justifyContent="space-between"
                          alignItems="flex-start"
                          gap={2}
                        >
                          <Stack gap={1} style={{ flex: 1, minWidth: 0 }}>
                            <Flex gap={2} alignItems="center" flexWrap="wrap">
                              <Text fontWeight="semibold">{c.title}</Text>
                              <Badge tone={status.tone}>{status.label}</Badge>
                            </Flex>
                            <Text fontSize="sm" muted style={{ whiteSpace: "pre-wrap" }}>
                              {c.description || "—"}
                            </Text>
                            <Text fontSize="xs" muted>
                              Período: {c.starts_on || "—"} → {c.ends_on || "—"}
                            </Text>
                          </Stack>
                          <Flex gap={1} flexShrink={0}>
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              onClick={() => editCampaign(c)}
                            >
                              Editar
                            </Button>
                            <Button
                              type="button"
                              variant="danger"
                              size="sm"
                              onClick={() => removeCampaign(c.id)}
                            >
                              Remover
                            </Button>
                          </Flex>
                        </Flex>
                      </Box>
                    );
                  })
                )}
              </Stack>
            </Stack>
          </Section>

          <Flex justifyContent="flex-end" gap={2}>
            <Button
              onClick={() => save.mutate()}
              disabled={save.isPending}
            >
              {save.isPending ? "Salvando…" : saved ? "Salvo!" : "Salvar configurações"}
            </Button>
          </Flex>
        </>
      )}
    </Stack>
  );
}
