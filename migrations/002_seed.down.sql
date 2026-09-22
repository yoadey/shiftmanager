-- 002_seed.down.sql
-- Removes the seed data inserted in 002_seed.up.sql.

DELETE FROM email_templates
WHERE name IN ('kiosk-confirmation', 'shift-confirmation', 'shift-cancellation', 'reminder');

DELETE FROM fee_tiers
WHERE club_year_id = '00000000-0000-0000-0000-000000002026';

DELETE FROM app_settings
WHERE key IN ('nameMode', 'kioskSearch', 'reservationHours', 'deregisterDeadlineH', 'billingMode');

DELETE FROM branding_config WHERE id = 1;

DELETE FROM club_years
WHERE id = '00000000-0000-0000-0000-000000002026';
