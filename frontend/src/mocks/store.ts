import type {
  ChannelType,
  ConversationDetail,
  Customer,
  DashboardSummary,
  FollowUp,
  Intent,
  Message,
  MetaSettings,
  PipelineStage,
  Pop,
  SocialEngagement,
  User,
} from "@/domain";
import { CHANNELS, INTENTS } from "@/domain";
import {
  lossReasons,
  mockAdmin,
  mockAgent,
  mockConversations,
  mockCustomers,
  mockEngagements,
  mockFollowUps,
  mockMessages,
  mockMetaSettings,
  mockPops,
  mockUnits,
} from "./data";

export const db = {
  users: structuredClone([mockAdmin, mockAgent]) as User[],
  customers: structuredClone(mockCustomers) as Customer[],
  conversations: structuredClone(mockConversations) as ConversationDetail[],
  messages: structuredClone(mockMessages) as Record<string, Message[]>,
  followups: structuredClone(mockFollowUps) as FollowUp[],
  pops: structuredClone(mockPops) as Pop[],
  engagements: structuredClone(mockEngagements) as SocialEngagement[],
  settings: structuredClone(mockMetaSettings) as MetaSettings,
};

export { lossReasons, mockAdmin, mockAgent, mockUnits };

export function userFromToken(token: string | null | undefined): User {
  if (token?.includes("admin")) return mockAdmin;
  return mockAgent;
}

export function unitsForUser(user: User) {
  if (user.role === "admin") return mockUnits;
  return mockUnits.filter((u) => user.unit_ids?.includes(u.id));
}

const STALE_MS = 24 * 3600_000;
const WINDOW_EXPIRING_MS = 4 * 3600_000;

function isOpenStage(stage: PipelineStage): boolean {
  return stage !== "fechado" && stage !== "nao_fechado";
}

function emptyChannelCounts(): Record<ChannelType, number> {
  return { whatsapp: 0, instagram: 0, facebook: 0 };
}

function emptyIntentCounts(): Record<Intent, number> {
  return { agendar: 0, precos: 0, duvidas: 0, reclamacao: 0, outro: 0 };
}

function lastMessageDirection(conversationId: string): Message["direction"] | null {
  const list = db.messages[conversationId];
  if (!list?.length) return null;
  const sorted = [...list].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  );
  return sorted[sorted.length - 1]?.direction ?? null;
}

export function buildDashboard(unitId?: string | null): DashboardSummary {
  const now = Date.now();
  const items = unitId
    ? db.conversations.filter((c) => c.unit_id === unitId)
    : db.conversations;

  const by_stage: Record<PipelineStage, number> = {
    em_atendimento: 0,
    em_negociacao: 0,
    aguardando_fechamento: 0,
    fechado: 0,
    nao_fechado: 0,
  };
  const by_channel = emptyChannelCounts();
  const closed_by_channel = emptyChannelCounts();
  const not_closed_by_channel = emptyChannelCounts();
  const by_intent = emptyIntentCounts();

  let unclaimed = 0;
  let awaiting_reply = 0;
  let stale_open = 0;
  let awaiting_phone = 0;
  let window_expiring = 0;
  let ai_triage = 0;
  let human = 0;

  for (const c of items) {
    by_stage[c.pipeline_stage] += 1;
    by_channel[c.channel] += 1;

    if (c.pipeline_stage === "fechado") closed_by_channel[c.channel] += 1;
    if (c.pipeline_stage === "nao_fechado") not_closed_by_channel[c.channel] += 1;

    if (!isOpenStage(c.pipeline_stage)) continue;

    const intentKey: Intent = c.intent && INTENTS.includes(c.intent) ? c.intent : "outro";
    by_intent[intentKey] += 1;

    if (!c.owner_id) unclaimed += 1;
    if (c.mode === "ai_triage") ai_triage += 1;
    if (c.mode === "human") human += 1;
    if (c.phone_gate === "required" || c.phone_gate === "pending_verification") {
      awaiting_phone += 1;
    }

    const lastAt = new Date(c.last_message_at).getTime();
    if (now - lastAt > STALE_MS) stale_open += 1;

    if (c.window_expires_at) {
      const expires = new Date(c.window_expires_at).getTime();
      if (expires >= now && expires - now <= WINDOW_EXPIRING_MS) {
        window_expiring += 1;
      }
    }

    if (lastMessageDirection(c.id) === "in") awaiting_reply += 1;
  }

  const closed = by_stage.fechado;
  const not_closed = by_stage.nao_fechado;
  const decided = closed + not_closed;
  const open = items.filter((c) => isOpenStage(c.pipeline_stage)).length;

  const followupsScoped = unitId
    ? db.followups.filter((f) => f.unit_id === unitId)
    : db.followups;
  const openFollowups = followupsScoped.filter((f) => f.status === "open");
  const awaiting_followup = openFollowups.length;
  const overdue_followups = openFollowups.filter(
    (f) => new Date(f.due_at).getTime() < now,
  ).length;

  const engagementsScoped = unitId
    ? db.engagements.filter((e) => e.unit_id === unitId)
    : db.engagements;
  const open_engagements = engagementsScoped.filter((e) => e.status === "open").length;

  const units = mockUnits.map((u) => {
    const uItems = db.conversations.filter((c) => c.unit_id === u.id);
    const uOpenItems = uItems.filter((c) => isOpenStage(c.pipeline_stage));
    const uClosed = uItems.filter((c) => c.pipeline_stage === "fechado").length;
    const uLost = uItems.filter((c) => c.pipeline_stage === "nao_fechado").length;
    const d = uClosed + uLost;
    return {
      unit_id: u.id,
      unit_name: u.name,
      open: uOpenItems.length,
      closed: uClosed,
      not_closed: uLost,
      conversion_rate: d === 0 ? 0 : Math.round((uClosed / d) * 100),
      unclaimed: uOpenItems.filter((c) => !c.owner_id).length,
      awaiting_followup: db.followups.filter(
        (f) => f.unit_id === u.id && f.status === "open",
      ).length,
    };
  });

  // silence unused CHANNELS import warning if tree-shaken — keep for contract parity docs
  void CHANNELS;

  return {
    open_conversations: open,
    by_stage,
    by_channel,
    closed,
    not_closed,
    decided,
    conversion_rate: decided === 0 ? 0 : Math.round((closed / decided) * 100),
    ai_triage,
    human,
    unclaimed,
    awaiting_reply,
    stale_open,
    awaiting_phone,
    window_expiring,
    awaiting_followup,
    overdue_followups,
    open_engagements,
    by_intent,
    closed_by_channel,
    not_closed_by_channel,
    units,
  };
}

export function newId(prefix: string): string {
  return `${prefix}-${crypto.randomUUID()}`;
}
