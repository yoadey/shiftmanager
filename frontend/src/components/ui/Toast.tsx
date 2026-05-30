import { useAppStore } from '@/store/app.store';

export function Toast() {
  const toast = useAppStore((s) => s.toast);
  if (!toast) return null;

  const dotColor =
    toast.kind === 'crit'
      ? 'var(--crit)'
      : toast.kind === 'warn'
        ? 'var(--warn)'
        : 'var(--ok)';

  return (
    <div className="sm-toast-wrap">
      <div className="sm-toast">
        <span className="tdot" style={{ background: dotColor }} />
        {toast.msg}
      </div>
    </div>
  );
}
