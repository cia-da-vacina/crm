import type { Intent, PhoneGateStatus } from "./enums";

/**
 * Structured output produced by the AI triage flow for a conversation,
 * used to brief the human agent (or the AI itself) before/at handoff.
 *
 * Phone collection is gated: only when the backend sets `phone_gate` to
 * `required` (typically Instagram/Facebook + intents that need a real
 * customer record). After the number is submitted, gate moves to
 * `pending_verification` until a WhatsApp OTP confirms possession.
 * WhatsApp threads never ask — Meta already provides the number.
 */
export interface TriageSummary {
  conversation_id: string;
  intent: Intent;
  /** 0-1 confidence score from the triage model, when available. */
  confidence?: number;
  summary: string;
  /** POP ids the AI believes are relevant to answer/close this conversation. */
  suggested_pops?: string[];
  ready_for_handoff: boolean;
  /**
   * Whether a phone must be collected / confirmed before gated CRM data unlocks.
   * Driven entirely by backend rules — frontend never decides this.
   */
  phone_gate: PhoneGateStatus;
  /**
   * Masked phone waiting for WhatsApp OTP, when `phone_gate === "pending_verification"`.
   * Example: `+55•••••8888`. Never the full number in client payloads if avoidable.
   */
  pending_phone_masked?: string | null;
  /**
   * Structured facts collected during triage.
   * Full `phone_e164` may exist server-side during pending verification but
   * should not be required in this payload for the agent UI.
   * Example: { vacina: "gripe", unidade: "Centro" }.
   */
  collected_fields?: Record<string, string>;
}
