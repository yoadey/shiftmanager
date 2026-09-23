-- 007_club_year_carry_over.down.sql
-- Reverts 007_club_year_carry_over.up.sql.

ALTER TABLE club_years DROP COLUMN IF EXISTS carry_over_enabled;
