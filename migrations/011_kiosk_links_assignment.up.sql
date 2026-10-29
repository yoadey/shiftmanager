-- 011_kiosk_links_assignment.up.sql
-- Kiosk entries by free-text name or e-mail that an organiser assigns to a
-- member (K-014, K-015, K-006), and long-lived kiosk links (K-016).

ALTER TABLE registrations ADD COLUMN IF NOT EXISTS self_registered BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE registrations ADD COLUMN IF NOT EXISTS pending_member_id UUID REFERENCES members (id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS kiosk_links (
    id         UUID PRIMARY KEY,
    label      TEXT        NOT NULL DEFAULT '',
    token_hash TEXT        NOT NULL UNIQUE,
    created_by UUID        REFERENCES members (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

INSERT INTO email_templates (name, subject, body) VALUES
    (
        'kiosk-assigned',
        'Du bist eingetragen: {{.EventName}}',
        E'Hallo {{.MemberName}},\n\ndeine Eintragung am Kiosk wurde deinem Konto zugeordnet: Schicht "{{.ShiftName}}" bei "{{.EventName}}".\nOrt: {{.Location}}\nBeginn: {{.StartAt}}\n\nFalls das nicht du warst, melde dich bitte beim Vorstand.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    )
ON CONFLICT (name) DO NOTHING;
