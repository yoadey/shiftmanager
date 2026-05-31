import type { ShiftOccupancy } from '@/types';

const colors: Record<string, string> = {
  ok: 'var(--ok)',
  warn: 'var(--warn)',
  crit: 'var(--crit)',
  full: 'var(--ink-2)',
};

interface OccFillProps {
  o: ShiftOccupancy;
}

export function OccFill({ o }: OccFillProps) {
  const width = Math.max(o.count / o.max * 100, o.count ? 8 : 0);
  return (
    <div className="occ-track">
      <div
        className="occ-fill"
        style={{ width: width + '%', background: colors[o.key] ?? 'var(--primary)' }}
      />
    </div>
  );
}
