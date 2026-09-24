-- 008_registration_guest_name.up.sql
-- An organizer can add a non-member helper to a shift by name (optionally
-- with an email), distinct from the existing self-service kiosk guest
-- registration, which is tracked by email only (guest_email).

ALTER TABLE registrations ADD COLUMN IF NOT EXISTS guest_name TEXT;
