import type { ChannelType, EngagementStatus, EngagementType } from "./enums";

/**
 * A Meta interaction outside the normal 1:1 messaging thread: a reply to
 * one of our Instagram stories, a mention in a story, a comment on a post
 * or live broadcast, or a private reply we sent to a commenter.
 *
 * These arrive via distinct webhook payloads (messages webhook with
 * `reply_to.story` for story replies, `messaging_referrals` for story
 * mentions, `comments` webhook for post/live comments) and are triaged in
 * their own queue before optionally being converted into a full
 * `ConversationDetail` once a real back-and-forth thread exists.
 */
export interface SocialEngagement {
  id: string;
  /** Known once the author has been matched/linked to a CRM customer. */
  customer_id?: string | null;
  customer_name?: string | null;
  channel: ChannelType;
  type: EngagementType;
  status: EngagementStatus;
  unit_id: string;
  /** Meta media id of the story/post/live the engagement is attached to. */
  media_id?: string | null;
  media_url?: string | null;
  media_caption?: string | null;
  /** Text content of the story reply / comment. */
  body: string;
  /** Meta-side id of the engagement itself (comment id, story reply message id). */
  external_id: string;
  /** IGSID/PSID of the person who authored the engagement. */
  author_external_id: string;
  /** Populated once this engagement was answered and promoted into a real conversation thread. */
  conversation_id?: string | null;
  created_at: string;
  replied_at?: string | null;
}
