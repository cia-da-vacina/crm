-- Engagements Meta-nativos (story reply/mention, comentário de post/live) —
-- fora do fluxo normal de mensagens 1:1 (docs/BACKEND-CONTRACT.md §5). Só
-- Instagram/Facebook têm isso; WhatsApp não tem posts/stories/comments.
CREATE TABLE social_engagements (
    id                 UUID PRIMARY KEY,
    customer_id        UUID REFERENCES customers(id) ON DELETE SET NULL,
    channel            TEXT NOT NULL CHECK (channel IN ('instagram', 'facebook')),
    type               TEXT NOT NULL CHECK (type IN ('story_reply', 'story_mention', 'post_comment', 'live_comment', 'private_reply')),
    status             TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'replied', 'dismissed', 'converted_to_conversation')),
    unit_id            UUID NOT NULL REFERENCES units(id),
    media_id           TEXT,
    media_url          TEXT,
    media_caption      TEXT,
    body               TEXT NOT NULL DEFAULT '',
    external_id        TEXT NOT NULL,
    author_external_id TEXT NOT NULL,
    conversation_id    UUID REFERENCES conversations(id) ON DELETE SET NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    replied_at         TIMESTAMPTZ
);

-- Idempotência de ingestão via webhook, mesmo princípio de messages.meta_message_id.
CREATE UNIQUE INDEX social_engagements_channel_external_id_unique_idx ON social_engagements(channel, external_id);
CREATE INDEX social_engagements_unit_id_idx ON social_engagements(unit_id);
CREATE INDEX social_engagements_created_at_id_idx ON social_engagements(created_at DESC, id DESC);

-- Adiado da migration 000010 de propósito: só faz sentido depois que
-- social_engagements existe (docs/BACKEND-CONTRACT.md §4, campo
-- reply_to_engagement_id do Message).
ALTER TABLE messages ADD COLUMN reply_to_engagement_id UUID REFERENCES social_engagements(id);
