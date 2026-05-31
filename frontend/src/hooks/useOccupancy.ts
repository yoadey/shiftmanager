import type { Shift, ShiftOccupancy } from '@/types';

/**
 * Calculates shift occupancy — pure function, matches prototype occ().
 * Counts 'angemeldet', 'bestätigt', 'reserviert' as occupying a slot.
 */
export function calcOccupancy(shift: Shift): ShiftOccupancy {
  const count = shift.signups.filter(
    (s) => s.status === 'angemeldet' || s.status === 'bestätigt' || s.status === 'reserviert',
  ).length;

  let key: ShiftOccupancy['key'];
  let label: string;

  if (count >= shift.max) {
    key = 'full';
    label = 'Ausgebucht';
  } else if (count >= shift.min) {
    key = 'ok';
    label = 'Besetzt';
  } else if (count > 0) {
    key = 'warn';
    label = 'Teilweise';
  } else {
    key = 'crit';
    label = 'Offen';
  }

  return {
    count,
    min: shift.min,
    max: shift.max,
    key,
    label,
    free: shift.max - count,
    needsMore: count < shift.min,
  };
}

export function useOccupancy() {
  return calcOccupancy;
}
