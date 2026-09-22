-- 001_init.down.sql
-- Drops all objects created in 001_init.up.sql in reverse dependency order.

DROP TABLE IF EXISTS email_templates;
DROP TABLE IF EXISTS app_settings;
DROP TABLE IF EXISTS branding_config;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS hour_entries;
DROP TABLE IF EXISTS registrations;
DROP TABLE IF EXISTS shifts;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS fee_tiers;
DROP TABLE IF EXISTS hour_targets;
DROP TABLE IF EXISTS oidc_links;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS club_years;
