import type { ChannelType, Intent } from "./enums";

/**
 * Configuration for a single Meta channel connection. This type is safe to
 * send to the client: it never carries the raw access token, only a masked
 * preview (`token_masked`) for display purposes.
 */
export interface MetaChannelConfig {
  channel: ChannelType;
  enabled: boolean;
  /** WABA id (whatsapp), IG business account id (instagram) or Page id (facebook). */
  account_id: string;
  display_name: string;
  /** WhatsApp-only: the phone number id used to send/receive messages. */
  phone_number_id?: string | null;
  webhook_verified: boolean;
  /** Never the full token — e.g. `EAAG...9f2a`. */
  token_masked: string;
}

export interface AICampaign {
  id: string;
  title: string;
  description: string;
  starts_on: string;
  ends_on: string;
  active: boolean;
}

export interface MetaSettings {
  channels: MetaChannelConfig[];
  ai_enabled: boolean;
  ai_system_prompt: string;
  ai_context: string;
  ai_campaigns: AICampaign[];
  triage_enabled: boolean;
  /** Intents that trigger automatic handoff from AI triage to a human agent. */
  triage_handoff_intents: Intent[];
}

/**
 * Payload for `PUT/PATCH` settings updates. `channel_tokens` is the only way
 * to rotate a channel's access token: raw values are forwarded straight
 * through to the BFF route handler and on to the backend, and must never be
 * echoed back into client state (the response should always be re-fetched
 * as a fresh `MetaSettings`, which only ever contains `token_masked`).
 */
export type UpdateMetaSettingsPayload = Partial<
  Omit<MetaSettings, "channels">
> & {
  channels?: Array<
    Partial<Omit<MetaChannelConfig, "token_masked">> & {
      channel: ChannelType;
    }
  >;
  /** Raw new tokens, keyed by channel. Write-only — never stored client-side after save. */
  channel_tokens?: Partial<Record<ChannelType, string>>;
};
