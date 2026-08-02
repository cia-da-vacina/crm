import type {
  ChannelType,
  MessageDirection,
  MessageKind,
  MessageStatus,
  SenderType,
} from "./enums";

export interface Message {
  id: string;
  conversation_id: string;
  direction: MessageDirection;
  sender_type: SenderType;
  kind: MessageKind;
  channel: ChannelType;
  body: string;
  status: MessageStatus;
  /** Meta's own message id (`wamid.*` for WA, `mid.*` for IG/FB), for delivery/read reconciliation. */
  meta_message_id?: string | null;
  /** Set when this message was sent in reply to a story reply/mention or comment engagement. */
  reply_to_engagement_id?: string | null;
  /** Populated when kind is `image` or `audio`. */
  media_url?: string | null;
  media_mime_type?: string | null;
  /** Populated when kind is `template` (Meta message templates require a registered name). */
  template_name?: string | null;
  created_at: string;
}
