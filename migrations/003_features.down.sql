-- 003_features.down.sql
-- Reverts 003_features.up.sql.

DELETE FROM email_templates WHERE name IN (
    'shift-deregistration', 'reminder-1w', 'reminder-1d', 'shift-cancelled',
    'year-billing', 'missing-hours-warning', 'shift-understaffed'
);

DELETE FROM app_settings WHERE key IN (
    'reminderHourOfDay', 'reminderLeadWeeks', 'billingWarningLeadWeeks', 'kioskLocked'
);

DROP TABLE IF EXISTS member_fee_tiers;
DROP TABLE IF EXISTS email_log;

ALTER TABLE members DROP COLUMN IF EXISTS reminder_opt_out;
