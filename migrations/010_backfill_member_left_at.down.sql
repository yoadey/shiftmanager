-- 010_backfill_member_left_at.down.sql
-- Data-only migration: the backfilled leave dates cannot be told apart from
-- genuine ones, so there is nothing to undo.
SELECT 1;
