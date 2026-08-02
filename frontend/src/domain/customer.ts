import type { ChannelType, CustomerIdentification } from "./enums";

/**
 * A single channel-specific identity linked to a CRM customer.
 *
 * Meta does not expose a unified ID across WhatsApp / Instagram / Facebook.
 * Cross-channel identity is owned by the CRM and keyed primarily by
 * `Customer.primary_phone` (E.164) once the privacy wall is crossed.
 */
export interface CustomerIdentity {
  id: string;
  customer_id: string;
  channel: ChannelType;
  /** Channel-native identifier (wa_id, IGSID, PSID). */
  external_id: string;
  /** @username (Instagram) or profile name (Messenger). Absent for WhatsApp. */
  display_handle?: string | null;
  /**
   * Phone known on this channel identity.
   * Always set for WhatsApp (from Meta). For IG/FB, only after the user
   * voluntarily provided it during a gated triage step.
   */
  phone_e164?: string | null;
  /** When a human confirmed this identity truly belongs to the customer. */
  verified_at?: string | null;
  created_at: string;
}

/**
 * CRM customer.
 *
 * - `anonymous`: conversing without a phone — limited data surface (privacy wall).
 * - `identified`: has `primary_phone` — full history, merge across channels, gated actions.
 *
 * WhatsApp contacts are created already `identified` (Meta supplies the number).
 * Instagram / Messenger start as `anonymous` until a phone is provided AND
 * confirmed via WhatsApp OTP (possession proof). Typing a number alone never
 * identifies or merges the customer.
 */
export interface Customer {
  id: string;
  display_name: string;
  identification: CustomerIdentification;
  /**
   * Canonical business key for matching across channels (E.164).
   * Required when `identification === "identified"`; null while anonymous.
   * Unique in the backend when present.
   */
  primary_phone: string | null;
  /** Unit this customer is primarily associated with, if known. */
  unit_id?: string | null;
  identities: CustomerIdentity[];
  created_at: string;
  updated_at: string;
}
