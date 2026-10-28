-- 009_event_header_image.down.sql
-- Reverts 009_event_header_image.up.sql.

ALTER TABLE events DROP COLUMN IF EXISTS header_image_url;
