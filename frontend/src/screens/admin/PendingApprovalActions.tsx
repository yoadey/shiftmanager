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
  // Approving and rejecting the same account at once would leave whichever
  // request lands last in effect; block the other action meanwhile.
  const busy = approve.isPending || reject.isPending;

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
        <Button size="sm" variant="ghost" disabled={reject.isPending} onClick={() => setConfirmReject(false)}>Nein</Button>
        <Button size="sm" variant="danger" disabled={busy} loading={reject.isPending} onClick={onReject}>Ja, ablehnen</Button>
      </div>
    );
  }
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <Button size="sm" variant="ghost" disabled={busy} onClick={() => setConfirmReject(true)}>Ablehnen</Button>
      <Button size="sm" icon="check" disabled={busy} loading={approve.isPending} onClick={onApprove}>
        Freischalten
      </Button>
    </div>
  );
}
