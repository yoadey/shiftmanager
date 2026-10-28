-- 010_backfill_member_left_at.up.sql
-- ML-011: "waiting for activation" means is_active = false AND left_at IS NULL.
-- Before this, deactivating a member via PUT /members/{id} {isActive:false}
-- did not record a leave date, so such members would now look like pending
-- self-registrations — and an OIDC role expression (A-006) could re-admit
-- them on their next login.
--
-- Self-registrations are always created together with a 'member.autoregister'
-- audit entry. Keep only members whose audit history consists solely of that
-- entry as pending; every other inactive member without a leave date counts
-- as deactivated. When in doubt, a member ends up "left" (an admin can
-- reactivate them) rather than pending (where the IdP could admit them).
--
-- The same backfill also runs from the app's own GORM migration
-- (internal/adapter/db/data_migrations.go). Both record it in
-- data_migrations and skip it when already recorded, so it runs exactly once
-- per database whichever path gets there first: a second run could flip a
-- pending registration that an admin edited in the meantime.
CREATE TABLE IF NOT EXISTS data_migrations (
    name       TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ
);

UPDATE members
SET left_at = CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM data_migrations WHERE name = '010_backfill_member_left_at')
  AND is_active = false
  AND left_at IS NULL
  AND NOT (
    EXISTS (
      SELECT 1 FROM audit_log a
      WHERE a.entity = 'member'
        AND a.entity_id = CAST(members.id AS TEXT)
        AND a.action = 'member.autoregister'
    )
    AND NOT EXISTS (
      SELECT 1 FROM audit_log a
      WHERE a.entity = 'member'
        AND a.entity_id = CAST(members.id AS TEXT)
        AND a.action <> 'member.autoregister'
    )
  );

INSERT INTO data_migrations (name, applied_at)
VALUES ('010_backfill_member_left_at', CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;
