-- 001_init.up.sql
-- Initial ShiftManager schema.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Club / financial years.
CREATE TABLE club_years (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label                TEXT        NOT NULL,
    start_date           TIMESTAMPTZ NOT NULL,
    end_date             TIMESTAMPTZ NOT NULL,
    default_target_hours DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_active            BOOLEAN     NOT NULL DEFAULT false
);

-- Members.
CREATE TABLE members (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name            TEXT        NOT NULL,
    last_name             TEXT        NOT NULL,
    email                 TEXT        NOT NULL,
    joined_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    left_at               TIMESTAMPTZ,
    is_active             BOOLEAN     NOT NULL DEFAULT true,
    individual_goal_hours DOUBLE PRECISION,
    oidc_subject          TEXT,
    role                  TEXT        NOT NULL DEFAULT 'mitglied'
);

CREATE UNIQUE INDEX idx_members_email ON members (email);
CREATE INDEX idx_members_active ON members (is_active);

-- OIDC provider links.
CREATE TABLE oidc_links (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id UUID        NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    provider  TEXT        NOT NULL,
    subject   TEXT        NOT NULL,
    linked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, subject)
);

CREATE INDEX idx_oidc_links_member ON oidc_links (member_id);

-- Per-member hour targets that override the club-year default.
CREATE TABLE hour_targets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id    UUID        NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    club_year_id UUID        NOT NULL REFERENCES club_years (id) ON DELETE CASCADE,
    target_hours DOUBLE PRECISION NOT NULL DEFAULT 0,
    UNIQUE (member_id, club_year_id)
);

-- Fee tiers (Abgeltungsbetrag) per club year.
CREATE TABLE fee_tiers (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_year_id UUID    NOT NULL REFERENCES club_years (id) ON DELETE CASCADE,
    position     INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    UNIQUE (club_year_id, position)
);

CREATE INDEX idx_fee_tiers_year ON fee_tiers (club_year_id);

-- Events.
CREATE TABLE events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    location    TEXT        NOT NULL DEFAULT '',
    category    TEXT        NOT NULL DEFAULT '',
    start_date  TIMESTAMPTZ NOT NULL,
    end_date    TIMESTAMPTZ NOT NULL,
    status      TEXT        NOT NULL DEFAULT 'draft',
    visibility  TEXT        NOT NULL DEFAULT 'public',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_status ON events (status);
CREATE INDEX idx_events_start ON events (start_date);

-- Shifts within events.
CREATE TABLE shifts (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id               UUID        NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    name                   TEXT        NOT NULL DEFAULT '',
    start_at               TIMESTAMPTZ NOT NULL,
    end_at                 TIMESTAMPTZ NOT NULL,
    min_helpers            INTEGER     NOT NULL DEFAULT 0,
    max_helpers            INTEGER     NOT NULL DEFAULT 0,
    required_qualification TEXT        NOT NULL DEFAULT '',
    shift_date             TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_shifts_event_start ON shifts (event_id, start_at);

-- Registrations link members (or guests) to shifts.
CREATE TABLE registrations (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shift_id           UUID        NOT NULL REFERENCES shifts (id) ON DELETE CASCADE,
    member_id          UUID        REFERENCES members (id) ON DELETE CASCADE,
    guest_email        TEXT,
    state              TEXT        NOT NULL DEFAULT 'registered',
    comment            TEXT        NOT NULL DEFAULT '',
    reserved_until     TIMESTAMPTZ,
    booked_hours       DOUBLE PRECISION,
    confirmation_token UUID,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_registrations_shift ON registrations (shift_id);
CREATE INDEX idx_registrations_member ON registrations (member_id);
CREATE INDEX idx_registrations_token ON registrations (confirmation_token);

-- Hour entries (shift-derived or manual).
CREATE TABLE hour_entries (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id    UUID        NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    shift_id     UUID        REFERENCES shifts (id) ON DELETE SET NULL,
    club_year_id UUID        NOT NULL REFERENCES club_years (id) ON DELETE CASCADE,
    hours        DOUBLE PRECISION NOT NULL DEFAULT 0,
    type         TEXT        NOT NULL DEFAULT 'manual',
    status       TEXT        NOT NULL DEFAULT 'pending',
    booked_by    UUID        REFERENCES members (id) ON DELETE SET NULL,
    description  TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_hour_entries_member_year ON hour_entries (member_id, club_year_id);
CREATE INDEX idx_hour_entries_year ON hour_entries (club_year_id);

-- Append-only audit log.
CREATE TABLE audit_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id     UUID,
    action       TEXT        NOT NULL,
    entity       TEXT        NOT NULL,
    entity_id    TEXT        NOT NULL,
    before_state JSONB,
    after_state  JSONB,
    changed_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_log_changed_at ON audit_log (changed_at DESC);
CREATE INDEX idx_audit_log_entity ON audit_log (entity, entity_id);

-- Branding configuration (single row, id = 1).
CREATE TABLE branding_config (
    id            INTEGER PRIMARY KEY DEFAULT 1,
    club_name     TEXT NOT NULL DEFAULT '',
    primary_color TEXT NOT NULL DEFAULT '#000000',
    accent_color  TEXT NOT NULL DEFAULT '#F4B63F',
    logo_url      TEXT NOT NULL DEFAULT '',
    CONSTRAINT branding_singleton CHECK (id = 1)
);

-- Key/value application settings.
CREATE TABLE app_settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Email templates (overridable text/template bodies).
CREATE TABLE email_templates (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name    TEXT NOT NULL UNIQUE,
    subject TEXT NOT NULL DEFAULT '',
    body    TEXT NOT NULL DEFAULT ''
);
