-- 009_event_header_image.up.sql
-- V-010: an event can have a single dedicated header image, shown at the top
-- of the event detail page, separate from the general attachments list
-- (event_attachments, see 005_event_attachments).

ALTER TABLE events ADD COLUMN IF NOT EXISTS header_image_url TEXT;
