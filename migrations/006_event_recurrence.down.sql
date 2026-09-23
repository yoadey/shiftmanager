-- 006_event_recurrence.down.sql
-- Reverts 006_event_recurrence.up.sql.

DROP INDEX IF EXISTS idx_events_recurrence_group_id;
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_group_id;
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_until;
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_frequency;
