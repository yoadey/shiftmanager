// ── Architecture note ─────────────────────────────────────────────────────────
//
// Single source of truth:  api/openapi.yaml
// Generated API types:     src/api/generated/types.gen.ts  (DO NOT EDIT)
// Regenerate:              npm run generate  (or: make generate)
//
// The types below are used by UI components. They are either:
//   (a) re-exported directly from the generated spec (shape matches API 1:1), or
//   (b) UI-adapter types with shorter field names that the api/ mappers convert
//       from the raw generated shapes (e.g. firstName → first).
//
// At the API boundary (src/api/*.ts) always use the generated types for
// request/response bodies, then map to UI types before returning from hooks.

// ── (a) Direct re-exports from generated spec ─────────────────────────────────
// These types are used as-is; their shape matches the backend JSON exactly.

export type {
  MemberRole,
  MemberWrite,
  MemberPreferences,
  AppSettings,
  BrandingConfig,
  BrandingUpdateResult,
  FeeTier,
  MemberFeeTier,
  EmailTemplate,
  EmailLogEntry,
  SystemStats,
  ClubYear,
  HourEntry,
  HourEntryStatus,
  MemberHourAccount,
  ManualBookingRequest,
  ConfirmShiftHoursRequest,
  KioskRegisterRequest,
  ConfirmResult,
  MessageResponse,
  ErrorResponse,
  // Raw backend member shape (camelCase) — used in api/members.ts mapper.
  Member as RawMember,
} from '@/api/generated/types.gen';

// ── (b) UI-adapter types ───────────────────────────────────────────────────────
// Components use these shorter-named types; api/* mappers convert raw→UI.

export type UserRole = import('@/api/generated/types.gen').MemberRole;

export interface Member {
  id: string;
  first: string;   // ← firstName
  last: string;    // ← lastName
  email: string;
  since: string;   // ← joinedAt
  goal: number | null;   // ← individualGoalHours
  active?: boolean;      // ← isActive
  reminderOptOut?: boolean;
}

export type EventStatus =
  | 'draft' | 'published' | 'cancelled' | 'completed'
  | 'entwurf' | 'veröffentlicht' | 'abgeschlossen' | 'abgesagt'; // German prototype variants

export type EventVisibility = import('@/api/generated/types.gen').EventVisibility;

// Signup as stored in shift signups (UI view, state mapped from backend 'state')
export interface Signup {
  memberId: string;
  status: 'angemeldet' | 'bestätigt' | 'reserviert' | 'nichterschienen';
  comment?: string;
  hours?: number;
  guest?: string;
}

// Shift uses short HH:MM strings for display; api/events.ts maps from ISO datetimes
export interface Shift {
  id: string;
  name: string;
  start: string; // HH:MM (mapped from startAt)
  end: string;   // HH:MM (mapped from endAt)
  min: number;   // mapped from minHelpers
  max: number;   // mapped from maxHelpers
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

export interface EventTimeline {
  event: Event;
  days: ShiftDay[];
}

// Computed occupancy (richer than the raw API ShiftOccupancy)
export interface ShiftOccupancy {
  count: number;
  min: number;
  max: number;
  key: 'ok' | 'warn' | 'crit' | 'full';
  label: string;
  free: number;
  needsMore: boolean;
}

export interface AuditEntry {
  ts: string;
  who: string;
  what: string;
  cat: string;
}

export interface AuthUser {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  first?: string;
  last?: string;
}

export type RegistrationState = 'registered' | 'reserved' | 'confirmed' | 'no_show';

export interface Registration {
  id: string;
  shiftId: string;
  memberId: string;
  state: RegistrationState;
  comment?: string;
  hours?: number;
  createdAt: string;
}

export interface Tweaks {
  primaryColor: string;
  radius: 'klein' | 'standard' | 'weich';
  warmth: 'cool' | 'neutral' | 'warm';
}
