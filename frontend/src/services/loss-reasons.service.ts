import type { LossReason } from "@/domain";
import { bffGet } from "./http";

/** GET /api/proxy/loss-reasons */
export async function list(): Promise<LossReason[]> {
  return bffGet<LossReason[]>("/api/proxy/loss-reasons");
}

export const lossReasonsService = {
  list,
};
