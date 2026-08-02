import type {
  ChannelType,
  ConversationMode,
  CursorPage,
  InboxItem,
  PipelineStage,
} from "@/domain";
import { bffGet, toQueryString } from "./http";

export interface ListInboxParams {
  unit_id?: string;
  stage?: PipelineStage;
  channel?: ChannelType;
  mode?: ConversationMode;
  cursor?: string;
  limit?: number;
}

/** GET /api/proxy/inbox — proxied to backend `GET /inbox`. */
export async function listInbox(
  params: ListInboxParams = {},
): Promise<CursorPage<InboxItem>> {
  return bffGet<CursorPage<InboxItem>>(`/api/proxy/inbox${toQueryString(params)}`);
}

export const inboxService = {
  listInbox,
};
