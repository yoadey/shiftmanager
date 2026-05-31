import { Icon } from './Icon';

interface StepperProps {
  value: number;
  set: (v: number) => void;
  min?: number;
  max?: number;
  label?: string;
}

export function Stepper({ value, set, min = 0, max = 20, label }: StepperProps) {
  const btn = (delta: number, icon: string) => (
    <button
      className="pressable"
      onClick={() => set(Math.max(min, Math.min(max, value + delta)))}
      style={{
        width: 40,
        height: 40,
        borderRadius: 12,
        border: '1.5px solid var(--line-2)',
        background: 'var(--surface)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: 'pointer',
        color: 'var(--ink)',
      }}
    >
      <Icon name={icon} size={18} stroke={2.4} />
    </button>
  );

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
      {btn(-1, 'minus')}
      <div
        style={{
          minWidth: 30,
          textAlign: 'center',
          fontWeight: 800,
          fontSize: 19,
          fontFamily: 'Bricolage Grotesque',
        }}
      >
        {value}
      </div>
      {btn(1, 'plus')}
      {label && (
        <span style={{ color: 'var(--muted)', fontSize: 13, fontWeight: 600 }}>{label}</span>
      )}
    </div>
  );
}
