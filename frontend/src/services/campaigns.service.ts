import type { AICampaign } from "@/domain";
import { metaService } from "./meta.service";

/** List campaigns from Meta settings (`ai_campaigns`). */
export async function listCampaigns(): Promise<AICampaign[]> {
  const settings = await metaService.getSettings();
  return settings.ai_campaigns ?? [];
}

/**
 * Replace the full campaign list via PUT settings/meta.
 * Preserves other Meta settings fields (only `ai_campaigns` is patched).
 */
export async function saveCampaigns(
  campaigns: AICampaign[],
): Promise<AICampaign[]> {
  const updated = await metaService.updateSettings({ ai_campaigns: campaigns });
  return updated.ai_campaigns ?? [];
}

export const campaignsService = {
  list: listCampaigns,
  save: saveCampaigns,
};
