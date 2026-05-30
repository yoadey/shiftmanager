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

// ── Tweaks ────────────────────────────────────────────────────────────────────

export interface Tweaks {
  primaryColor: string;
  radius: 'klein' | 'standard' | 'weich';
  warmth: 'warm' | 'neutral';
}
