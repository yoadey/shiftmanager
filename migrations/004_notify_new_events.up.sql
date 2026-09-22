-- 004_notify_new_events.up.sql
-- Adds the "new event published" opt-in preference (KANN) and its template.

-- Opt-in flag: member receives an e-mail when a new event is published.
ALTER TABLE members ADD COLUMN IF NOT EXISTS notify_new_events BOOLEAN NOT NULL DEFAULT false;

INSERT INTO email_templates (name, subject, body) VALUES
    (
        'new-event',
        'Neue Veranstaltung: {{.EventName}}',
        E'Hallo,\n\nes gibt eine neue Veranstaltung: "{{.EventName}}".\nOrt: {{.Location}}\nBeginn: {{.StartAt}}\n\nSchau vorbei und melde dich fuer eine Schicht an!\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen'
    )
ON CONFLICT (name) DO UPDATE
    SET subject = EXCLUDED.subject, body = EXCLUDED.body;
