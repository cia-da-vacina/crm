import type { MetaSettings, UpdateMetaSettingsPayload } from "@/domain";
import { bffGet, bffPut } from "./http";

/** GET /api/proxy/settings/meta */
export async function getSettings(): Promise<MetaSettings> {
  return bffGet<MetaSettings>("/api/proxy/settings/meta");
}

/** PUT /api/proxy/settings/meta */
export async function updateSettings(
  payload: UpdateMetaSettingsPayload,
): Promise<MetaSettings> {
  return bffPut<MetaSettings>("/api/proxy/settings/meta", payload);
}

export const metaService = {
  getSettings,
  updateSettings,
};
