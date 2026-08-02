import type { ChannelType, Intent, PipelineStage } from "./enums";

export interface DashboardUnitSummary {
  unit_id: string;
  unit_name: string;
  open: number;
  closed: number;
  not_closed: number;
  conversion_rate: number;
  unclaimed: number;
  awaiting_followup: number;
}

/**
 * Operational snapshot for managers. All fields are derived from live CRM
 * flow (conversations, messages, follow-ups, engagements) — no financials.
 */
export interface DashboardSummary {
  open_conversations: number;
  by_stage: Record<PipelineStage, number>;
  by_channel: Record<ChannelType, number>;
  closed: number;
  not_closed: number;
  /** closed + not_closed — denominator of conversion_rate. */
  decided: number;
  conversion_rate: number;
  ai_triage: number;
  human: number;
  /** Open conversations with no owner_id. */
  unclaimed: number;
  /**
   * Open conversations whose latest message is inbound
   * (customer waiting for a reply — soft SLA signal).
   */
  awaiting_reply: number;
  /** Open conversations with last_message_at older than 24h. */
  stale_open: number;
  /** Threads blocked on phone_gate required | pending_verification. */
  awaiting_phone: number;
  /** Threads whose Meta 24h window expires within the next 4 hours. */
  window_expiring: number;
  awaiting_followup: number;
  /** Open follow-ups with due_at in the past. */
  overdue_followups: number;
  /** Unresolved story replies/mentions/comments awaiting triage. */
  open_engagements: number;
  /** Open conversations grouped by AI intent (null → outro). */
  by_intent: Record<Intent, number>;
  closed_by_channel: Record<ChannelType, number>;
  not_closed_by_channel: Record<ChannelType, number>;
  units: DashboardUnitSummary[];
}
