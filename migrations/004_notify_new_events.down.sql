-- 004_notify_new_events.down.sql
-- Reverts 004_notify_new_events.up.sql.

DELETE FROM email_templates WHERE name = 'new-event';

ALTER TABLE members DROP COLUMN IF EXISTS notify_new_events;
