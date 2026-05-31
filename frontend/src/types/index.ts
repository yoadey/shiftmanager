// ── Enums / literals ──────────────────────────────────────────────────────────

export type RegistrationState = 'registered' | 'reserved' | 'confirmed' | 'no_show';

export type EventStatus = 'draft' | 'published' | 'completed' | 'cancelled'
  // German variants used in prototype data
  | 'entwurf' | 'veröffentlicht' | 'abgeschlossen' | 'abgesagt';

export type UserRole = 'admin' | 'vorstand' | 'veranstaltungsleiter' | 'mitglied';

// ── Domain models ─────────────────────────────────────────────────────────────

export interface Member {
  id: string;
  first: string;
  last: string;
  email: string;
  since: string; // ISO date
  goal: number | null;
  active?: boolean;
  // N-001: when true, the member has opted out of (non-mandatory) reminder mails.
  reminderOptOut?: boolean;
}

export interface Signup {
  memberId: string;
  status: 'angemeldet' | 'bestätigt' | 'reserviert' | 'nichterschienen';
  comment?: string;
  hours?: number;
  guest?: string;
}

export interface Shift {
  id: string;
  name: string;
  start: string; // HH:MM
  end: string;   // HH:MM
  min: number;
  max: number;
  qual?: string;
  desc?: string;
  signups: Signup[];
}

export interface ShiftDay {
  date: string; // ISO date
  shifts: Shift[];
}

export interface Event {
  id: string;
  name: string;
  category: string;
  location: string;
  status: EventStatus;
  description: string;
  days: ShiftDay[];
}

export interface Registration {
  id: string;
  shiftId: string;
  memberId: string;
  state: RegistrationState;
  comment?: string;
  hours?: number;
  createdAt: string;
}

export interface HourEntry {
  id: string;
  memberId: string;
  date: string;
  hours: number;
  desc: string;
  eventName?: string;
  shiftName?: string;
  manual?: boolean;
  by?: string;
}

export interface ClubYear {
  year: string;
  goal: number;
}

export interface AuditEntry {
  ts: string;
  who: string;
  what: string;
  cat: string;
}

// ── App settings ──────────────────────────────────────────────────────────────

export interface AppSettings {
  clubName: string;
  yearGoal: number;
  clubYear: string;
  nameMode: 'abbrev' | 'full';
  reservationHours: number;
  billingMode: 'auto' | 'manuell';
  kioskSearch: boolean;
  feeSchedule: number[];
  deregisterDeadlineH: number;
  // K-012: when true the public kiosk routes are disabled.
  kioskLocked?: boolean;
}

export interface BrandingConfig {
  primaryColor: string;
  logoUrl?: string;
  clubName: string;
}

export interface FeeTier {
  missingHour: number;
  amount: number;
}

// ── Computed / view types ─────────────────────────────────────────────────────

export interface ShiftOccupancy {
  count: number;
  min: number;
  max: number;
  key: 'ok' | 'warn' | 'crit' | 'full';
  label: string;
  free: number;
  needsMore: boolean;
}

export type ShiftWithRegistrations = Shift;

export interface EventTimeline {
  event: Event;
  days: { date: string; shifts: ShiftWithRegistrations[] }[];
}

export interface MemberAccount {
  confirmed: number;
  reserved: number;
  goal: number;
}

// ── Auth ──────────────────────────────────────────────────────────────────────

export interface AuthUser {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  first?: string;
  last?: string;
}

// ── Admin statistics (D-004) ───────────────────────────────────────────────────

export interface SystemStats {
  clubYearId: string;
  clubYearLabel: string;
  totalConfirmedHours: number;
  openShifts: number;
  upcomingShifts: number;
  activeMembers: number;
  membersBelowTarget: number;
}

// ── Branding update result (B-003) ──────────────────────────────────────────────

export interface BrandingUpdateResult {
  branding: BrandingConfig;
  warnings?: string[];
}

// ── Email templates / log (Section 4, N-004) ───────────────────────────────────

export interface EmailTemplate {
  id: string;
  name: string;
  subject: string;
  body: string;
}

export type EmailLogStatus = 'sent' | 'failed';

export interface EmailLogEntry {
  id: string;
  to: string;
  template: string;
  subject: string;
  body: string;
  status: EmailLogStatus;
  error: string;
  createdAt: string;
}

// ── Per-member fee-tier overrides (G-004) ───────────────────────────────────────
// Mirrors the backend domain.FeeTier struct (camelCase JSON).

export interface MemberFeeTier {
  id: string;
  clubYearId: string;
  position: number;
  amountCents: number;
}

// ── GDPR data export (DS-003) ───────────────────────────────────────────────────
// The export is an opaque document; we download it as JSON rather than render it.

export interface MemberDataExport {
  member: unknown;
  registrations: unknown[];
  hourEntries: unknown[];
  hourTargets: unknown[];
  exportedAt: string;
}

// ── Tweaks ────────────────────────────────────────────────────────────────────

export interface Tweaks {
  primaryColor: string;
  radius: 'klein' | 'standard' | 'weich';
  warmth: 'warm' | 'neutral';
}
