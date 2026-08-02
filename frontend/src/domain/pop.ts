import type { Intent } from "./enums";

/** POP: pre-approved response/script used by agents and by the AI triage flow. */
export interface Pop {
  id: string;
  title: string;
  body: string;
  intent_tags: Intent[];
  active: boolean;
  created_at?: string;
  updated_at?: string;
}

/** Reason code selected when a conversation's pipeline stage moves to `nao_fechado`. */
export interface LossReason {
  code: string;
  label: string;
}
