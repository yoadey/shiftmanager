import { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { useAppStore } from '@/store/app.store';
import { useUpdateMember, useDeactivateMember } from '@/api/members';

/**
 * "Freischalten" / "Ablehnen" for a registration pending approval (ML-011).
 * Rejecting asks for confirmation first; it deactivates the account with a
 * leave date so it no longer counts as waiting.
 */
export function PendingApprovalActions({ id, name }: { id: string; name: string }) {
  const { showToast } = useAppStore();
  const approve = useUpdateMember();
  const reject = useDeactivateMember();
  const [confirmReject, setConfirmReject] = useState(false);

  const onApprove = () =>
    approve.mutate(
      { id, active: true },
      {
        onSuccess: () => showToast(`${name} wurde freigeschaltet.`),
        onError: () => showToast('Freischaltung fehlgeschlagen.', 'crit'),
      },
    );

  const onReject = () =>
    reject.mutate(id, {
      onSuccess: () => { setConfirmReject(false); showToast(`Registrierung von ${name} abgelehnt.`); },
      onError: () => showToast('Ablehnen fehlgeschlagen.', 'crit'),
    });

  if (confirmReject) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--ink-2)' }}>Ablehnen?</span>
        <Button size="sm" variant="ghost" onClick={() => setConfirmReject(false)}>Nein</Button>
        <Button size="sm" variant="danger" loading={reject.isPending} onClick={onReject}>Ja, ablehnen</Button>
      </div>
    );
  }
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <Button size="sm" variant="ghost" onClick={() => setConfirmReject(true)}>Ablehnen</Button>
      <Button size="sm" icon="check" loading={approve.isPending} onClick={onApprove}>
        Freischalten
      </Button>
    </div>
  );
}
