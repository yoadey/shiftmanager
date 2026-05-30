import { describe, it, expect } from 'vitest';
import { calcOccupancy } from './useOccupancy';
import type { Shift, Signup } from '@/types';

function su(status: Signup['status'], memberId = 'm'): Signup {
  return { memberId, status };
}

function makeShift(signups: Signup[], min = 2, max = 4): Shift {
  return {
    id: 's1',
    name: 'Theke',
    start: '10:00',
    end: '14:00',
    min,
    max,
    signups,
  };
}

describe('calcOccupancy', () => {
  it('0 signups → offen / crit, needsMore, full free count', () => {
    const o = calcOccupancy(makeShift([], 2, 4));
    expect(o.key).toBe('crit');
    expect(o.label).toBe('Offen');
    expect(o.count).toBe(0);
    expect(o.free).toBe(4);
    expect(o.needsMore).toBe(true);
  });

  it('below min → teilweise / warn, still needsMore', () => {
    const o = calcOccupancy(makeShift([su('angemeldet')], 2, 4));
    expect(o.key).toBe('warn');
    expect(o.label).toBe('Teilweise');
    expect(o.count).toBe(1);
    expect(o.free).toBe(3);
    expect(o.needsMore).toBe(true);
  });

  it('>=min and <max → besetzt / ok, not needsMore', () => {
    const o = calcOccupancy(
      makeShift([su('angemeldet', 'a'), su('bestätigt', 'b')], 2, 4),
    );
    expect(o.key).toBe('ok');
    expect(o.label).toBe('Besetzt');
    expect(o.count).toBe(2);
    expect(o.free).toBe(2);
    expect(o.needsMore).toBe(false);
  });

  it('>=max → ausgebucht / full, free 0', () => {
    const o = calcOccupancy(
      makeShift(
        [su('angemeldet', 'a'), su('bestätigt', 'b'), su('reserviert', 'c'), su('angemeldet', 'd')],
        2,
        4,
      ),
    );
    expect(o.key).toBe('full');
    expect(o.label).toBe('Ausgebucht');
    expect(o.count).toBe(4);
    expect(o.free).toBe(0);
    expect(o.needsMore).toBe(false);
  });

  it('counts angemeldet, bestätigt and reserviert; ignores nichterschienen', () => {
    const o = calcOccupancy(
      makeShift(
        [su('angemeldet', 'a'), su('bestätigt', 'b'), su('reserviert', 'c'), su('nichterschienen', 'd')],
        2,
        5,
      ),
    );
    expect(o.count).toBe(3);
    expect(o.key).toBe('ok');
    expect(o.free).toBe(2);
  });
});
