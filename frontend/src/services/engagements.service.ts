import type {
  ChannelType,
  ConversationDetail,
  CursorPage,
  EngagementStatus,
  EngagementType,
  SocialEngagement,
} from "@/domain";
import { bffGet, bffPost, toQueryString } from "./http";

export interface ListEngagementsParams {
  unit_id?: string;
  channel?: ChannelType;
  type?: EngagementType;
  status?: EngagementStatus;
  cursor?: string;
  limit?: number;
}

/** GET /api/proxy/engagements */
export async function list(
  params: ListEngagementsParams = {},
): Promise<CursorPage<SocialEngagement>> {
  return bffGet<CursorPage<SocialEngagement>>(
    `/api/proxy/engagements${toQueryString(params)}`,
  );
}

/** GET /api/proxy/engagements/:id */
export async function get(id: string): Promise<SocialEngagement> {
  return bffGet<SocialEngagement>(`/api/proxy/engagements/${id}`);
}

/** POST /api/proxy/engagements/:id/reply */
export async function reply(id: string, body: string): Promise<SocialEngagement> {
  return bffPost<SocialEngagement>(`/api/proxy/engagements/${id}/reply`, { body });
}

/** POST /api/proxy/engagements/:id/dismiss */
export async function dismiss(id: string): Promise<SocialEngagement> {
  return bffPost<SocialEngagement>(`/api/proxy/engagements/${id}/dismiss`);
}

/** POST /api/proxy/engagements/:id/convert — promotes an engagement into a full conversation thread. */
export async function convertToConversation(id: string): Promise<ConversationDetail> {
  return bffPost<ConversationDetail>(`/api/proxy/engagements/${id}/convert`);
}

export const engagementsService = {
  list,
  get,
  reply,
  dismiss,
  convertToConversation,
};
