interface HourBarProps {
  confirmed: number;
  reserved: number;
  goal: number;
}

/**
 * Two-segment progress bar: confirmed (solid) + reserved (striped) toward goal.
 */
export function HourBar({ confirmed, reserved, goal }: HourBarProps) {
  // A goal of 0 h (exempt member) is met from the start: full bar, no NaN widths.
  const c = goal > 0 ? Math.min(confirmed / goal, 1) * 100 : 100;
  const r = goal > 0 ? Math.min(Math.max(reserved - confirmed, 0) / goal, 1 - c / 100) * 100 : 0;
  return (
    <div className="sm-track">
      <div className="seg-c" style={{ width: c + '%' }} />
      <div className="seg-r" style={{ width: r + '%' }} />
    </div>
  );
}
