CREATE TABLE messages (
    id               UUID PRIMARY KEY,
    conversation_id  UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    direction        TEXT NOT NULL CHECK (direction IN ('in', 'out')),
    sender_type      TEXT NOT NULL CHECK (sender_type IN ('contact', 'agent', 'ai', 'system')),
    sender_user_id   UUID REFERENCES users(id),
    kind             TEXT NOT NULL DEFAULT 'text'
        CHECK (kind IN ('text', 'image', 'document', 'audio', 'video', 'template', 'system')),
    channel          TEXT NOT NULL CHECK (channel IN ('whatsapp', 'instagram', 'facebook')),
    body             TEXT NOT NULL DEFAULT '',
    status           TEXT NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'delivered', 'read', 'failed')),
    meta_message_id  TEXT,
    media_url        TEXT,
    media_mime_type  TEXT,
    template_name    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
    -- reply_to_engagement_id entra na fase 8 (Engagements), quando
    -- social_engagements existir — ver docs/BACKEND-CONTRACT.md §4.
);

CREATE INDEX messages_conversation_id_created_at_idx ON messages(conversation_id, created_at DESC);
CREATE UNIQUE INDEX messages_meta_message_id_unique_idx ON messages(meta_message_id) WHERE meta_message_id IS NOT NULL;
