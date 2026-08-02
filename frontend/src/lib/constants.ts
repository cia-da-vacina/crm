export const APP_NAME = "CRM Cia da Vacina";

export const APP_DESCRIPTION =
  "Atendimento multicanal (WhatsApp, Instagram e Messenger), triagem com IA e pipeline comercial das unidades.";

/**
 * sessionStorage keys. These are UI-only preferences (never auth state or
 * PII) so `sessionStorage` — cleared per tab — is an acceptable and
 * intentional choice here.
 */
export const STORAGE = {
  PWA_INSTALL_DISMISSED: "pwa-install-dismissed",
  /** Persisted appearance preference across sessions. */
  THEME_MODE: "cv-theme-mode",
} as const;

/** Shared React Query `staleTime` presets, in milliseconds. */
export const QUERY_STALE = {
  /** Near-real-time data: inbox list, conversation messages, engagements queue. */
  REALTIME: 15_000,
  /** Default for most reads: pipeline, follow-ups, dashboard. */
  DEFAULT: 60_000,
  /** Rarely-changing reference data: units, POPs, meta settings. */
  LONG: 5 * 60_000,
} as const;
