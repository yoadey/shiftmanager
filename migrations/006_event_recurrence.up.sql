-- 006_event_recurrence.up.sql
-- V-007: events can be configured as recurring (weekly, monthly).

ALTER TABLE events ADD COLUMN IF NOT EXISTS recurrence_frequency TEXT NOT NULL DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS recurrence_until TIMESTAMPTZ;
ALTER TABLE events ADD COLUMN IF NOT EXISTS recurrence_group_id UUID;

CREATE INDEX IF NOT EXISTS idx_events_recurrence_group_id ON events (recurrence_group_id);
