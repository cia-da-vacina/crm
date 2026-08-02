import type { Customer } from "./customer";
import type {
  ChannelType,
  ConversationMode,
  ConversationStatus,
  CustomerIdentification,
  Intent,
  PhoneGateStatus,
  PipelineStage,
} from "./enums";

/**
 * Shared shape between the inbox list item and the full conversation
 * detail. A conversation always belongs to exactly one channel thread.
 * The CRM-level customer may be anonymous (no phone) or identified.
 */
export interface ConversationSummary {
  id: string;
  customer_id: string;
  customer_name: string;
  /** Present only when the customer has crossed the phone privacy wall. */
  customer_phone?: string | null;
  /** Mirrors `Customer.identification` for list rendering without extra fetch. */
  identification: CustomerIdentification;
  /**
   * Whether this thread currently needs a phone (and WhatsApp OTP) before
   * gated actions. WhatsApp threads are typically `not_needed` / `collected`.
   */
  phone_gate: PhoneGateStatus;
  /** Masked pending number while awaiting WhatsApp OTP confirmation. */
  pending_phone_masked?: string | null;
  channel: ChannelType;
  /** Meta-side thread identifier (WA conversation id, IG/FB thread id). */
  channel_thread_id?: string | null;
  unit_id: string;
  pipeline_stage: PipelineStage;
  mode: ConversationMode;
  status: ConversationStatus;
  owner_id: string | null;
  intent: Intent | null;
  ai_summary: string | null;
  /** Free-form notes captured by the AI triage flow before handoff. */
  triage_notes?: string | null;
  last_message_preview: string;
  last_message_at: string;
  /** WhatsApp/Meta 24h customer-service-window expiry, when applicable. */
  window_expires_at?: string | null;
  unread_count?: number;
}

/** A row in the inbox list. */
export type InboxItem = ConversationSummary;

/** Full conversation payload used on the conversation detail screen. */
export interface ConversationDetail extends ConversationSummary {
  customer: Customer;
  created_at: string;
  updated_at: string;
}
