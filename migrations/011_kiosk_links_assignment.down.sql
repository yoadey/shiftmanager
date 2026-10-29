-- 011_kiosk_links_assignment.down.sql
DELETE FROM email_templates WHERE name = 'kiosk-assigned';
DROP TABLE IF EXISTS kiosk_links;
ALTER TABLE registrations DROP COLUMN IF EXISTS pending_member_id;
ALTER TABLE registrations DROP COLUMN IF EXISTS self_registered;
