"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { useMemo, useState } from "react";
import styled from "styled-components";
import {
  Badge,
  Button,
  DataList,
  DataListRow,
  Flex,
  Heading,
  PageHeader,
  Spinner,
  Stack,
  Text,
  TextField,
  useToast,
} from "@cia-da-vacina/design-system";
import { ModalShell } from "@/components/modal-shell";
import { fadeUp, staggerContainer, staggerItem } from "@/lib/motion";
import { useAuth } from "@/providers/auth-provider";
import { useThemeMode } from "@/providers/theme-provider";
import { campaignsService } from "@/services";
import type { AICampaign } from "@/domain";

const WEEKDAYS = ["Seg", "Ter", "Qua", "Qui", "Sex", "Sáb", "Dom"] as const;

const Layout = styled.div`
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
  width: 100%;

  @media (min-width: 1100px) {
    grid-template-columns: minmax(0, 1.55fr) minmax(300px, 0.85fr);
    align-items: start;
  }
`;

const Panel = styled.section`
  padding: 20px 22px;
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  border-radius: ${({ theme }) => theme.radii.md};
  min-width: 0;
  overflow: visible;
`;

const CalHead = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
`;

const CAL_GAP = 8;

const WeekRow = styled.div`
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: ${CAL_GAP}px;
  margin-bottom: ${CAL_GAP}px;
  padding: 0 2px;
`;

const DayGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: ${CAL_GAP}px;
  overflow: visible;
`;

type RangeStrokeFlags = {
  showLeft: boolean;
  showRight: boolean;
};

const DayCell = styled.div<{
  $today?: boolean;
  $muted?: boolean;
  $inRange?: boolean;
  $dimmed?: boolean;
  $rangeBg?: string;
}>`
  appearance: none;
  position: relative;
  z-index: ${({ $inRange }) => ($inRange ? 2 : 1)};
  text-align: left;
  min-height: 132px;
  padding: 12px 10px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 6px;
  overflow: visible;
  opacity: ${({ $dimmed }) => ($dimmed ? 0.4 : 1)};
  transition: background 140ms ease, opacity 140ms ease, border-color 140ms ease,
    filter 140ms ease;
  filter: ${({ $muted, $inRange }) =>
    $muted && !$inRange ? "saturate(0.55) brightness(0.92)" : "none"};

  background: ${({ theme, $muted, $inRange, $rangeBg }) => {
    if ($inRange) return $rangeBg ?? theme.colors["bg.brand.subtle"];
    if ($muted) return theme.colors["bg.surface.muted"];
    return theme.colors["bg.surface"];
  }};

  border: 1px solid
    ${({ theme, $today, $inRange, $muted }) =>
      $inRange
        ? "transparent"
        : $today
          ? theme.colors["border.brand"]
          : $muted
            ? theme.colors["border.subtle"]
            : theme.colors["border.default"]};
  border-radius: ${({ theme, $inRange }) => ($inRange ? "0" : theme.radii.md)};

  & > *:not([data-range-stroke]) {
    position: relative;
    z-index: 1;
  }

  &:hover {
    z-index: 4;
  }
`;

/**
 * Draws the range outline. Extends fully into the grid gap so neighboring
 * strokes meet. Left/right borders only when the campaign does not continue
 * on that side within the same week row.
 */
const RangeStroke = styled.span<{
  $showLeft: boolean;
  $showRight: boolean;
  $color: string;
}>`
  position: absolute;
  top: 0;
  bottom: 0;
  left: ${({ $showLeft }) => ($showLeft ? "0" : `-${CAL_GAP}px`)};
  right: ${({ $showRight }) => ($showRight ? "0" : `-${CAL_GAP}px`)};
  box-sizing: border-box;
  pointer-events: none;
  z-index: 0;
  border-top: 2px solid ${({ $color }) => $color};
  border-bottom: 2px solid ${({ $color }) => $color};
  border-left: ${({ $showLeft, $color }) =>
    $showLeft ? `2px solid ${$color}` : "none"};
  border-right: ${({ $showRight, $color }) =>
    $showRight ? `2px solid ${$color}` : "none"};
  border-top-left-radius: ${({ theme, $showLeft }) =>
    $showLeft ? theme.radii.md : "0"};
  border-bottom-left-radius: ${({ theme, $showLeft }) =>
    $showLeft ? theme.radii.md : "0"};
  border-top-right-radius: ${({ theme, $showRight }) =>
    $showRight ? theme.radii.md : "0"};
  border-bottom-right-radius: ${({ theme, $showRight }) =>
    $showRight ? theme.radii.md : "0"};
`;

const DayNum = styled.span<{ $muted?: boolean; $accent?: string }>`
  font-size: ${({ theme }) => theme.fontSizes.sm};
  font-weight: ${({ $muted }) => ($muted ? 500 : 650)};
  line-height: 1;
  color: ${({ theme, $muted, $accent }) =>
    $accent
      ? $accent
      : $muted
        ? theme.colors["text.muted"]
        : theme.colors["text.primary"]};
`;

const Chip = styled.button<{
  $bg: string;
  $fg: string;
  $active?: boolean;
  $faded?: boolean;
  $ring?: string;
}>`
  appearance: none;
  border: none;
  width: 100%;
  text-align: left;
  padding: 5px 8px;
  border-radius: ${({ theme }) => theme.radii.sm};
  font-size: 11px;
  line-height: 1.35;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
  opacity: ${({ $faded }) => ($faded ? 0.55 : 1)};
  color: ${({ $fg }) => $fg};
  background: ${({ $bg }) => $bg};
  outline: ${({ $active, $ring, $bg }) =>
    $active ? `2px solid ${$ring ?? $bg}` : "none"};
  outline-offset: 1px;
  transition: opacity 140ms ease, outline-color 140ms ease, filter 140ms ease;

  &:hover {
    filter: brightness(1.08);
  }
`;

const More = styled.span`
  font-size: 11px;
  font-weight: 500;
  color: ${({ theme }) => theme.colors["text.muted"]};
  padding: 2px 2px 0;
`;

const ListItem = styled.div<{ $highlighted?: boolean; $accent?: string; $subtle?: string }>`
  border-radius: ${({ theme }) => theme.radii.sm};
  outline: ${({ $highlighted, $accent }) =>
    $highlighted && $accent ? `2px solid ${$accent}` : "none"};
  background: ${({ $highlighted, $subtle }) =>
    $highlighted && $subtle ? $subtle : "transparent"};
  transition: background 140ms ease, outline-color 140ms ease;
`;

const ColorDot = styled.span<{ $color: string }>`
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: ${({ $color }) => $color};
  flex-shrink: 0;
`;

const TextArea = styled.textarea`
  width: 100%;
  min-height: 96px;
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

const CheckRow = styled.label`
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: ${({ theme }) => theme.fontSizes.sm};
  color: ${({ theme }) => theme.colors["text.primary"]};
  cursor: pointer;
`;

type Draft = {
  id: string | null;
  title: string;
  description: string;
  starts_on: string;
  ends_on: string;
  active: boolean;
};

function todayIso(): string {
  return new Date().toISOString().slice(0, 10);
}

function toIsoDate(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function monthLabel(year: number, month: number): string {
  const raw = new Date(year, month, 1).toLocaleDateString("pt-BR", {
    month: "long",
    year: "numeric",
  });
  return raw.charAt(0).toUpperCase() + raw.slice(1);
}

function campaignStatus(c: AICampaign) {
  if (!c.active) return { label: "Inativa", tone: "neutral" as const, chip: "off" as const };
  const today = todayIso();
  if (c.starts_on && today < c.starts_on) {
    return { label: "Agendada", tone: "info" as const, chip: "soon" as const };
  }
  if (c.ends_on && today > c.ends_on) {
    return { label: "Encerrada", tone: "warning" as const, chip: "done" as const };
  }
  return { label: "Vigente", tone: "success" as const, chip: "active" as const };
}

/** Stable hue 0–359 from campaign title (FNV-1a). */
function hueFromName(name: string): number {
  let hash = 2166136261;
  const raw = name.trim().toLowerCase();
  for (let i = 0; i < raw.length; i += 1) {
    hash ^= raw.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return Math.abs(hash) % 360;
}

type CampaignPalette = {
  solid: string;
  border: string;
  subtle: string;
  ink: string;
  onSolid: string;
};

function campaignPalette(
  name: string,
  chip: "active" | "soon" | "done" | "off" = "active",
  mode: "light" | "dark" = "light",
): CampaignPalette {
  const hue = hueFromName(name);
  const dark = mode === "dark";

  if (chip === "off" || chip === "done") {
    if (dark) {
      return {
        solid: `hsl(${hue} 10% 24%)`,
        border: `hsl(${hue} 12% 40%)`,
        subtle: `hsl(${hue} 12% 16% / 0.55)`,
        ink: `hsl(${hue} 14% 70%)`,
        onSolid: `hsl(${hue} 12% 78%)`,
      };
    }
    return {
      solid: `hsl(${hue} 10% 88%)`,
      border: `hsl(${hue} 12% 72%)`,
      subtle: `hsl(${hue} 14% 96%)`,
      ink: `hsl(${hue} 12% 42%)`,
      onSolid: `hsl(${hue} 14% 34%)`,
    };
  }

  if (dark) {
    const sat = chip === "soon" ? 34 : 40;
    return {
      solid: `hsl(${hue} ${sat}% 26%)`,
      border: `hsl(${hue} ${sat + 10}% 48%)`,
      subtle: `hsl(${hue} 28% 18% / 0.55)`,
      ink: `hsl(${hue} ${sat + 8}% 78%)`,
      onSolid: `hsl(${hue} 45% 90%)`,
    };
  }

  const sat = chip === "soon" ? 28 : 32;
  return {
    solid: `hsl(${hue} ${sat}% 90%)`,
    border: `hsl(${hue} ${sat + 8}% 62%)`,
    subtle: `hsl(${hue} 24% 95%)`,
    ink: `hsl(${hue} ${sat + 10}% 34%)`,
    onSolid: `hsl(${hue} ${sat + 12}% 28%)`,
  };
}

function overlapsDay(c: AICampaign, dayIso: string): boolean {
  const start = c.starts_on || dayIso;
  const end = c.ends_on || start;
  return start <= dayIso && dayIso <= end;
}

function addDaysIso(iso: string, delta: number): string {
  const d = new Date(`${iso}T12:00:00`);
  d.setDate(d.getDate() + delta);
  return toIsoDate(d);
}

/**
 * Per week-row stroke flags.
 * - Continues from previous day in the same row → no left border
 * - Continues into next day in the same row → no right border
 * Works across months (e.g. Jul 31 → Aug 1) as long as they share the row.
 * New week row always restarts left border (Mon after Sun).
 */
function rangeStrokeForDay(
  campaign: AICampaign | null,
  dayIso: string,
  cellIndex: number,
): RangeStrokeFlags | null {
  if (!campaign || !overlapsDay(campaign, dayIso)) return null;

  const col = cellIndex % 7;
  const prevInSameRow =
    col > 0 && overlapsDay(campaign, addDaysIso(dayIso, -1));
  const nextInSameRow =
    col < 6 && overlapsDay(campaign, addDaysIso(dayIso, 1));

  return {
    showLeft: !prevInSameRow,
    showRight: !nextInSameRow,
  };
}

function overlapsMonth(c: AICampaign, year: number, month: number): boolean {
  const first = toIsoDate(new Date(year, month, 1));
  const last = toIsoDate(new Date(year, month + 1, 0));
  const start = c.starts_on || first;
  const end = c.ends_on || start;
  return start <= last && end >= first;
}

function buildMonthCells(year: number, month: number) {
  const first = new Date(year, month, 1);
  // Monday-first: JS getDay() Sun=0 → shift
  const startOffset = (first.getDay() + 6) % 7;
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const prevDays = new Date(year, month, 0).getDate();

  const cells: { date: Date; inMonth: boolean }[] = [];
  for (let i = startOffset - 1; i >= 0; i--) {
    cells.push({
      date: new Date(year, month - 1, prevDays - i),
      inMonth: false,
    });
  }
  for (let d = 1; d <= daysInMonth; d++) {
    cells.push({ date: new Date(year, month, d), inMonth: true });
  }
  while (cells.length % 7 !== 0) {
    const next = cells.length - (startOffset + daysInMonth) + 1;
    cells.push({ date: new Date(year, month + 1, next), inMonth: false });
  }
  return cells;
}

function blankDraft(startsOn?: string): Draft {
  const start = startsOn || todayIso();
  return {
    id: null,
    title: "",
    description: "",
    starts_on: start,
    ends_on: start,
    active: true,
  };
}

function fromCampaign(c: AICampaign): Draft {
  return {
    id: c.id,
    title: c.title,
    description: c.description,
    starts_on: c.starts_on,
    ends_on: c.ends_on,
    active: c.active,
  };
}

export default function CampaignsPage() {
  const qc = useQueryClient();
  const toast = useToast();
  const { user } = useAuth();
  const { mode } = useThemeMode();
  const canWrite = user?.role === "admin" || user?.role === "manager";

  const now = new Date();
  const [cursor, setCursor] = useState({
    year: now.getFullYear(),
    month: now.getMonth(),
  });
  const [draft, setDraft] = useState<Draft | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [hoveredCampaignId, setHoveredCampaignId] = useState<string | null>(null);

  const { data: campaigns = [], isLoading } = useQuery({
    queryKey: ["campaigns"],
    queryFn: () => campaignsService.list(),
  });

  const cells = useMemo(
    () => buildMonthCells(cursor.year, cursor.month),
    [cursor.year, cursor.month],
  );

  const monthCampaigns = useMemo(
    () =>
      [...campaigns]
        .filter((c) => overlapsMonth(c, cursor.year, cursor.month))
        .sort((a, b) => a.starts_on.localeCompare(b.starts_on)),
    [campaigns, cursor.year, cursor.month],
  );

  const save = useMutation({
    mutationFn: async (next: AICampaign[]) => campaignsService.save(next),
    onSuccess: async () => {
      toast.push("Campanhas atualizadas", "success");
      setDraft(null);
      setFormError(null);
      await qc.invalidateQueries({ queryKey: ["campaigns"] });
      await qc.invalidateQueries({ queryKey: ["meta-settings"] });
    },
    onError: () => {
      toast.push("Não foi possível salvar", "danger");
    },
  });

  function openCreate(dayIso?: string) {
    if (!canWrite) return;
    setFormError(null);
    setDraft(blankDraft(dayIso));
  }

  function openEdit(c: AICampaign) {
    setFormError(null);
    setDraft(fromCampaign(c));
  }

  function submitDraft() {
    if (!draft || !canWrite) return;
    const title = draft.title.trim();
    if (!title) {
      setFormError("Informe um título.");
      return;
    }
    if (!draft.starts_on || !draft.ends_on) {
      setFormError("Informe início e fim.");
      return;
    }
    if (draft.ends_on < draft.starts_on) {
      setFormError("A data de fim deve ser igual ou posterior ao início.");
      return;
    }

    const item: AICampaign = {
      id: draft.id ?? `camp-${crypto.randomUUID()}`,
      title,
      description: draft.description.trim(),
      starts_on: draft.starts_on,
      ends_on: draft.ends_on,
      active: draft.active,
    };

    const next = draft.id
      ? campaigns.map((c) => (c.id === draft.id ? item : c))
      : [...campaigns, item];

    save.mutate(next);
  }

  function removeDraft() {
    if (!draft?.id || !canWrite) return;
    const next = campaigns.filter((c) => c.id !== draft.id);
    save.mutate(next);
  }

  const hoveredCampaign = useMemo(
    () => campaigns.find((c) => c.id === hoveredCampaignId) ?? null,
    [campaigns, hoveredCampaignId],
  );

  const hoveredPalette = useMemo(() => {
    if (!hoveredCampaign) return null;
    return campaignPalette(
      hoveredCampaign.title,
      campaignStatus(hoveredCampaign).chip,
      mode,
    );
  }, [hoveredCampaign, mode]);

  const today = todayIso();

  return (
    <motion.div {...fadeUp} style={{ width: "100%" }}>
      <Stack gap={3} width="100%">
        <PageHeader
          title="Agenda"
          description="Campanhas e promoções no calendário. A IA usa as vigentes na triagem."
          actions={
            canWrite ? (
              <Button type="button" onClick={() => openCreate()}>
                Nova campanha
              </Button>
            ) : undefined
          }
        />

        {isLoading ? (
          <Flex gap={2} alignItems="center">
            <Spinner />
            <Text muted>Carregando agenda…</Text>
          </Flex>
        ) : (
          <Layout>
            <Panel>
              <CalHead>
                <div>
                  <Text fontWeight="semibold" fontSize="lg">
                    {monthLabel(cursor.year, cursor.month)}
                  </Text>
                  {hoveredCampaign ? (
                    <Text fontSize="xs" muted style={{ marginTop: 4 }}>
                      Destacando: {hoveredCampaign.title} ({hoveredCampaign.starts_on} a{" "}
                      {hoveredCampaign.ends_on})
                    </Text>
                  ) : (
                    <Text fontSize="xs" muted style={{ marginTop: 4 }}>
                      Passe o mouse na lista ao lado para destacar o período
                    </Text>
                  )}
                </div>
                <Flex gap={1}>
                  <Button
                    type="button"
                    size="sm"
                    variant="secondary"
                    onClick={() =>
                      setCursor((c) => {
                        const d = new Date(c.year, c.month - 1, 1);
                        return { year: d.getFullYear(), month: d.getMonth() };
                      })
                    }
                  >
                    Anterior
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="ghost"
                    onClick={() =>
                      setCursor({
                        year: now.getFullYear(),
                        month: now.getMonth(),
                      })
                    }
                  >
                    Hoje
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="secondary"
                    onClick={() =>
                      setCursor((c) => {
                        const d = new Date(c.year, c.month + 1, 1);
                        return { year: d.getFullYear(), month: d.getMonth() };
                      })
                    }
                  >
                    Próximo
                  </Button>
                </Flex>
              </CalHead>

              <WeekRow>
                {WEEKDAYS.map((d) => (
                  <Text
                    key={d}
                    fontSize="xs"
                    fontWeight="semibold"
                    muted
                    style={{ textAlign: "center", letterSpacing: "0.02em" }}
                  >
                    {d}
                  </Text>
                ))}
              </WeekRow>

              <DayGrid>
                {cells.map(({ date, inMonth }, cellIndex) => {
                  const iso = toIsoDate(date);
                  const dayCamps = campaigns.filter((c) => overlapsDay(c, iso));
                  const ordered = hoveredCampaignId
                    ? [
                        ...dayCamps.filter((c) => c.id === hoveredCampaignId),
                        ...dayCamps.filter((c) => c.id !== hoveredCampaignId),
                      ]
                    : dayCamps;
                  const shown = ordered.slice(0, 2);
                  const extra = Math.max(0, ordered.length - shown.length);
                  const inRange = hoveredCampaign
                    ? overlapsDay(hoveredCampaign, iso)
                    : false;
                  const dimmed = Boolean(hoveredCampaign && !inRange);
                  const stroke = rangeStrokeForDay(
                    hoveredCampaign,
                    iso,
                    cellIndex,
                  );
                  return (
                    <DayCell
                      key={iso + String(inMonth)}
                      role="button"
                      tabIndex={0}
                      $today={iso === today}
                      $muted={!inMonth}
                      $inRange={inRange}
                      $dimmed={dimmed}
                      $rangeBg={hoveredPalette?.subtle}
                      onClick={() => openCreate(iso)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault();
                          openCreate(iso);
                        }
                      }}
                      aria-label={`Dia ${iso}`}
                    >
                      {stroke && hoveredPalette ? (
                        <RangeStroke
                          data-range-stroke
                          $showLeft={stroke.showLeft}
                          $showRight={stroke.showRight}
                          $color={hoveredPalette.border}
                        />
                      ) : null}
                      <DayNum
                        $muted={!inMonth}
                        $accent={inRange ? hoveredPalette?.ink : undefined}
                      >
                        {date.getDate()}
                      </DayNum>
                      {shown.map((c) => {
                        const st = campaignStatus(c);
                        const palette = campaignPalette(c.title, st.chip, mode);
                        const isHovered = hoveredCampaignId === c.id;
                        return (
                          <Chip
                            key={c.id}
                            type="button"
                            $bg={palette.solid}
                            $fg={palette.onSolid}
                            $ring={palette.border}
                            $active={isHovered}
                            $faded={
                              !inMonth || Boolean(hoveredCampaignId && !isHovered)
                            }
                            title={`${c.title}${c.active ? "" : " (rascunho)"}`}
                            onClick={(e) => {
                              e.stopPropagation();
                              openEdit(c);
                            }}
                          >
                            {c.title}
                          </Chip>
                        );
                      })}
                      {extra > 0 ? <More>+{extra} mais</More> : null}
                    </DayCell>
                  );
                })}
              </DayGrid>
            </Panel>

            <Panel>
              <Text fontWeight="semibold" fontSize="sm" style={{ marginBottom: 4 }}>
                Neste mês
              </Text>
              <Text fontSize="xs" muted style={{ marginBottom: 14 }}>
                {monthCampaigns.length === 0
                  ? "Nenhuma campanha neste mês."
                  : `${monthCampaigns.length} campanha(s) no período`}
              </Text>

              {monthCampaigns.length === 0 ? null : (
                <motion.div variants={staggerContainer} initial="initial" animate="animate">
                  <DataList>
                    {monthCampaigns.map((c) => {
                      const st = campaignStatus(c);
                      const palette = campaignPalette(c.title, st.chip, mode);
                      return (
                        <motion.div key={c.id} variants={staggerItem}>
                          <ListItem
                            $highlighted={hoveredCampaignId === c.id}
                            $accent={palette.border}
                            $subtle={palette.subtle}
                            onMouseEnter={() => setHoveredCampaignId(c.id)}
                            onMouseLeave={() => setHoveredCampaignId(null)}
                          >
                            <DataListRow
                              interactive
                              onClick={() => openEdit(c)}
                              leading={
                                <Stack gap={1} style={{ minWidth: 0 }}>
                                  <Flex gap={2} alignItems="center" flexWrap="wrap">
                                    <ColorDot $color={palette.solid} aria-hidden />
                                    <Text fontWeight="semibold">{c.title}</Text>
                                    <Badge tone={st.tone}>{st.label}</Badge>
                                  </Flex>
                                  <Text fontSize="xs" muted>
                                    {c.starts_on} a {c.ends_on}
                                  </Text>
                                  {c.description ? (
                                    <Text fontSize="sm" muted>
                                      {c.description}
                                    </Text>
                                  ) : null}
                                </Stack>
                              }
                            />
                          </ListItem>
                        </motion.div>
                      );
                    })}
                  </DataList>
                </motion.div>
              )}
            </Panel>
          </Layout>
        )}
      </Stack>

      <ModalShell
        open={Boolean(draft)}
        onClose={() => {
          if (!save.isPending) setDraft(null);
        }}
        closeOnBackdrop={!save.isPending}
        maxWidth="480px"
      >
        {draft ? (
          <Stack gap={3}>
            <Heading as="h3">
              {draft.id ? "Editar campanha" : "Nova campanha"}
            </Heading>

            {!canWrite ? (
              <Text muted fontSize="sm">
                Você pode visualizar. Apenas admin ou gerente edita campanhas.
              </Text>
            ) : null}

            <TextField
              label="Título"
              value={draft.title}
              onChange={(e) =>
                setDraft((d) => (d ? { ...d, title: e.target.value } : d))
              }
              disabled={!canWrite || save.isPending}
              placeholder="Ex.: Campanha gripe"
            />

            <div>
              <Text fontSize="xs" muted style={{ marginBottom: 6 }}>
                Descrição (o que a IA pode mencionar)
              </Text>
              <TextArea
                value={draft.description}
                onChange={(e) =>
                  setDraft((d) =>
                    d ? { ...d, description: e.target.value } : d,
                  )
                }
                disabled={!canWrite || save.isPending}
                placeholder="Desconto, pacote, validade…"
              />
            </div>

            <Flex gap={2} flexWrap="wrap">
              <div style={{ flex: 1, minWidth: 140 }}>
                <TextField
                  label="Início"
                  type="date"
                  value={draft.starts_on}
                  onChange={(e) =>
                    setDraft((d) =>
                      d ? { ...d, starts_on: e.target.value } : d,
                    )
                  }
                  disabled={!canWrite || save.isPending}
                />
              </div>
              <div style={{ flex: 1, minWidth: 140 }}>
                <TextField
                  label="Fim"
                  type="date"
                  value={draft.ends_on}
                  onChange={(e) =>
                    setDraft((d) =>
                      d ? { ...d, ends_on: e.target.value } : d,
                    )
                  }
                  disabled={!canWrite || save.isPending}
                />
              </div>
            </Flex>

            <CheckRow>
              <input
                type="checkbox"
                checked={draft.active}
                disabled={!canWrite || save.isPending}
                onChange={(e) =>
                  setDraft((d) =>
                    d ? { ...d, active: e.target.checked } : d,
                  )
                }
              />
              Campanha ativa (visível para a IA quando vigente)
            </CheckRow>

            {formError ? (
              <Text color="text.danger" fontSize="sm">
                {formError}
              </Text>
            ) : null}

            {canWrite ? (
              <Flex gap={2} justifyContent="space-between">
                {draft.id ? (
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={removeDraft}
                    disabled={save.isPending}
                  >
                    Excluir
                  </Button>
                ) : (
                  <span />
                )}
                <Flex gap={2}>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => setDraft(null)}
                    disabled={save.isPending}
                  >
                    Cancelar
                  </Button>
                  <Button
                    type="button"
                    onClick={submitDraft}
                    disabled={save.isPending}
                  >
                    {save.isPending ? "Salvando…" : "Salvar"}
                  </Button>
                </Flex>
              </Flex>
            ) : null}
          </Stack>
        ) : null}
      </ModalShell>
    </motion.div>
  );
}
