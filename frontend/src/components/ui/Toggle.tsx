interface ToggleProps {
  on: boolean;
  onClick: () => void;
}

export function Toggle({ on, onClick }: ToggleProps) {
  return (
    <button
      onClick={onClick}
      className="pressable"
      style={{
        width: 50,
        height: 30,
        borderRadius: 999,
        border: 'none',
        cursor: 'pointer',
        padding: 3,
        background: on ? 'var(--primary)' : 'var(--line-2)',
        transition: 'background .18s',
        flexShrink: 0,
      }}
    >
      <div
        style={{
          width: 24,
          height: 24,
          borderRadius: '50%',
          background: '#fff',
          boxShadow: '0 1px 3px rgba(0,0,0,.25)',
          transform: on ? 'translateX(20px)' : 'translateX(0)',
          transition: 'transform .18s',
        }}
      />
    </button>
  );
}
