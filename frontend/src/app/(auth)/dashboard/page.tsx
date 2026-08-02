"use client";

import { useQuery } from "@tanstack/react-query";
import { animate, motion } from "framer-motion";
import { useEffect, useMemo, useState } from "react";
import styled from "styled-components";
import {
  Flex,
  PageHeader,
  Spinner,
  Stack,
  Text,
} from "@cia-da-vacina/design-system";
import { fadeUp, staggerContainer, staggerItem } from "@/lib/motion";
import { useAuth } from "@/providers/auth-provider";
import { dashboardService } from "@/services";
import {
  CHANNEL_LABELS,
  CHANNELS,
  INTENT_LABELS,
  INTENTS,
  PIPELINE_STAGES,
  STAGE_LABELS,
  type ChannelType,
  type DashboardSummary,
  type Intent,
  type PipelineStage,
} from "@/domain";

/* -------------------------------------------------------------------------- */
/* Layout                                                                     */
/* -------------------------------------------------------------------------- */

const Zone = styled.section<{ $variant: "unit" | "network" }>`
  display: flex;
  flex-direction: column;
  gap: 14px;
  width: 100%;
  padding: 16px 16px 16px 18px;
  border-radius: ${({ theme }) => theme.radii.lg};
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  background: ${({ theme, $variant }) =>
    $variant === "unit"
      ? theme.colors["bg.surface"]
      : theme.colors["bg.surface.muted"]};
  box-shadow: ${({ theme, $variant }) =>
    $variant === "unit"
      ? `inset 3px 0 0 ${theme.colors["border.brand"]}`
      : "none"};
`;

const ZoneHead = styled.div`
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 2px 2px 6px;
`;

const ZoneEyebrow = styled(Text)`
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: ${({ theme }) => theme.colors["text.brand"]};
`;

const ZoneTitle = styled(Text)`
  font-size: ${({ theme }) => theme.fontSizes.lg};
  font-weight: 650;
  letter-spacing: -0.02em;
  line-height: 1.2;
`;

const UnitTabs = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
`;

const UnitTab = styled.button<{ $active?: boolean; $crm?: boolean }>`
  appearance: none;
  border: 1px solid
    ${({ theme, $active }) =>
      $active ? theme.colors["border.brand"] : theme.colors["border.default"]};
  background: ${({ theme, $active }) =>
    $active ? theme.colors["bg.surface"] : "transparent"};
  color: ${({ theme, $active }) =>
    $active ? theme.colors["text.brand"] : theme.colors["text.secondary"]};
  border-radius: ${({ theme }) => theme.radii.md};
  padding: 8px 12px;
  font-size: ${({ theme }) => theme.fontSizes.sm};
  font-weight: ${({ $active }) => ($active ? 650 : 500)};
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition:
    background 140ms ease,
    border-color 140ms ease,
    color 140ms ease,
    transform 100ms ease;

  &:hover {
    border-color: ${({ theme }) => theme.colors["border.brand"]};
    color: ${({ theme }) => theme.colors["text.brand"]};
  }

  &:active {
    transform: scale(0.98);
  }

  box-shadow: ${({ $active, $crm, theme }) =>
    $active
      ? `0 0 0 1px ${theme.colors["border.brand"]}`
      : $crm
        ? `inset 0 0 0 1px ${theme.colors["border.subtle"]}`
        : "none"};
`;

const CrmTag = styled.span`
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding: 2px 6px;
  border-radius: 999px;
  background: ${({ theme }) => theme.colors["bg.brand.solid"]};
  color: ${({ theme }) => theme.colors["button.primary.text"]};
`;

const Grid2 = styled.div`
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  width: 100%;

  @media (min-width: 960px) {
    grid-template-columns: 1.15fr 1fr;
  }
`;

const Panel = styled(motion.section)`
  padding: 16px 18px;
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.subtle"]};
  border-radius: ${({ theme }) => theme.radii.md};
  min-width: 0;
`;

const PanelTitle = styled(Text)`
  font-weight: 600;
  font-size: ${({ theme }) => theme.fontSizes.sm};
  margin-bottom: 4px;
`;

const PanelHint = styled(Text)`
  font-size: ${({ theme }) => theme.fontSizes.xs};
  color: ${({ theme }) => theme.colors["text.muted"]};
  margin-bottom: 12px;
`;

const StatRow = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
  gap: 10px;
`;

const StatCell = styled.div`
  min-width: 0;
`;

const AttentionGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(132px, 1fr));
  gap: 8px;
`;

const AttentionChip = styled.div<{ $tone: "ok" | "warn" | "danger" }>`
  padding: 10px 12px;
  border-radius: ${({ theme }) => theme.radii.sm};
  background: ${({ theme }) => theme.colors["bg.surface.muted"]};
  border: 1px solid ${({ theme }) => theme.colors["border.subtle"]};
  border-left: 3px solid
    ${({ theme, $tone }) =>
      $tone === "danger"
        ? theme.colors["text.danger"]
        : $tone === "warn"
          ? theme.colors["text.warning"]
          : theme.colors["border.subtle"]};
`;

const InsightBox = styled.div`
  padding: 12px 14px;
  border-radius: ${({ theme }) => theme.radii.md};
  background: ${({ theme }) => theme.colors["bg.surface.muted"]};
  border: 1px solid ${({ theme }) => theme.colors["border.subtle"]};
  display: flex;
  flex-direction: column;
  gap: 6px;
`;

const BarTrack = styled.div`
  width: 100%;
  height: 10px;
  border-radius: 999px;
  background: ${({ theme }) => theme.colors["bg.surface.muted"]};
  overflow: hidden;
  display: flex;
`;

const BarFill = styled(motion.div)<{
  $tone?: "brand" | "info" | "warn" | "success" | "danger" | "ai";
}>`
  height: 100%;
  background: ${({ theme, $tone = "brand" }) => {
    const dark = theme.name === "ciaVacinaDark";
    switch ($tone) {
      case "success":
        return theme.colors["bg.brand.solid"];
      case "danger":
        return dark
          ? theme.colors["text.danger"]
          : theme.colors["bg.danger.solid"];
      case "warn":
        return dark
          ? theme.colors["text.warning"]
          : theme.colors["bg.warning.solid"];
      case "info":
        return dark
          ? theme.colors["text.link"]
          : theme.colors["bg.info.solid"];
      case "ai":
        return dark
          ? theme.colors["mode.ai.bg"]
          : theme.colors["mode.ai.text"];
      default:
        return theme.colors["bg.brand.solid"];
    }
  }};
`;

const HBarRow = styled.div`
  display: grid;
  grid-template-columns: minmax(88px, 130px) 1fr auto;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;

  &:last-child {
    margin-bottom: 0;
  }
`;

const FunnelList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 10px;
`;

const FunnelRow = styled.div<{ $emphasis?: boolean }>`
  display: grid;
  grid-template-columns: minmax(110px, 150px) 1fr auto;
  gap: 10px;
  align-items: center;
  padding: ${({ $emphasis }) => ($emphasis ? "6px 10px" : "0")};
  margin: ${({ $emphasis }) => ($emphasis ? "0 -10px" : "0")};
  border-radius: ${({ theme }) => theme.radii.sm};
  background: ${({ theme, $emphasis }) =>
    $emphasis ? theme.colors["bg.surface.muted"] : "transparent"};
  box-shadow: ${({ theme, $emphasis }) =>
    $emphasis ? `inset 3px 0 0 ${theme.colors["text.warning"]}` : "none"};
`;

const UnitTable = styled.div`
  display: flex;
  flex-direction: column;
`;

const UnitRow = styled(motion.button)<{ $focus?: boolean; $crm?: boolean }>`
  appearance: none;
  width: 100%;
  text-align: left;
  cursor: pointer;
  display: grid;
  grid-template-columns: 28px minmax(90px, 1.2fr) 1fr minmax(120px, auto);
  gap: 12px;
  align-items: center;
  padding: 12px 10px;
  margin: 0 -10px;
  border: 1px solid
    ${({ theme, $focus }) =>
      $focus ? theme.colors["border.brand"] : "transparent"};
  border-radius: ${({ theme }) => theme.radii.md};
  background: ${({ theme, $focus }) =>
    $focus ? theme.colors["bg.surface.muted"] : "transparent"};
  box-shadow: ${({ theme, $focus }) =>
    $focus ? `inset 3px 0 0 ${theme.colors["border.brand"]}` : "none"};
  color: inherit;
  transition:
    background 140ms ease,
    border-color 140ms ease,
    box-shadow 140ms ease,
    transform 100ms ease;

  &:hover {
    background: ${({ theme }) => theme.colors["bg.surface.muted"]};
    border-color: ${({ theme, $focus }) =>
      $focus ? theme.colors["border.brand"] : theme.colors["border.default"]};
    box-shadow: ${({ theme, $focus }) =>
      $focus
        ? `inset 3px 0 0 ${theme.colors["border.brand"]}`
        : `inset 3px 0 0 ${theme.colors["border.subtle"]}`};
  }

  &:active {
    transform: scale(0.995);
  }

  @media (max-width: 700px) {
    grid-template-columns: 24px 1fr;
    gap: 6px 10px;

    & > :nth-child(3),
    & > :nth-child(4) {
      grid-column: 2;
    }
  }
`;

const Rank = styled.span`
  font-size: ${({ theme }) => theme.fontSizes.xs};
  font-weight: 600;
  color: ${({ theme }) => theme.colors["text.muted"]};
`;

const SegmentTrack = styled.div`
  display: flex;
  height: 12px;
  border-radius: 999px;
  overflow: hidden;
  background: ${({ theme }) => theme.colors["bg.surface.muted"]};
`;

const Segment = styled.div<{ $pct: number; $tone: "ai" | "human" }>`
  width: ${({ $pct }) => `${$pct}%`};
  min-width: ${({ $pct }) => ($pct > 0 ? "2px" : "0")};
  background: ${({ theme, $tone }) =>
    $tone === "ai"
      ? theme.name === "ciaVacinaDark"
        ? theme.colors["mode.ai.bg"]
        : theme.colors["mode.ai.text"]
      : theme.colors["bg.brand.solid"]};
`;

const WinLossTrack = styled.div`
  width: 100%;
  height: 12px;
  border-radius: 999px;
  background: ${({ theme }) => theme.colors["bg.surface.muted"]};
  overflow: hidden;
  display: flex;
  margin-top: 8px;
`;

const WinSeg = styled(motion.div)`
  height: 100%;
  background: ${({ theme }) => theme.colors["bg.brand.solid"]};
`;

const LossSeg = styled(motion.div)`
  height: 100%;
  background: ${({ theme }) =>
    theme.name === "ciaVacinaDark"
      ? theme.colors["text.danger"]
      : theme.colors["bg.danger.solid"]};
`;

const EmptySeg = styled.div`
  width: 100%;
  height: 100%;
  background: ${({ theme }) => theme.colors["border.subtle"]};
`;

const MetaRow = styled.div`
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
`;

const NetworkStatGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 10px;
`;

const NetworkStat = styled.div`
  padding: 12px 14px;
  border-radius: ${({ theme }) => theme.radii.md};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.subtle"]};
`;

/* -------------------------------------------------------------------------- */
/* Helpers                                                                    */
/* -------------------------------------------------------------------------- */

function CountUp({
  value,
  suffix = "",
  size = "lg",
}: {
  value: number;
  suffix?: string;
  size?: "lg" | "xl" | "md";
}) {
  const [display, setDisplay] = useState(0);

  useEffect(() => {
    const controls = animate(0, value, {
      duration: 0.7,
      ease: [0.22, 1, 0.36, 1],
      onUpdate: (v) => setDisplay(Math.round(v)),
    });
    return controls.stop;
  }, [value]);

  const fontSize =
    size === "xl" ? "2.5rem" : size === "lg" ? "1.75rem" : "1.25rem";

  return (
    <span
      style={{
        display: "block",
        fontSize,
        fontWeight: 650,
        lineHeight: 1.1,
        letterSpacing: "-0.02em",
        fontVariantNumeric: "tabular-nums",
      }}
    >
      {display}
      {suffix}
    </span>
  );
}

function pct(part: number, total: number): number {
  if (total <= 0) return 0;
  return Math.round((part / total) * 100);
}

function channelConversion(data: DashboardSummary, channel: ChannelType): number {
  const won = data.closed_by_channel[channel] ?? 0;
  const lost = data.not_closed_by_channel[channel] ?? 0;
  const d = won + lost;
  return d === 0 ? 0 : Math.round((won / d) * 100);
}

function buildUnitInsights(data: DashboardSummary, unitName: string): string[] {
  const lines: string[] = [];

  if (data.decided === 0) {
    lines.push(`${unitName}: ainda sem decisões no funil.`);
  } else {
    lines.push(
      `${unitName}: ${data.conversion_rate}% de conversão (${data.closed} fechadas / ${data.not_closed} não fechadas).`,
    );
  }

  const openStages = PIPELINE_STAGES.filter(
    (s) => s !== "fechado" && s !== "nao_fechado",
  );
  let bottleneck: PipelineStage | null = null;
  let bottleneckCount = 0;
  for (const s of openStages) {
    const n = data.by_stage[s] ?? 0;
    if (n > bottleneckCount) {
      bottleneck = s;
      bottleneckCount = n;
    }
  }
  if (bottleneck && bottleneckCount > 0) {
    lines.push(
      `Gargalo local: ${STAGE_LABELS[bottleneck]} (${bottleneckCount}).`,
    );
  }

  const risks: string[] = [];
  if (data.overdue_followups > 0) {
    risks.push(`${data.overdue_followups} follow-up vencido`);
  }
  if (data.awaiting_reply > 0) {
    risks.push(`${data.awaiting_reply} aguardando resposta`);
  }
  if (data.window_expiring > 0) {
    risks.push(`${data.window_expiring} janela Meta`);
  }
  if (data.awaiting_phone > 0) {
    risks.push(`${data.awaiting_phone} sem telefone`);
  }
  if (risks.length) {
    lines.push(`Atenção nesta unidade: ${risks.join(", ")}.`);
  }

  return lines.slice(0, 3);
}

function buildNetworkInsights(data: DashboardSummary): string[] {
  const lines: string[] = [];
  const ranked = [...data.units].sort((a, b) => {
    if (b.conversion_rate !== a.conversion_rate) {
      return b.conversion_rate - a.conversion_rate;
    }
    return b.closed - a.closed;
  });
  const leader = ranked.find((u) => u.closed + u.not_closed > 0);
  if (leader) {
    lines.push(
      `${leader.unit_name} lidera a rede (${leader.conversion_rate}%, ${leader.closed} fechados).`,
    );
  }

  const stressed = [...data.units]
    .filter((u) => u.unclaimed + u.awaiting_followup > 0)
    .sort(
      (a, b) =>
        b.unclaimed + b.awaiting_followup - (a.unclaimed + a.awaiting_followup),
    );
  if (stressed[0]) {
    lines.push(
      `${stressed[0].unit_name} pede atenção: ${stressed[0].unclaimed} sem dono, ${stressed[0].awaiting_followup} follow-ups.`,
    );
  }

  lines.push(
    `Rede: ${data.open_conversations} abertas, ${data.conversion_rate}% de conversão geral.`,
  );

  return lines.slice(0, 3);
}

function toneFor(count: number): "ok" | "warn" | "danger" {
  if (count <= 0) return "ok";
  if (count >= 3) return "danger";
  return "warn";
}

function AttentionItem({
  label,
  value,
  tone,
}: {
  label: string;
  value: number;
  tone: "ok" | "warn" | "danger";
}) {
  return (
    <AttentionChip $tone={tone}>
      <Text fontSize="xs" muted>
        {label}
      </Text>
      <Text fontWeight="semibold" fontSize="lg" style={{ marginTop: 2 }}>
        {value}
      </Text>
    </AttentionChip>
  );
}

function WinLossBar({
  closed,
  notClosed,
}: {
  closed: number;
  notClosed: number;
}) {
  const total = closed + notClosed;
  const win = total ? (closed / total) * 100 : 0;
  const loss = total ? (notClosed / total) * 100 : 0;

  return (
    <WinLossTrack>
      {total > 0 ? (
        <>
          <WinSeg
            initial={{ width: 0 }}
            animate={{ width: `${win}%` }}
            transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
          />
          <LossSeg
            initial={{ width: 0 }}
            animate={{ width: `${loss}%` }}
            transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
          />
        </>
      ) : (
        <EmptySeg />
      )}
    </WinLossTrack>
  );
}

/* -------------------------------------------------------------------------- */
/* Page                                                                       */
/* -------------------------------------------------------------------------- */

export default function DashboardPage() {
  const { activeUnitId, units } = useAuth();
  const [focusUnitId, setFocusUnitId] = useState<string | null>(activeUnitId);

  useEffect(() => {
    setFocusUnitId(activeUnitId);
  }, [activeUnitId]);

  const unitQuery = useQuery({
    queryKey: ["dashboard", "unit", focusUnitId],
    queryFn: () =>
      dashboardService.getSummary({ unit_id: focusUnitId ?? undefined }),
    enabled: Boolean(focusUnitId),
  });

  const networkQuery = useQuery({
    queryKey: ["dashboard", "network"],
    queryFn: () => dashboardService.getSummary({}),
    enabled: Boolean(activeUnitId),
  });

  const data = unitQuery.data;
  const network = networkQuery.data;
  const focusName =
    units.find((u) => u.id === focusUnitId)?.name ??
    data?.units.find((u) => u.unit_id === focusUnitId)?.unit_name ??
    "Unidade";
  const crmName = units.find((u) => u.id === activeUnitId)?.name ?? "CRM";
  const peekingOther = Boolean(
    focusUnitId && activeUnitId && focusUnitId !== activeUnitId,
  );

  const unitInsights = useMemo(
    () => (data ? buildUnitInsights(data, focusName) : []),
    [data, focusName],
  );
  const networkInsights = useMemo(
    () => (network ? buildNetworkInsights(network) : []),
    [network],
  );

  const openStageMax = useMemo(() => {
    if (!data) return 1;
    return Math.max(
      1,
      ...PIPELINE_STAGES.filter(
        (s) => s !== "fechado" && s !== "nao_fechado",
      ).map((s) => data.by_stage[s] ?? 0),
    );
  }, [data]);

  const channelMax = useMemo(() => {
    if (!data) return 1;
    return Math.max(1, ...CHANNELS.map((c) => data.by_channel[c] ?? 0));
  }, [data]);

  const intentMax = useMemo(() => {
    if (!data) return 1;
    return Math.max(1, ...INTENTS.map((i) => data.by_intent[i] ?? 0));
  }, [data]);

  const sortedUnits = useMemo(() => {
    if (!network) return [];
    return [...network.units].sort((a, b) => {
      if (b.conversion_rate !== a.conversion_rate) {
        return b.conversion_rate - a.conversion_rate;
      }
      return b.closed - a.closed || b.open - a.open;
    });
  }, [network]);

  const bottleneckStage = useMemo(() => {
    if (!data) return null;
    let best: PipelineStage | null = null;
    let n = 0;
    for (const s of PIPELINE_STAGES) {
      if (s === "fechado" || s === "nao_fechado") continue;
      const v = data.by_stage[s] ?? 0;
      if (v > n) {
        best = s;
        n = v;
      }
    }
    return best;
  }, [data]);

  const loadSplit = data ? data.ai_triage + data.human : 0;
  const aiPct = data && loadSplit ? pct(data.ai_triage, loadSplit) : 0;
  const humanPct = loadSplit ? 100 - aiPct : 0;
  const winPct = data && data.decided > 0 ? pct(data.closed, data.decided) : 0;
  const lossPct = data && data.decided > 0 ? 100 - winPct : 0;

  const loading = unitQuery.isLoading || networkQuery.isLoading;

  return (
    <motion.div {...fadeUp} style={{ width: "100%" }}>
      <Stack gap={4} width="100%">
        <PageHeader
          title="Resultado"
          description="Unidade em foco e comparativo da rede. Trocar a aba abaixo não muda a unidade do CRM."
        />

        {loading || !data || !network ? (
          <Flex gap={2} alignItems="center">
            <Spinner />
            <Text muted>Carregando indicadores…</Text>
          </Flex>
        ) : (
          <motion.div
            variants={staggerContainer}
            initial="initial"
            animate="animate"
            style={{ display: "flex", flexDirection: "column", gap: 20 }}
          >
            {/* ---------- Unidade em foco ---------- */}
            <Zone $variant="unit">
              <ZoneHead>
                <ZoneEyebrow>Unidade em foco</ZoneEyebrow>
                <ZoneTitle>{focusName}</ZoneTitle>
                <Text fontSize="sm" muted>
                  {peekingOther
                    ? `Você está só olhando ${focusName}. O CRM continua em ${crmName} (seletor do topo).`
                    : `Mesma unidade do CRM (${crmName}). Use as abas para comparar sem trocar o atendimento.`}
                </Text>
              </ZoneHead>

              <UnitTabs>
                {units.map((u) => (
                  <UnitTab
                    key={u.id}
                    type="button"
                    $active={u.id === focusUnitId}
                    $crm={u.id === activeUnitId}
                    onClick={() => setFocusUnitId(u.id)}
                  >
                    {u.name}
                    {u.id === activeUnitId ? <CrmTag>CRM</CrmTag> : null}
                  </UnitTab>
                ))}
              </UnitTabs>

              <Panel variants={staggerItem}>
                <Flex
                  justifyContent="space-between"
                  alignItems="flex-start"
                  flexWrap="wrap"
                  gap={3}
                >
                  <div>
                    <Text fontSize="xs" muted>
                      Taxa de conversão · {focusName}
                    </Text>
                    <CountUp value={data.conversion_rate} suffix="%" size="xl" />
                    <Text fontSize="sm" muted style={{ marginTop: 6 }}>
                      {data.decided === 0
                        ? "Nenhuma conversa decidida ainda"
                        : `${data.decided} decididas, ${data.closed} fechadas, ${data.not_closed} não fechadas`}
                    </Text>
                  </div>
                  <StatRow style={{ flex: 1, maxWidth: 360 }}>
                    <StatCell>
                      <Text fontSize="xs" muted>
                        Fechadas
                      </Text>
                      <CountUp value={data.closed} size="md" />
                    </StatCell>
                    <StatCell>
                      <Text fontSize="xs" muted>
                        Não fechadas
                      </Text>
                      <CountUp value={data.not_closed} size="md" />
                    </StatCell>
                    <StatCell>
                      <Text fontSize="xs" muted>
                        Em aberto
                      </Text>
                      <CountUp value={data.open_conversations} size="md" />
                    </StatCell>
                  </StatRow>
                </Flex>

                <div style={{ marginTop: 16 }}>
                  <MetaRow>
                    <Text fontSize="xs">Ganhou {winPct}%</Text>
                    <Text fontSize="xs" muted>
                      Perdeu {lossPct}%
                    </Text>
                  </MetaRow>
                  <WinLossBar closed={data.closed} notClosed={data.not_closed} />
                </div>
              </Panel>

              <Grid2>
                <Panel variants={staggerItem}>
                  <PanelTitle>Operação nesta unidade</PanelTitle>
                  <PanelHint>Fila ativa e distribuição IA × humano</PanelHint>
                  <StatRow>
                    <StatCell>
                      <Text fontSize="xs" muted>
                        Abertas
                      </Text>
                      <CountUp value={data.open_conversations} size="md" />
                    </StatCell>
                    <StatCell>
                      <Text fontSize="xs" muted>
                        Sem dono
                      </Text>
                      <CountUp value={data.unclaimed} size="md" />
                    </StatCell>
                    <StatCell>
                      <Text fontSize="xs" muted>
                        IA
                      </Text>
                      <CountUp value={data.ai_triage} size="md" />
                    </StatCell>
                    <StatCell>
                      <Text fontSize="xs" muted>
                        Humano
                      </Text>
                      <CountUp value={data.human} size="md" />
                    </StatCell>
                  </StatRow>
                  <div style={{ marginTop: 16 }}>
                    <MetaRow>
                      <Text fontSize="xs" muted>
                        Triagem IA {aiPct}%
                      </Text>
                      <Text fontSize="xs" muted>
                        Humano {humanPct}%
                      </Text>
                    </MetaRow>
                    <SegmentTrack>
                      <Segment $pct={aiPct} $tone="ai" />
                      <Segment $pct={humanPct} $tone="human" />
                    </SegmentTrack>
                  </div>
                </Panel>

                <Panel variants={staggerItem}>
                  <PanelTitle>Atenção nesta unidade</PanelTitle>
                  <PanelHint>Riscos só de {focusName}</PanelHint>
                  <AttentionGrid>
                    <AttentionItem
                      label="Follow-ups vencidos"
                      value={data.overdue_followups}
                      tone={toneFor(data.overdue_followups)}
                    />
                    <AttentionItem
                      label="Aguardando resposta"
                      value={data.awaiting_reply}
                      tone={toneFor(data.awaiting_reply)}
                    />
                    <AttentionItem
                      label="Paradas +24h"
                      value={data.stale_open}
                      tone={toneFor(data.stale_open)}
                    />
                    <AttentionItem
                      label="Sem telefone"
                      value={data.awaiting_phone}
                      tone={toneFor(data.awaiting_phone)}
                    />
                    <AttentionItem
                      label="Janela Meta (4h)"
                      value={data.window_expiring}
                      tone={toneFor(data.window_expiring)}
                    />
                    <AttentionItem
                      label="Interações sociais"
                      value={data.open_engagements}
                      tone={toneFor(data.open_engagements)}
                    />
                  </AttentionGrid>
                </Panel>
              </Grid2>

              {unitInsights.length > 0 && (
                <InsightBox>
                  <Text fontSize="xs" fontWeight="semibold">
                    Leitura de {focusName}
                  </Text>
                  {unitInsights.map((line) => (
                    <Text key={line} fontSize="sm">
                      {line}
                    </Text>
                  ))}
                </InsightBox>
              )}

              <Panel variants={staggerItem}>
                <PanelTitle>Funil de {focusName}</PanelTitle>
                <PanelHint>
                  {bottleneckStage
                    ? `Maior concentração em aberto: ${STAGE_LABELS[bottleneckStage]}`
                    : "Distribuição do pipeline desta unidade"}
                </PanelHint>
                <FunnelList>
                  {PIPELINE_STAGES.map((stage) => {
                    const count = data.by_stage[stage] ?? 0;
                    const isOpen =
                      stage !== "fechado" && stage !== "nao_fechado";
                    const widthPct = isOpen
                      ? pct(count, openStageMax)
                      : pct(
                          count,
                          Math.max(
                            1,
                            data.by_stage.fechado,
                            data.by_stage.nao_fechado,
                          ),
                        );
                    const barTone =
                      stage === "fechado"
                        ? "success"
                        : stage === "nao_fechado"
                          ? "danger"
                          : stage === bottleneckStage
                            ? "warn"
                            : "brand";
                    return (
                      <FunnelRow
                        key={stage}
                        $emphasis={stage === bottleneckStage && count > 0}
                      >
                        <Text fontSize="sm">{STAGE_LABELS[stage]}</Text>
                        <BarTrack style={{ height: 14 }}>
                          <BarFill
                            $tone={barTone}
                            initial={{ width: 0 }}
                            animate={{ width: `${widthPct}%` }}
                            transition={{
                              duration: 0.55,
                              ease: [0.22, 1, 0.36, 1],
                            }}
                          />
                        </BarTrack>
                        <Text
                          fontWeight="semibold"
                          fontSize="sm"
                          style={{
                            fontVariantNumeric: "tabular-nums",
                            minWidth: 28,
                            textAlign: "right",
                          }}
                        >
                          {count}
                        </Text>
                      </FunnelRow>
                    );
                  })}
                </FunnelList>
              </Panel>

              <Grid2>
                <Panel variants={staggerItem}>
                  <PanelTitle>Canais de {focusName}</PanelTitle>
                  <PanelHint>Volume na fila e conversão nas decisões</PanelHint>
                  {CHANNELS.map((channel) => {
                    const volume = data.by_channel[channel] ?? 0;
                    const won = data.closed_by_channel[channel] ?? 0;
                    const lost = data.not_closed_by_channel[channel] ?? 0;
                    const conv = channelConversion(data, channel);
                    return (
                      <HBarRow key={channel}>
                        <div>
                          <Text fontSize="sm" fontWeight="medium">
                            {CHANNEL_LABELS[channel]}
                          </Text>
                          <Text fontSize="xs" muted>
                            {won + lost > 0
                              ? `${conv}% conv., ${won} ok / ${lost} perdidos`
                              : `${volume} na fila`}
                          </Text>
                        </div>
                        <BarTrack>
                          <BarFill
                            $tone="brand"
                            initial={{ width: 0 }}
                            animate={{ width: `${pct(volume, channelMax)}%` }}
                            transition={{
                              duration: 0.5,
                              ease: [0.22, 1, 0.36, 1],
                            }}
                          />
                        </BarTrack>
                        <Text
                          fontSize="sm"
                          fontWeight="semibold"
                          style={{ fontVariantNumeric: "tabular-nums" }}
                        >
                          {volume}
                        </Text>
                      </HBarRow>
                    );
                  })}
                </Panel>

                <Panel variants={staggerItem}>
                  <PanelTitle>Intenções em {focusName}</PanelTitle>
                  <PanelHint>Só conversas em aberto desta unidade</PanelHint>
                  {INTENTS.map((intent: Intent) => {
                    const count = data.by_intent[intent] ?? 0;
                    return (
                      <HBarRow key={intent}>
                        <Text fontSize="sm">{INTENT_LABELS[intent]}</Text>
                        <BarTrack>
                          <BarFill
                            $tone="info"
                            initial={{ width: 0 }}
                            animate={{ width: `${pct(count, intentMax)}%` }}
                            transition={{
                              duration: 0.5,
                              ease: [0.22, 1, 0.36, 1],
                            }}
                          />
                        </BarTrack>
                        <Text
                          fontSize="sm"
                          fontWeight="semibold"
                          style={{ fontVariantNumeric: "tabular-nums" }}
                        >
                          {count}
                        </Text>
                      </HBarRow>
                    );
                  })}
                </Panel>
              </Grid2>
            </Zone>

            {/* ---------- Rede ---------- */}
            <Zone $variant="network">
              <ZoneHead>
                <ZoneEyebrow>Comparativo da rede</ZoneEyebrow>
                <ZoneTitle>Todas as unidades</ZoneTitle>
                <Text fontSize="sm" muted>
                  Totais e ranking da rede. Clique numa unidade para vê-la em
                  foco acima, sem mudar o CRM.
                </Text>
              </ZoneHead>

              <NetworkStatGrid>
                <NetworkStat>
                  <Text fontSize="xs" muted>
                    Conversão da rede
                  </Text>
                  <CountUp value={network.conversion_rate} suffix="%" size="md" />
                </NetworkStat>
                <NetworkStat>
                  <Text fontSize="xs" muted>
                    Abertas na rede
                  </Text>
                  <CountUp value={network.open_conversations} size="md" />
                </NetworkStat>
                <NetworkStat>
                  <Text fontSize="xs" muted>
                    Fechadas
                  </Text>
                  <CountUp value={network.closed} size="md" />
                </NetworkStat>
                <NetworkStat>
                  <Text fontSize="xs" muted>
                    Sem dono (rede)
                  </Text>
                  <CountUp value={network.unclaimed} size="md" />
                </NetworkStat>
              </NetworkStatGrid>

              {networkInsights.length > 0 && (
                <InsightBox>
                  <Text fontSize="xs" fontWeight="semibold">
                    Leitura da rede
                  </Text>
                  {networkInsights.map((line) => (
                    <Text key={line} fontSize="sm">
                      {line}
                    </Text>
                  ))}
                </InsightBox>
              )}

              <Panel variants={staggerItem}>
                <PanelTitle>Ranking por unidade</PanelTitle>
                <PanelHint>
                  Ordenado por conversão. A linha destacada é a unidade em foco.
                </PanelHint>
                <UnitTable>
                  {sortedUnits.map((u, idx) => {
                    const isFocus = u.unit_id === focusUnitId;
                    const isCrm = u.unit_id === activeUnitId;
                    return (
                      <UnitRow
                        key={u.unit_id}
                        type="button"
                        $focus={isFocus}
                        $crm={isCrm}
                        variants={staggerItem}
                        onClick={() => setFocusUnitId(u.unit_id)}
                      >
                        <Rank>#{idx + 1}</Rank>
                        <div>
                          <Flex gap={2} alignItems="center" flexWrap="wrap">
                            <Text fontWeight="medium">{u.unit_name}</Text>
                            {isCrm ? <CrmTag>CRM</CrmTag> : null}
                            {isFocus && !isCrm ? (
                              <Text fontSize="xs" muted>
                                em foco
                              </Text>
                            ) : null}
                          </Flex>
                          <Text fontSize="xs" muted>
                            {u.open} abertas, {u.unclaimed} sem dono,{" "}
                            {u.awaiting_followup} follow-ups
                          </Text>
                        </div>
                        <BarTrack>
                          <BarFill
                            $tone="success"
                            initial={{ width: 0 }}
                            animate={{ width: `${u.conversion_rate}%` }}
                            transition={{
                              duration: 0.55,
                              ease: [0.22, 1, 0.36, 1],
                            }}
                          />
                        </BarTrack>
                        <Text fontSize="sm" style={{ textAlign: "right" }}>
                          <strong style={{ fontVariantNumeric: "tabular-nums" }}>
                            {u.conversion_rate}%
                          </strong>
                          <Text as="span" muted fontSize="sm">
                            {" "}
                            ({u.closed} ok / {u.not_closed} perdidos)
                          </Text>
                        </Text>
                      </UnitRow>
                    );
                  })}
                </UnitTable>
              </Panel>
            </Zone>
          </motion.div>
        )}
      </Stack>
    </motion.div>
  );
}
