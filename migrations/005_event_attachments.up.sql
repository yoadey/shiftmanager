-- 005_event_attachments.up.sql
-- V-008: events can be furnished with images and other attachments.

CREATE TABLE IF NOT EXISTS event_attachments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id     UUID        NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    file_name    TEXT        NOT NULL,
    url          TEXT        NOT NULL,
    content_type TEXT        NOT NULL DEFAULT '',
    size_bytes   BIGINT      NOT NULL DEFAULT 0,
    uploaded_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_event_attachments_event_id ON event_attachments (event_id);
