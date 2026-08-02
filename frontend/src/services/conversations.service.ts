import type {
  ConversationDetail,
  CursorPage,
  Message,
  MessageKind,
  PipelineStage,
  TriageSummary,
} from "@/domain";
import { bffGet, bffPatch, bffPost, toQueryString } from "./http";

export interface ListMessagesParams {
  cursor?: string;
  limit?: number;
}

export interface SendMessagePayload {
  body: string;
  kind?: MessageKind;
}

export interface UpdatePipelinePayload {
  stage: PipelineStage;
  reason_code?: string;
  reason_text?: string;
}

/** GET /api/proxy/conversations/:id */
export async function get(id: string): Promise<ConversationDetail> {
  return bffGet<ConversationDetail>(`/api/proxy/conversations/${id}`);
}

/** GET /api/proxy/conversations/:id/messages */
export async function listMessages(
  id: string,
  params: ListMessagesParams = {},
): Promise<CursorPage<Message>> {
  return bffGet<CursorPage<Message>>(
    `/api/proxy/conversations/${id}/messages${toQueryString(params)}`,
  );
}

/** POST /api/proxy/conversations/:id/messages */
export async function sendMessage(
  id: string,
  payload: SendMessagePayload,
): Promise<Message> {
  return bffPost<Message>(`/api/proxy/conversations/${id}/messages`, payload);
}

/** POST /api/proxy/conversations/:id/claim */
export async function claim(id: string): Promise<ConversationDetail> {
  return bffPost<ConversationDetail>(`/api/proxy/conversations/${id}/claim`);
}

/** PATCH /api/proxy/conversations/:id/pipeline */
export async function updatePipeline(
  id: string,
  payload: UpdatePipelinePayload,
): Promise<ConversationDetail> {
  return bffPatch<ConversationDetail>(
    `/api/proxy/conversations/${id}/pipeline`,
    payload,
  );
}

/** GET /api/proxy/conversations/:id/triage */
export async function getTriage(id: string): Promise<TriageSummary> {
  return bffGet<TriageSummary>(`/api/proxy/conversations/${id}/triage`);
}

/** POST /api/proxy/conversations/:id/phone — start WhatsApp OTP verification */
export async function attachPhone(
  id: string,
  phone_e164: string,
): Promise<ConversationDetail> {
  return bffPost<ConversationDetail>(`/api/proxy/conversations/${id}/phone`, {
    phone_e164,
  });
}

/** POST /api/proxy/conversations/:id/phone/confirm — confirm OTP code */
export async function confirmPhone(
  id: string,
  code: string,
): Promise<ConversationDetail> {
  return bffPost<ConversationDetail>(
    `/api/proxy/conversations/${id}/phone/confirm`,
    { code },
  );
}

/** POST /api/proxy/conversations/:id/phone/resend — resend WhatsApp OTP */
export async function resendPhoneVerification(
  id: string,
): Promise<ConversationDetail> {
  return bffPost<ConversationDetail>(
    `/api/proxy/conversations/${id}/phone/resend`,
  );
}

export const conversationsService = {
  get,
  listMessages,
  sendMessage,
  claim,
  updatePipeline,
  getTriage,
  attachPhone,
  confirmPhone,
  resendPhoneVerification,
};
