interface HourBarProps {
  confirmed: number;
  reserved: number;
  goal: number;
}

/**
 * Two-segment progress bar: confirmed (solid) + reserved (striped) toward goal.
 */
export function HourBar({ confirmed, reserved, goal }: HourBarProps) {
  const c = Math.min(confirmed / goal, 1) * 100;
  const r = Math.min(Math.max(reserved - confirmed, 0) / goal, 1 - c / 100) * 100;
  return (
    <div className="sm-track">
      <div className="seg-c" style={{ width: c + '%' }} />
      <div className="seg-r" style={{ width: r + '%' }} />
    </div>
  );
}
