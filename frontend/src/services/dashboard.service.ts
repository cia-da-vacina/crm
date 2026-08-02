import type { DashboardSummary } from "@/domain";
import { bffGet, toQueryString } from "./http";

export interface DashboardSummaryParams {
  unit_id?: string;
}

/** GET /api/proxy/dashboard/summary */
export async function getSummary(
  params: DashboardSummaryParams = {},
): Promise<DashboardSummary> {
  return bffGet<DashboardSummary>(`/api/proxy/dashboard/summary${toQueryString(params)}`);
}

export const dashboardService = {
  getSummary,
};
