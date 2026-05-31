import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { HourBar } from './HourBar';

function widths(container: HTMLElement) {
  const c = container.querySelector('.seg-c') as HTMLElement;
  const r = container.querySelector('.seg-r') as HTMLElement;
  return { c: c.style.width, r: r.style.width };
}

describe('HourBar', () => {
  it('confirmed and reserved segments scale to the goal', () => {
    // confirmed 10/40 = 25%; reserved beyond confirmed (20-10)/40 = 25%
    const { container } = render(<HourBar confirmed={10} reserved={20} goal={40} />);
    expect(widths(container)).toEqual({ c: '25%', r: '25%' });
  });

  it('reserved below confirmed contributes no reserved segment', () => {
    const { container } = render(<HourBar confirmed={20} reserved={10} goal={40} />);
    expect(widths(container)).toEqual({ c: '50%', r: '0%' });
  });

  it('confirmed clamps to 100% when it exceeds the goal', () => {
    const { container } = render(<HourBar confirmed={60} reserved={70} goal={40} />);
    const { c, r } = widths(container);
    expect(c).toBe('100%');
    // no room left for the reserved segment
    expect(r).toBe('0%');
  });

  it('confirmed + reserved together never exceed 100%', () => {
    const { container } = render(<HourBar confirmed={20} reserved={100} goal={40} />);
    const { c, r } = widths(container);
    expect(c).toBe('50%');
    // reserved is clamped to the remaining 50%
    expect(r).toBe('50%');
    expect(parseFloat(c) + parseFloat(r)).toBeLessThanOrEqual(100);
  });
});
