-- 008_registration_guest_name.down.sql
-- Reverts 008_registration_guest_name.up.sql.

ALTER TABLE registrations DROP COLUMN IF EXISTS guest_name;
