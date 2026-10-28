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
UPDATE members
SET left_at = CURRENT_TIMESTAMP
WHERE is_active = false
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
