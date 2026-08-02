import type { CursorPage, FollowUp, FollowUpStatus, PipelineStage } from "@/domain";
import { bffGet, bffPost, toQueryString } from "./http";

export interface ListFollowUpsParams {
  unit_id?: string;
  status?: FollowUpStatus;
  stage?: PipelineStage;
  cursor?: string;
  limit?: number;
}

/** GET /api/proxy/followups */
export async function list(
  params: ListFollowUpsParams = {},
): Promise<CursorPage<FollowUp>> {
  return bffGet<CursorPage<FollowUp>>(`/api/proxy/followups${toQueryString(params)}`);
}

/** POST /api/proxy/followups/:id/complete */
export async function complete(id: string): Promise<FollowUp> {
  return bffPost<FollowUp>(`/api/proxy/followups/${id}/complete`);
}

/** POST /api/proxy/followups/:id/cancel */
export async function cancel(id: string): Promise<FollowUp> {
  return bffPost<FollowUp>(`/api/proxy/followups/${id}/cancel`);
}

export const followupsService = {
  list,
  complete,
  cancel,
};
