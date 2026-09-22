// ── Backend ↔ UI shape mapping for events / shifts / registrations ──────────
//
// The backend exposes events as a flat list (GET /events) plus a nested
// "timeline" per event (GET /events/{id}/timeline) shaped as:
//
//   { event, days: [ { date, shifts: [ { shift, registrations, occupancy } ] } ] }
//
// The UI works with a single nested `Event` model whose shifts use short HH:MM
// strings and whose registrations are flattened into `signups`. These mappers
// bridge the two and ALWAYS return arrays, so screens can iterate days/shifts/
// signups without defensive guards and never crash on missing data.

import type { Event, EventStatus, Shift, ShiftDay, Signup } from '@/types';

// ── Raw backend shapes (only the fields we consume) ──────────────────────────

export interface RawEvent {
  id: string;
  name: string;
  description?: string;
  location?: string;
  category?: string;
  startDate?: string;
  endDate?: string;
  status?: string;
  visibility?: string;
}

interface RawShift {
  id: string;
  eventId?: string;
  name: string;
  startAt?: string;
  endAt?: string;
  minHelpers?: number;
  maxHelpers?: number;
  requiredQualification?: string;
  date?: string;
}

interface RawRegistration {
  id: string;
  shiftId?: string;
  memberId?: string | null;
  guestEmail?: string | null;
  state?: string;
  comment?: string;
  bookedHours?: number | null;
}

interface RawShiftWithDetails {
  shift: RawShift;
  registrations?: RawRegistration[] | null;
}

interface RawTimelineDay {
  date?: string;
  shifts?: RawShiftWithDetails[] | null;
}

export interface RawTimeline {
  event: RawEvent;
  days?: RawTimelineDay[] | null;
}

// ── Field-level helpers ──────────────────────────────────────────────────────

/** ISO datetime → "HH:MM" (local). Returns '' for missing/invalid input. */
function hhmm(iso?: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
}

/** ISO datetime → "YYYY-MM-DD". Returns '' for missing/invalid input. */
function isoDate(iso?: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toISOString().slice(0, 10);
}

const STATE_TO_STATUS: Record<string, Signup['status']> = {
  registered: 'angemeldet',
  reserved: 'reserviert',
  confirmed: 'bestätigt',
  no_show: 'nichterschienen',
};

// ── Mappers (each tolerant of null/undefined; always returns valid shapes) ────

function toSignup(r: RawRegistration): Signup {
  return {
    id: r.id ?? '',
    memberId: r.memberId ?? '',
    status: STATE_TO_STATUS[r.state ?? ''] ?? 'angemeldet',
    comment: r.comment || undefined,
    hours: r.bookedHours ?? undefined,
    guest: r.guestEmail ?? undefined,
  };
}

function toShift(swd: RawShiftWithDetails): Shift {
  const s = swd.shift ?? ({} as RawShift);
  return {
    id: s.id ?? '',
    name: s.name ?? '',
    start: hhmm(s.startAt),
    end: hhmm(s.endAt),
    min: s.minHelpers ?? 0,
    max: s.maxHelpers ?? 0,
    qual: s.requiredQualification || undefined,
    signups: (swd.registrations ?? []).map(toSignup),
  };
}

function toDay(d: RawTimelineDay): ShiftDay {
  return {
    date: isoDate(d.date),
    shifts: (d.shifts ?? []).map(toShift),
  };
}

/** Flat list event → UI event with an empty timeline (days filled separately). */
export function flatToEvent(e: RawEvent): Event {
  return {
    id: e.id,
    name: e.name ?? '',
    category: e.category ?? '',
    location: e.location ?? '',
    status: (e.status as EventStatus) ?? 'draft',
    description: e.description ?? '',
    days: [],
  };
}

/** Full timeline payload → UI event with nested days/shifts/signups. */
export function timelineToEvent(tl: RawTimeline): Event {
  const e = tl.event ?? ({} as RawEvent);
  return {
    ...flatToEvent(e),
    days: (tl.days ?? []).map(toDay),
  };
}
