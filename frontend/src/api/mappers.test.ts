import { describe, it, expect } from 'vitest';
import { flatToEvent, timelineToEvent, type RawTimeline } from './mappers';

describe('flatToEvent', () => {
  it('maps a flat backend event and defaults missing fields safely', () => {
    const ev = flatToEvent({ id: 'e1', name: 'Sommerfest', status: 'published' });
    expect(ev).toMatchObject({
      id: 'e1',
      name: 'Sommerfest',
      status: 'published',
      category: '',
      location: '',
      description: '',
    });
    // days is always an array, never undefined — screens iterate it directly.
    expect(ev.days).toEqual([]);
  });

  it('falls back to draft status when none is provided', () => {
    expect(flatToEvent({ id: 'e', name: 'x' }).status).toBe('draft');
  });
});

describe('timelineToEvent', () => {
  it('flattens days/shifts/registrations into the UI nested shape', () => {
    const tl: RawTimeline = {
      event: { id: 'e1', name: 'Turnier', category: 'Turnier', location: 'Halle', status: 'published' },
      days: [
        {
          date: '2026-06-01T00:00:00Z',
          shifts: [
            {
              shift: {
                id: 's1',
                name: 'Aufbau',
                startAt: '2026-06-01T08:00:00Z',
                endAt: '2026-06-01T12:00:00Z',
                minHelpers: 2,
                maxHelpers: 4,
                requiredQualification: 'Theke',
              },
              registrations: [
                { id: 'r1', memberId: 'm1', state: 'confirmed', comment: 'gern', bookedHours: 4 },
                { id: 'r2', guestEmail: 'gast@example.de', state: 'reserved' },
              ],
            },
          ],
        },
      ],
    };

    const ev = timelineToEvent(tl);
    expect(ev.id).toBe('e1');
    expect(ev.days).toHaveLength(1);

    const day = ev.days[0];
    expect(day.date).toBe('2026-06-01');
    expect(day.shifts).toHaveLength(1);

    const shift = day.shifts[0];
    expect(shift).toMatchObject({ id: 's1', name: 'Aufbau', min: 2, max: 4, qual: 'Theke' });
    // startAt/endAt rendered as HH:MM
    expect(shift.start).toMatch(/^\d{2}:\d{2}$/);
    expect(shift.end).toMatch(/^\d{2}:\d{2}$/);

    expect(shift.signups).toHaveLength(2);
    expect(shift.signups[0]).toMatchObject({ memberId: 'm1', status: 'bestätigt', hours: 4 });
    expect(shift.signups[1]).toMatchObject({ memberId: '', status: 'reserviert', guest: 'gast@example.de' });
  });

  it('tolerates null/missing days, shifts and registrations without throwing', () => {
    const ev = timelineToEvent({ event: { id: 'e', name: 'x' }, days: null });
    expect(ev.days).toEqual([]);

    const ev2 = timelineToEvent({
      event: { id: 'e', name: 'x' },
      days: [{ date: undefined, shifts: null }],
    });
    expect(ev2.days[0].shifts).toEqual([]);
    expect(ev2.days[0].date).toBe('');
  });
});
