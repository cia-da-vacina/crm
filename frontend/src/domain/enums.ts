/**
 * Central source of truth for all domain-level union types (enums) and their
 * human-readable (pt-BR) labels. Keep this file framework-agnostic: it must
 * be safe to import from client components, server components and route
 * handlers alike.
 */

export type UserRole = "admin" | "manager" | "supervisor" | "agent";

export type PipelineStage =
  | "em_atendimento"
  | "em_negociacao"
  | "aguardando_fechamento"
  | "fechado"
  | "nao_fechado";

export type ConversationStatus = "open" | "pending" | "resolved";

export type ConversationMode = "ai_triage" | "human";

export type MessageDirection = "in" | "out";

export type SenderType = "contact" | "agent" | "ai" | "system";

export type MessageStatus =
  | "accepted"
  | "sent"
  | "delivered"
  | "read"
  | "failed";

export type FollowUpStatus = "open" | "done" | "canceled";

export type Intent = "agendar" | "precos" | "duvidas" | "reclamacao" | "outro";

/**
 * Meta channels supported by the CRM. "facebook" always refers to Messenger
 * (Page inbox), never the Facebook feed/ads product.
 */
export type ChannelType = "whatsapp" | "instagram" | "facebook";

/**
 * Social engagements are interactions Meta surfaces outside of the normal
 * 1:1 messaging thread (stories, comments, live broadcasts). They are
 * triaged separately from `ConversationDetail` and may later be converted
 * into a full conversation once a human/AI responds and a thread exists.
 */
export type EngagementType =
  | "story_reply"
  | "story_mention"
  | "post_comment"
  | "live_comment"
  | "private_reply";

export type EngagementStatus =
  | "open"
  | "replied"
  | "dismissed"
  | "converted_to_conversation";

export type MessageKind =
  | "text"
  | "template"
  | "image"
  | "audio"
  | "story_reply"
  | "comment_reply"
  | "system";

/**
 * Whether the CRM has a verified phone for this customer.
 * Controls the privacy wall: anonymous conversations only see/public data;
 * identified unlocks history, scheduling and cross-channel merge.
 *
 * A phone typed on IG/FB does NOT identify the customer until WhatsApp
 * OTP confirmation succeeds — until then they remain `anonymous`.
 */
export type CustomerIdentification = "anonymous" | "identified";

/**
 * Phone collection / verification gate for a conversation.
 * Backend owns transitions; frontend only displays and blocks gated UX.
 *
 * Flow on Instagram/Facebook when intent requires identity:
 *   required → (user gives number) → pending_verification → (WA OTP ok) → collected
 * WhatsApp threads skip this (Meta already proved possession of the number).
 */
export type PhoneGateStatus =
  | "not_needed"
  | "required"
  | "pending_verification"
  | "collected";

export const USER_ROLES: readonly UserRole[] = [
  "admin",
  "manager",
  "supervisor",
  "agent",
];

export const USER_ROLE_LABELS: Record<UserRole, string> = {
  admin: "Administrador",
  manager: "Gerente",
  supervisor: "Supervisor",
  agent: "Atendente",
};

export const STAGE_LABELS: Record<PipelineStage, string> = {
  em_atendimento: "Em atendimento",
  em_negociacao: "Em negociação",
  aguardando_fechamento: "Aguardando fechamento",
  fechado: "Fechado",
  nao_fechado: "Não fechado",
};

export const PIPELINE_STAGES: readonly PipelineStage[] = [
  "em_atendimento",
  "em_negociacao",
  "aguardando_fechamento",
  "fechado",
  "nao_fechado",
];

export const CONVERSATION_STATUS_LABELS: Record<ConversationStatus, string> = {
  open: "Aberta",
  pending: "Pendente",
  resolved: "Resolvida",
};

export const CONVERSATION_MODE_LABELS: Record<ConversationMode, string> = {
  ai_triage: "Triagem por IA",
  human: "Atendimento humano",
};

export const MESSAGE_STATUS_LABELS: Record<MessageStatus, string> = {
  accepted: "Aceita",
  sent: "Enviada",
  delivered: "Entregue",
  read: "Lida",
  failed: "Falhou",
};

export const FOLLOW_UP_STATUS_LABELS: Record<FollowUpStatus, string> = {
  open: "Em aberto",
  done: "Concluído",
  canceled: "Cancelado",
};

export const INTENT_LABELS: Record<Intent, string> = {
  agendar: "Agendar",
  precos: "Preços",
  duvidas: "Dúvidas",
  reclamacao: "Reclamação",
  outro: "Outro",
};

export const INTENTS: readonly Intent[] = [
  "agendar",
  "precos",
  "duvidas",
  "reclamacao",
  "outro",
];

export const CHANNEL_LABELS: Record<ChannelType, string> = {
  whatsapp: "WhatsApp",
  instagram: "Instagram",
  facebook: "Messenger",
};

export const CHANNELS: readonly ChannelType[] = [
  "whatsapp",
  "instagram",
  "facebook",
];

export const ENGAGEMENT_TYPE_LABELS: Record<EngagementType, string> = {
  story_reply: "Resposta a story",
  story_mention: "Menção em story",
  post_comment: "Comentário em publicação",
  live_comment: "Comentário em live",
  private_reply: "Resposta privada",
};

export const ENGAGEMENT_STATUS_LABELS: Record<EngagementStatus, string> = {
  open: "Em aberto",
  replied: "Respondido",
  dismissed: "Descartado",
  converted_to_conversation: "Convertido em conversa",
};

export const MESSAGE_KIND_LABELS: Record<MessageKind, string> = {
  text: "Texto",
  template: "Modelo",
  image: "Imagem",
  audio: "Áudio",
  story_reply: "Resposta a story",
  comment_reply: "Resposta a comentário",
  system: "Sistema",
};

export const CUSTOMER_IDENTIFICATION_LABELS: Record<
  CustomerIdentification,
  string
> = {
  anonymous: "Sem telefone",
  identified: "Identificado",
};

export const PHONE_GATE_LABELS: Record<PhoneGateStatus, string> = {
  not_needed: "Telefone não necessário",
  required: "Aguardando telefone",
  pending_verification: "Confirmando no WhatsApp",
  collected: "Telefone confirmado",
};
