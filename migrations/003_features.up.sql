-- 003_features.up.sql
-- Adds schema for: reminder opt-out, email logging, per-member fee tier
-- overrides, seeds the full standard set of email templates, and extra app
-- settings keys.

-- N-001: per-member reminder opt-out flag.
ALTER TABLE members ADD COLUMN IF NOT EXISTS reminder_opt_out BOOLEAN NOT NULL DEFAULT false;

-- N-004: failed-email log. Records every send attempt (sent or failed).
CREATE TABLE IF NOT EXISTS email_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    to_address   TEXT        NOT NULL,
    template     TEXT        NOT NULL DEFAULT '',
    subject      TEXT        NOT NULL DEFAULT '',
    body         TEXT        NOT NULL DEFAULT '',
    status       TEXT        NOT NULL DEFAULT 'sent',
    error        TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_email_log_created_at ON email_log (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_email_log_status ON email_log (status);

-- G-004: per-member fee tier overrides. When a member has rows here for a club
-- year, billing uses them instead of the club-year-wide fee_tiers list.
CREATE TABLE IF NOT EXISTS member_fee_tiers (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id    UUID    NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    club_year_id UUID    NOT NULL REFERENCES club_years (id) ON DELETE CASCADE,
    position     INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    UNIQUE (member_id, club_year_id, position)
);

CREATE INDEX IF NOT EXISTS idx_member_fee_tiers_member_year ON member_fee_tiers (member_id, club_year_id);

-- Extra application settings (N-002 reminder time/lead, K-012 kiosk lock,
-- year-end billing warning lead weeks).
INSERT INTO app_settings (key, value) VALUES
    ('reminderHourOfDay', '8'),
    ('reminderLeadWeeks', '1'),
    ('billingWarningLeadWeeks', '4'),
    ('kioskLocked', 'false')
ON CONFLICT (key) DO NOTHING;

-- Full standard set of email templates. Uses flattened placeholders rendered by
-- the email service: {{.MemberName}}, {{.ShiftName}}, {{.EventName}},
-- {{.StartAt}}, {{.EndAt}}, {{.Location}}, {{.ConfirmURL}}, {{.DaysUntil}},
-- {{.MissingHours}}, {{.YearLabel}}, {{.AmountEUR}}.
INSERT INTO email_templates (name, subject, body) VALUES
    (
        'kiosk-confirmation',
        'Bitte bestaetige deine Anmeldung: {{.EventName}}',
        E'Hallo {{.MemberName}},\n\nvielen Dank fuer deine Anmeldung zur Schicht "{{.ShiftName}}" bei der Veranstaltung "{{.EventName}}".\n\nBitte bestaetige deine Anmeldung ueber folgenden Link:\n{{.ConfirmURL}}\n\nSchichtbeginn: {{.StartAt}}\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'shift-confirmation',
        'Anmeldung bestaetigt: {{.EventName}}',
        E'Hallo {{.MemberName}},\n\ndeine Anmeldung zur Schicht "{{.ShiftName}}" bei "{{.EventName}}" ist bestaetigt.\n\nOrt: {{.Location}}\nBeginn: {{.StartAt}}\nEnde: {{.EndAt}}\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'shift-deregistration',
        'Abmeldung bestaetigt: {{.EventName}}',
        E'Hallo {{.MemberName}},\n\ndeine Abmeldung von der Schicht "{{.ShiftName}}" bei "{{.EventName}}" wurde verarbeitet.\n\nFalls dies ein Versehen war, kannst du dich erneut anmelden.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'reminder-1w',
        'Erinnerung: deine Schicht in einer Woche',
        E'Hallo {{.MemberName}},\n\ndies ist eine Erinnerung an deine Schicht "{{.ShiftName}}" bei "{{.EventName}}".\n\nDie Schicht beginnt in {{.DaysUntil}} Tag(en) am {{.StartAt}}.\nOrt: {{.Location}}\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'reminder-1d',
        'Erinnerung: deine Schicht morgen',
        E'Hallo {{.MemberName}},\n\ndies ist eine Erinnerung an deine Schicht "{{.ShiftName}}" bei "{{.EventName}}".\n\nDie Schicht beginnt in {{.DaysUntil}} Tag(en) am {{.StartAt}}.\nOrt: {{.Location}}\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'shift-cancelled',
        'Schicht abgesagt: {{.EventName}}',
        E'Hallo {{.MemberName}},\n\ndie Schicht "{{.ShiftName}}" bei "{{.EventName}}" wurde leider abgesagt.\n\nWir bitten um dein Verstaendnis.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'year-billing',
        'Beitragsabrechnung Vereinsjahr {{.YearLabel}}',
        E'Hallo {{.MemberName}},\n\nim Vereinsjahr "{{.YearLabel}}" hast du {{.MissingHours}} Stunden nicht erbracht.\nDir wird ein Beitrag von {{.AmountEUR}} EUR berechnet.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'missing-hours-warning',
        'Offene Stunden im Vereinsjahr {{.YearLabel}}',
        E'Hallo {{.MemberName}},\n\ndu hast im Vereinsjahr "{{.YearLabel}}" noch {{.MissingHours}} offene Stunden.\n\nBitte melde dich rechtzeitig fuer weitere Schichten an, um deine Stunden zu erfuellen.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    ),
    (
        'shift-understaffed',
        'Schicht unterbesetzt: {{.EventName}}',
        E'Hallo,\n\ndie Schicht "{{.ShiftName}}" bei "{{.EventName}}" am {{.StartAt}} ist unterbesetzt.\nEs werden noch Helfer benoetigt.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    )
ON CONFLICT (name) DO UPDATE
    SET subject = EXCLUDED.subject, body = EXCLUDED.body;
