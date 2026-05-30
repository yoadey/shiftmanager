-- 002_seed.up.sql
-- Seeds the initial club year 2026, branding, application settings, fee tiers
-- and the default email templates.

-- Club year 2026 (fixed UUID so dependent rows can reference it deterministically).
INSERT INTO club_years (id, label, start_date, end_date, default_target_hours, is_active)
VALUES (
    '00000000-0000-0000-0000-000000002026',
    '2026',
    '2026-01-01T00:00:00Z',
    '2026-12-31T23:59:59Z',
    10,
    true
)
ON CONFLICT (id) DO NOTHING;

-- Branding: TSC Schwarz-Gelb Aachen with the club accent colour.
INSERT INTO branding_config (id, club_name, primary_color, accent_color, logo_url)
VALUES (1, 'TSC Schwarz-Gelb Aachen', '#000000', '#F4B63F', '')
ON CONFLICT (id) DO UPDATE
SET club_name = EXCLUDED.club_name,
    primary_color = EXCLUDED.primary_color,
    accent_color = EXCLUDED.accent_color;

-- Default application settings.
INSERT INTO app_settings (key, value) VALUES
    ('nameMode', 'abbrev'),
    ('kioskSearch', 'false'),
    ('reservationHours', '48'),
    ('deregisterDeadlineH', '24'),
    ('billingMode', 'manuell')
ON CONFLICT (key) DO NOTHING;

-- Fee tiers for club year 2026 (amounts in euro cents per missing hour).
INSERT INTO fee_tiers (club_year_id, position, amount_cents) VALUES
    ('00000000-0000-0000-0000-000000002026', 1, 500),
    ('00000000-0000-0000-0000-000000002026', 2, 700),
    ('00000000-0000-0000-0000-000000002026', 3, 1000),
    ('00000000-0000-0000-0000-000000002026', 4, 1500)
ON CONFLICT (club_year_id, position) DO NOTHING;

-- Default email templates.
INSERT INTO email_templates (name, subject, body) VALUES
    (
        'kiosk-confirmation',
        'Bitte bestaetige deine Anmeldung',
        E'Hallo,\n\nvielen Dank fuer deine Anmeldung. Bitte bestaetige sie ueber den folgenden Link:\n{{.ConfirmURL}}\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'shift-confirmation',
        'Anmeldung bestaetigt',
        E'Hallo,\n\ndeine Anmeldung zur Schicht "{{.Shift.Name}}" ist bestaetigt.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'shift-cancellation',
        'Anmeldung storniert',
        E'Hallo,\n\ndeine Anmeldung zur Schicht "{{.Shift.Name}}" wurde storniert.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'reminder',
        'Erinnerung an deine Schicht',
        E'Hallo,\n\ndies ist eine Erinnerung an deine Schicht "{{.Shift.Name}}" am {{.Shift.StartAt}}.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    )
ON CONFLICT (name) DO NOTHING;
