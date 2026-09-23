-- 007_club_year_carry_over.up.sql
-- S-006: carrying over excess hours into the following club year is
-- optionally configurable per club year.

ALTER TABLE club_years ADD COLUMN IF NOT EXISTS carry_over_enabled BOOLEAN NOT NULL DEFAULT false;
