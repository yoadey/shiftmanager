import { useLocation, useNavigate } from 'react-router-dom';
import { Icon } from '@/components/ui/Icon';
import { usePendingMembers } from '@/api/members';
import { useIsVorstand } from '@/hooks/useIsVorstand';
import { routes } from '@/routes';

/**
 * Notice at the top of every screen for vorstand/admin (the roles that can
 * activate members) while self-registered accounts wait for activation.
 * Hidden on the member list itself, which shows those accounts on top.
 */
export function PendingMembersBanner() {
  const isVorstand = useIsVorstand();
  const location = useLocation();
  const navigate = useNavigate();
  const { data: pending } = usePendingMembers(isVorstand);

  const count = pending?.length ?? 0;
  if (!isVorstand || count === 0 || location.pathname === routes.mitglieder) return null;

  return (
    <div style={{ padding: '12px 18px 0' }}>
      <button
        type="button"
        className="pressable sm-card"
        onClick={() => navigate(routes.mitglieder)}
        role="status"
        style={{
          width: '100%', display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px',
          background: 'var(--warn-bg)', border: '1px solid var(--line)', cursor: 'pointer', textAlign: 'left',
          font: 'inherit', color: 'inherit',
        }}
      >
        <Icon name="clock" size={20} color="var(--warn)" />
        <span style={{ flex: 1, minWidth: 0 }}>
          <span style={{ display: 'block', fontWeight: 800, fontSize: 14.5 }}>
            {count === 1 ? '1 neues Mitglied wartet' : `${count} neue Mitglieder warten`} auf Freischaltung
          </span>
          <span style={{ display: 'block', color: 'var(--ink-2)', fontSize: 12.5, fontWeight: 600 }}>
            Zur Mitgliederliste, um {count === 1 ? 'es' : 'sie'} freizuschalten oder abzulehnen.
          </span>
        </span>
        <Icon name="chevR" size={16} color="var(--ink-2)" stroke={2.4} />
      </button>
    </div>
  );
}
