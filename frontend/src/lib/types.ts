export type UserRole = "admin" | "manager" | "supervisor" | "agent";

export type PipelineStage =
  | "em_atendimento"
  | "em_negociacao"
  | "aguardando_fechamento"
  | "fechado"
  | "nao_fechado";

export type Unit = {
  id: string;
  name: string;
  code: string;
  timezone: string;
  active: boolean;
};

export type User = {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  active: boolean;
  unit_ids?: string[];
};

export type MeResponse = User & { units: Unit[] };

export type LoginResponse = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
};

export type InboxItem = {
  id: string;
  contact_name: string;
  contact_phone: string;
  unit_id: string;
  pipeline_stage: PipelineStage;
  mode: "ai_triage" | "human";
  owner_id: string | null;
  intent: string | null;
  ai_summary: string | null;
  last_message_preview: string;
  last_message_at: string;
  window_expires_at?: string | null;
};

export type Message = {
  id: string;
  conversation_id: string;
  direction: "in" | "out";
  sender_type: "contact" | "agent" | "ai" | "system";
  body: string;
  status: "accepted" | "sent" | "delivered" | "read" | "failed";
  created_at: string;
};

export type FollowUp = {
  id: string;
  conversation_id: string;
  contact_name: string;
  contact_phone: string;
  unit_id: string;
  pipeline_stage: PipelineStage;
  due_at: string;
  status: "open" | "done" | "canceled";
  note: string;
};

export type Pop = {
  id: string;
  title: string;
  body: string;
  intent_tags: string[];
  active: boolean;
};

export type LossReason = {
  code: string;
  label: string;
};

export type DashboardSummary = {
  open_conversations: number;
  by_stage: Record<PipelineStage, number>;
  closed: number;
  not_closed: number;
  conversion_rate: number;
  ai_triage: number;
  human: number;
  awaiting_followup: number;
  units: Array<{
    unit_id: string;
    unit_name: string;
    open: number;
    closed: number;
    conversion_rate: number;
  }>;
};

export type AICampaign = {
  id: string;
  title: string;
  description: string;
  starts_on: string;
  ends_on: string;
  active: boolean;
};

export type WhatsAppSettings = {
  waba_id: string;
  phone_number_id: string;
  display_phone: string;
  token_masked: string;
  ai_enabled: boolean;
  webhook_verified: boolean;
  ai_system_prompt: string;
  ai_context: string;
  ai_campaigns: AICampaign[];
};

export const STAGE_LABELS: Record<PipelineStage, string> = {
  em_atendimento: "Em atendimento",
  em_negociacao: "Em negociação",
  aguardando_fechamento: "Aguardando fechamento",
  fechado: "Fechado",
  nao_fechado: "Não fechado",
};

export const PIPELINE_STAGES: PipelineStage[] = [
  "em_atendimento",
  "em_negociacao",
  "aguardando_fechamento",
  "fechado",
  "nao_fechado",
];
